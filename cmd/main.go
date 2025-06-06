package main

import (
	"context"
	"fmt"
	"localcloud/internal/compute"
	"localcloud/internal/web"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "localcloud",
		Short: "LocalCloud - A local cloud computing platform",
	}

	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new container",
		RunE: func(cmd *cobra.Command, args []string) error {
			computeManager, err := compute.NewInstanceManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %v", err)
			}
			container, err := computeManager.CreateInstance(context.Background(), "nginx:latest", "80:80", nil)
			if err != nil {
				return fmt.Errorf("failed to create container: %v", err)
			}
			fmt.Printf("Created container: %s\n", container.ID)
			return nil
		},
	}

	webCmd = &cobra.Command{
		Use:   "web",
		Short: "Start the web interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			if port == 0 {
				port = 8080
			}
			computeManager, err := compute.NewInstanceManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %v", err)
			}
			server, err := web.NewServer(computeManager)
			if err != nil {
				return fmt.Errorf("failed to create web server: %v", err)
			}
			return server.Start(fmt.Sprintf(":%d", port))
		},
	}

	execCmd = &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a container",
		RunE: func(cmd *cobra.Command, args []string) error {
			containerID, _ := cmd.Flags().GetString("id")
			command, _ := cmd.Flags().GetString("c")
			if containerID == "" {
				return fmt.Errorf("container ID is required")
			}
			if command == "" {
				return fmt.Errorf("command is required")
			}
			computeManager, err := compute.NewInstanceManager()
			if err != nil {
				return fmt.Errorf("failed to initialize compute manager: %v", err)
			}
			output, err := computeManager.ExecCommand(containerID, command)
			if err != nil {
				return fmt.Errorf("failed to execute command: %v", err)
			}
			fmt.Print(output)
			return nil
		},
	}
)

func init() {
	webCmd.Flags().Int("port", 8080, "Port to run the web interface on")
	execCmd.Flags().String("id", "", "Container ID")
	execCmd.Flags().String("c", "", "Command to execute")
	execCmd.MarkFlagRequired("id")
	execCmd.MarkFlagRequired("c")
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(execCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
