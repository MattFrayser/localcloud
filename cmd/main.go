package main

import (
	"fmt"
	"localcloud/internal/api"
	"localcloud/internal/compute"
	"localcloud/internal/config"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "localcloud",
		Short: "LocalCloud",
	}
	
	// Start web interface
	webCmd = &cobra.Command{
		Use:   "web",
		Short: "Start the web interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			cfg := config.New()
			cfg.Port = port

			manager, err := compute.NewManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %w", err)
			}

			server := api.NewServer(manager, cfg)
			return server.Start()
		},
	}

	// List containers
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := compute.NewManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %w", err)
			}

			instances := manager.List()
			if len(instances) == 0 {
				fmt.Println("No containers found")
				return nil
			}

			fmt.Printf("%-12s %-20s %-30s %-15s\n", "ID", "NAME", "IMAGE", "STATUS")
			for _, instance := range instances {
				fmt.Printf("%-12s %-20s %-30s %-15s\n", 
					instance.ID[:12], instance.Name, instance.Image, instance.Status)
			}
			return nil
		},
	}

	// Create New container
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new container",
		RunE: func(cmd *cobra.Command, args []string) error {
			image, _ := cmd.Flags().GetString("image")
			name, _ := cmd.Flags().GetString("name")
			ports, _ := cmd.Flags().GetString("ports")

			manager, err := compute.NewManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %w", err)
			}

			instance, err := manager.Create(image, name, ports)
			if err != nil {
				return fmt.Errorf("failed to create container: %w", err)
			}

			fmt.Printf("Created container: %s (%s)\n", instance.Name, instance.ID[:12])
			return nil
		},
	}

	// Execute command in a container
	execCmd = &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a container",
		RunE: func(cmd *cobra.Command, args []string) error {
			containerID, _ := cmd.Flags().GetString("id")
			command, _ := cmd.Flags().GetString("command")

			if containerID == "" || command == "" {
				return fmt.Errorf("both --id and --command are required")
			}

			manager, err := compute.NewManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %w", err)
			}

			output, err := manager.Exec(containerID, command)
			if err != nil {
				return fmt.Errorf("failed to execute command: %w", err)
			}

			fmt.Print(output)
			return nil
		},
	}

	// Delete a container	
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a container",
		RunE: func(cmd *cobra.Command, args []string) error {
			containerID, _ := cmd.Flags().GetString("id")
			if containerID == "" {
				return fmt.Errorf("--id is required")
			}

			manager, err := compute.NewManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %w", err)
			}

			if err := manager.Delete(containerID); err != nil {
				return fmt.Errorf("failed to delete container: %w", err)
			}

			fmt.Printf("Deleted container: %s\n", containerID[:12])
			return nil
		},
	}
)

func init() {
	// Web command flags
	webCmd.Flags().Int("port", 8080, "Port to run the web interface on")

	// New command flags
	newCmd.Flags().String("image", "nginx:latest", "Container image")
	newCmd.Flags().String("name", "", "Container name (auto-generated if empty)")
	newCmd.Flags().String("ports", "80:80", "Port mapping (host:container)")

	// Exec command flags
	execCmd.Flags().String("id", "", "Container ID")
	execCmd.Flags().String("command", "", "Command to execute")
	execCmd.MarkFlagRequired("id")
	execCmd.MarkFlagRequired("command")

	// Delete command flags
	deleteCmd.Flags().String("id", "", "Container ID")
	deleteCmd.MarkFlagRequired("id")

	// Add commands
	rootCmd.AddCommand(webCmd, listCmd, newCmd, execCmd, deleteCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
