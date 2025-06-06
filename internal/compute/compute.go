package compute

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Instance struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Ports   string
	Created time.Time
}

type InstanceManager struct {
	client    *client.Client
	instances map[string]*Instance
	mu        sync.RWMutex
}

func NewInstanceManager() (*InstanceManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	return &InstanceManager{
		client:    cli,
		instances: make(map[string]*Instance),
	}, nil
}

func (m *InstanceManager) ListInstances() []Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Sync with Docker
	containers, err := m.client.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		fmt.Printf("Error listing containers: %v\n", err)
		return nil
	}

	// Update our instance map
	for _, container := range containers {
		ports := ""
		for _, port := range container.Ports {
			ports += fmt.Sprintf("%d->%d/%s ", port.PublicPort, port.PrivatePort, port.Type)
		}

		instance := &Instance{
			ID:      container.ID,
			Name:    container.Names[0][1:], // Remove leading slash
			Image:   container.Image,
			Status:  container.Status,
			Ports:   ports,
			Created: time.Unix(container.Created, 0),
		}
		m.instances[container.ID] = instance
	}

	// Convert map to slice
	instances := make([]Instance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, *instance)
	}

	return instances
}

func (m *InstanceManager) CreateInstance(ctx context.Context, image string, portMapping string, env map[string]string) (*Instance, error) {
	// Parse port mapping
	hostPort, containerPort, err := parsePortMapping(portMapping)
	if err != nil {
		return nil, fmt.Errorf("invalid port mapping: %v", err)
	}

	// Create container config
	config := &container.Config{
		Image: image,
		ExposedPorts: map[nat.Port]struct{}{
			nat.Port(containerPort): {},
		},
	}

	// Create host config
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(containerPort): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			},
		},
	}

	// Create container
	resp, err := m.client.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	// Start container
	if err := m.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Get container info
	container, err := m.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %v", err)
	}

	// Create instance
	createdUnix, _ := strconv.ParseInt(container.Created, 10, 64)
	instance := &Instance{
		ID:      container.ID,
		Name:    container.Name[1:], // Remove leading slash
		Image:   container.Config.Image,
		Status:  container.State.Status,
		Ports:   portMapping,
		Created: time.Unix(createdUnix, 0),
	}

	// Add to instances map
	m.mu.Lock()
	m.instances[container.ID] = instance
	m.mu.Unlock()

	return instance, nil
}

func (m *InstanceManager) GetLogs(ctx context.Context, id string, since time.Time, tail string) (string, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      since.Format(time.RFC3339),
		Tail:       tail,
	}

	reader, err := m.client.ContainerLogs(ctx, id, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %v", err)
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read container logs: %v", err)
	}

	return string(logs), nil
}

func (m *InstanceManager) ExecCommand(id string, command string) (string, error) {
	ctx := context.Background()
	execConfig := types.ExecConfig{
		Cmd:          []string{"sh", "-c", command},
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create exec instance
	execID, err := m.client.ContainerExecCreate(ctx, id, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec instance: %v", err)
	}

	// Start exec instance
	resp, err := m.client.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec instance: %v", err)
	}
	defer resp.Close()

	// Read output
	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read exec output: %v", err)
	}

	return string(output), nil
}

func parsePortMapping(mapping string) (string, string, error) {
	parts := strings.Split(mapping, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid port mapping format")
	}
	return parts[0], parts[1], nil
}
