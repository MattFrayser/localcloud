package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type Instance struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Image   string    `json:"image"`
	Status  string    `json:"status"`
	Ports   string    `json:"ports"`
	Created time.Time `json:"created"`
	Uptime  string    `json:"uptime"`
}

type Metrics struct {
	ID          string  `json:"id"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsage uint64  `json:"memory_usage"`
	MemoryLimit uint64  `json:"memory_limit"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkTx   uint64  `json:"network_tx"`
	Timestamp   time.Time `json:"timestamp"`
}

type Manager struct {
	client *client.Client
}

func NewManager() (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Test connection
	_, err = cli.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	return &Manager{client: cli}, nil
}

func (m *Manager) List() []Instance {
	containers, err := m.client.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return []Instance{}
	}

	instances := make([]Instance, 0, len(containers))
	for _, c := range containers {
		instance := m.containerToInstance(c)
		instances = append(instances, instance)
	}

	return instances
}

func (m *Manager) Create(image, name, portMapping string) (*Instance, error) {
	ctx := context.Background()

	// Generate name if not provided
	if name == "" {
		name = fmt.Sprintf("localcloud-%s", uuid.New().String()[:8])
	}

	// Parse port mapping
	hostConfig := &container.HostConfig{}
	if portMapping != "" {
		portBindings, exposedPorts, err := parsePortMapping(portMapping)
		if err != nil {
			return nil, fmt.Errorf("invalid port mapping: %w", err)
		}
		hostConfig.PortBindings = portBindings

		config := &container.Config{
			Image:        image,
			ExposedPorts: exposedPorts,
		}

		resp, err := m.client.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
		if err != nil {
			return nil, fmt.Errorf("failed to create container: %w", err)
		}

		if err := m.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return nil, fmt.Errorf("failed to start container: %w", err)
		}

		// Get updated container info
		containerJSON, err := m.client.ContainerInspect(ctx, resp.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect container: %w", err)
		}

		return m.inspectToInstance(containerJSON), nil
	}

	// Simple container without port mapping
	config := &container.Config{Image: image}
	resp, err := m.client.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	if err := m.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	containerJSON, err := m.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	return m.inspectToInstance(containerJSON), nil
}

func (m *Manager) Delete(containerID string) error {
	ctx := context.Background()

	// Stop container if running
	if err := m.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		// Continue even if stop fails (container might already be stopped)
	}

	// Remove container
	return m.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
}

func (m *Manager) Exec(containerID, command string) (string, error) {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		Cmd:          []string{"sh", "-c", command},
		AttachStdout: true,
		AttachStderr: true,
	}

	execID, err := m.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := m.client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("failed to attach exec: %w", err)
	}
	defer resp.Close()

	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read exec output: %w", err)
	}

	return string(output), nil
}

func (m *Manager) GetLogs(containerID string, tail int) (string, error) {
	ctx := context.Background()

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
	}

	reader, err := m.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(logs), nil
}

func (m *Manager) GetMetrics(containerID string) (*Metrics, error) {
	ctx := context.Background()

	stats, err := m.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer stats.Body.Close()

	var containerStats types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&containerStats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return &Metrics{
		ID:          containerID,
		CPUPercent:  calculateCPUPercent(&containerStats),
		MemoryUsage: containerStats.MemoryStats.Usage,
		MemoryLimit: containerStats.MemoryStats.Limit,
		NetworkRx:   getNetworkRx(containerStats.Networks),
		NetworkTx:   getNetworkTx(containerStats.Networks),
		Timestamp:   time.Now(),
	}, nil
}

func (m *Manager) containerToInstance(c types.Container) Instance {
	name := "unknown"
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}

	ports := ""
	for _, port := range c.Ports {
		if port.PublicPort > 0 {
			ports += fmt.Sprintf("%d:%d ", port.PublicPort, port.PrivatePort)
		}
	}

	uptime := ""
	if c.State == "running" {
		created := time.Unix(c.Created, 0)
		uptime = time.Since(created).Truncate(time.Second).String()
	}

	return Instance{
		ID:      c.ID,
		Name:    name,
		Image:   c.Image,
		Status:  c.Status,
		Ports:   strings.TrimSpace(ports),
		Created: time.Unix(c.Created, 0),
		Uptime:  uptime,
	}
}

func (m *Manager) inspectToInstance(c types.ContainerJSON) *Instance {
	name := strings.TrimPrefix(c.Name, "/")
	
	ports := ""
	if c.NetworkSettings != nil {
		for containerPort, bindings := range c.NetworkSettings.Ports {
			for _, binding := range bindings {
				ports += fmt.Sprintf("%s:%s ", binding.HostPort, containerPort.Port())
			}
		}
	}

	created, _ := time.Parse(time.RFC3339Nano, c.Created)
	
	uptime := ""
	if c.State.Running {
		uptime = time.Since(created).Truncate(time.Second).String()
	}

	return &Instance{
		ID:      c.ID,
		Name:    name,
		Image:   c.Config.Image,
		Status:  c.State.Status,
		Ports:   strings.TrimSpace(ports),
		Created: created,
		Uptime:  uptime,
	}
}

func parsePortMapping(mapping string) (nat.PortMap, nat.PortSet, error) {
	parts := strings.Split(mapping, ":")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("port mapping must be in format host:container")
	}

	containerPort := nat.Port(parts[1] + "/tcp")
	portBindings := nat.PortMap{
		containerPort: []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: parts[0]},
		},
	}
	exposedPorts := nat.PortSet{
		containerPort: struct{}{},
	}

	return portBindings, exposedPorts, nil
}

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	
	if systemDelta > 0 && cpuDelta > 0 {
		return (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return 0.0
}

func getNetworkRx(networks map[string]types.NetworkStats) uint64 {
	var total uint64
	for _, network := range networks {
		total += network.RxBytes
	}
	return total
}

func getNetworkTx(networks map[string]types.NetworkStats) uint64 {
	var total uint64
	for _, network := range networks {
		total += network.TxBytes
	}
	return total
}
