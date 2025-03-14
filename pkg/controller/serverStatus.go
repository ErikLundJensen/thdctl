package controller

import (
	"fmt"
	"net"
	"os"
	"time"

	v1alpha1 "github.com/eriklundjensen/thdctl/pkg/api/server/v1alpha"
	"github.com/sirupsen/logrus"

	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
)

// ServerStatus represents the current state of a server
type ServerStatus string

const (
	// TalosAPIAvailable indicates the Talos API is responding on port 50000
	TalosAPIAvailable ServerStatus = "TalosAPIAvailable"

	// TalosImageInstalled indicates a Talos image has been successfully installed
	TalosImageInstalled ServerStatus = "TalosImageInstalled"

	// Wait from reboot after rescue mode is enabled
	WaitForReboot ServerStatus = "WaitForReboot"

	// Requires reboot after rescue mode is enabled
	RequiresReboot ServerStatus = "RequiresReboot"

	// RescueModeInitiated indicates rescue mode has been requested but not yet confirmed ready
	RescueModeInitiated ServerStatus = "RescueModeInitiated"

	// Uninitialized indicates the server exists but has not been configured
	Uninitialized ServerStatus = "Uninitialized"

	// ServerNotFound indicates the specified server number does not exist
	ServerNotFound ServerStatus = "ServerNotFound"

	// MissingServerNumber indicates the server configuration is missing the required server number
	MissingServerNumber ServerStatus = "MissingServerNumber"

	// Unknown indicates the server state could not be determined
	Unknown ServerStatus = "Unknown"

	// Robot API not available
	RobotAPIUnavailable ServerStatus = "RobotAPIUnavailable"

	// SSHAvailable indicates the server is accessible via SSH
	SSHAvailable ServerStatus = "SSHAvailable"
)

// String returns the string representation of the ServerStatus
func (s ServerStatus) String() string {
	return string(s)
}

// VerifyTalosAPIPort checks if the Talos API is accessible on the given host
func VerifyTalosAPIPort(host string, timeoutSeconds int) (bool, error) {
	address := fmt.Sprintf("%s:50000", host)
	timeout := time.Duration(timeoutSeconds) * time.Second

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false, fmt.Errorf("failed to connect to Talos API at %s: %v", address, err)
	}
	defer conn.Close()
	return true, nil
}

// DetermineServerStatus checks the current state of a server and returns its status
func DetermineServerStatus(client robot.ClientInterface, sshClient hetznerapi.SSHClientInterface, server *v1alpha1.ServerParameters) ServerStatus {
	if server.ServerNumber == 0 {
		return MissingServerNumber
	}

	rescue, err := hetznerapi.GetRescueSystemDetails(client, server.ServerNumber)
	if err != nil {
		if err.StatusCode == 404 {
			return ServerNotFound
		}
		logrus.WithError(err).Error("Error getting rescue system status")
		return RobotAPIUnavailable
	}

	host := rescue.Rescue.ServerIP
	if rescue.Rescue.Active {
		logrus.Info("Rescue system active")

		return RescueModeInitiated
	}

	// check if SSH is available
	sshPassword := os.Getenv("HETZNER_SSH_PASSWORD") // Optional set Hetzner ssh password in environment variable

	sshClient.SetTargetHost(host, "22")
	sshUser := "root"
	if rescue.Rescue.Password != "" {
		sshPassword = rescue.Rescue.Password
	}
	sshClient.Auth(sshUser, sshPassword)
	if err := sshClient.EstablishSSHSession(); err == nil {
		return SSHAvailable
	}

	talosAPIAvailable, talosError := VerifyTalosAPIPort(host, 5)
	if talosAPIAvailable {
		return TalosAPIAvailable
	}
	if talosError != nil {
		logrus.WithError(talosError).Error("Talos API not available")
	}
	// Or waiting for Talos API to become available
	// We don't have access to the boot log (require KVM console and a human request towards Hetzner)
	return Unknown
}
