package thdctl

import (
	"fmt"
	"os"

	v1alpha1 "github.com/eriklundjensen/thdctl/pkg/api/server/v1alpha"
	"github.com/eriklundjensen/thdctl/pkg/controller"
	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	yaml "github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	filename string
	state    string

	reconcileCmd = &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile server configuration from file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename == "" {
				return fmt.Errorf("filename is required")
			}
			sshClient := &hetznerapi.SSHClient{}
			return reconcileFromFile(RobotClient, sshClient, filename, state)
		},
	}
)

func init() {
	reconcileCmd.Flags().StringVarP(&filename, "filename", "f", "", "filename containing server configuration (required)")
	reconcileCmd.Flags().StringVarP(&state, "state", "s", "", "initial state of the server (optional)")
	reconcileCmd.MarkFlagRequired("filename")
	addCommand(reconcileCmd)
}

func readServerConfig(filename string) (*v1alpha1.ServerParameters, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var server v1alpha1.ServerParameters
	if err := yaml.Unmarshal(data, &server); err != nil {
		return nil, fmt.Errorf("error parsing yaml: %v", err)
	}

	if server.TalosImage == "" && server.TalosVersion == "" {
		return nil, fmt.Errorf("TalosImage or TalosVersion must be set")
	}

	return &server, nil
}

func reconcileFromFile(client robot.ClientInterface, sshClient *hetznerapi.SSHClient, filename string, initialState string) error {
	server, err := readServerConfig(filename)
	if err != nil {
		return err
	}

	logrus.Infof("Read configuration for server %d", server.ServerNumber)

	sm := controller.NewStateMachine(client, sshClient, server, 5)
	if initialState != "" {
		sm.StateChange(controller.ServerStatus(initialState))
	}
	if err := sm.Run(); err != nil {
		return fmt.Errorf("failed to run state machine: %v", err)
	}

	return nil
}
