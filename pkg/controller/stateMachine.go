package controller

import (
	"fmt"
	"os"
	"time"

	v1alpha1 "github.com/eriklundjensen/thdctl/pkg/api/server/v1alpha"
	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
)

// StateMachine represents the state machine for server states
type StateMachine struct {
	client          robot.ClientInterface
	sshClient       hetznerapi.SSHClientInterface
	server          *v1alpha1.ServerParameters
	state           ServerStatus
	retries         int
	maxRetries      int
	lastSSHPassword string
}

// NewStateMachine creates a new StateMachine instance
func NewStateMachine(client robot.ClientInterface, sshClient hetznerapi.SSHClientInterface, server *v1alpha1.ServerParameters, maxRetries int) *StateMachine {
	return &StateMachine{
		client:     client,
		sshClient:  sshClient,
		server:     server,
		state:      Unknown,
		maxRetries: maxRetries,
	}
}

func (sm *StateMachine) StateChange(state ServerStatus) {
	if sm.state == state {
		return
	}
	fmt.Printf("State change from: %s to %s\n", sm.state, state)
	sm.state = state
}

// Run executes the state machine
func (sm *StateMachine) Run() error {
	extendedMaxRetries := sm.maxRetries * 2
	for {
		if (sm.retries >= sm.maxRetries && TalosImageInstalled != sm.state && Unknown != sm.state && WaitForReboot != sm.state) || sm.retries >= extendedMaxRetries {
			return fmt.Errorf("max retries reached for state: %s", sm.state)
		}

		switch sm.state {
		case Unknown:
			sm.state = DetermineServerStatus(sm.client, sm.sshClient, sm.server)
			// It is hard to determine the state of the server while rebooting, give it more time to settle until extended max retries eached
			if sm.state == Unknown && sm.retries == extendedMaxRetries-1 {
				sm.StateChange(Uninitialized)
				sm.retries = 0
			}
		case Uninitialized:
			sm.StateChange(sm.initialize())
		case RescueModeInitiated:
			sm.StateChange(sm.checkRescueMode())
		case RequiresReboot:
			sm.StateChange(sm.reboot())
		case WaitForReboot:
			sm.StateChange(sm.checkSSH())
		case SSHAvailable:
			sm.StateChange(installImage(sm))
		case TalosImageInstalled:
			sm.StateChange(sm.checkTalosAPI())
		case TalosAPIAvailable:
			fmt.Println("Talos API is available")
			return nil
		case ServerNotFound, MissingServerNumber, RobotAPIUnavailable:
			return fmt.Errorf("failed to reach a valid state: %s", sm.state)
		default:
			return fmt.Errorf("unknown state: %s", sm.state)
		}

		sm.retries++
		time.Sleep(5 * time.Second) // Add delay between state transitions
	}
}

func (sm *StateMachine) reboot() ServerStatus {
	hetznerapi.RebootServer(sm.client, sm.server.ServerNumber)
	sm.retries = 0
	return WaitForReboot
}

func (sm *StateMachine) initialize() ServerStatus {
	if sm.server.ServerNumber == 0 {
		return MissingServerNumber
	}
	rescue, err := hetznerapi.EnableRescueSystem(sm.client, sm.server.ServerNumber)
	if err != nil || rescue == nil {
		fmt.Printf("Rescue system state is not available: %v\n", err)
		return Uninitialized
	}
	sm.retries = 0
	sm.lastSSHPassword = rescue.Rescue.Password
	return RequiresReboot
}

func (sm *StateMachine) checkRescueMode() ServerStatus {
	rescue, err := hetznerapi.GetRescueSystemDetails(sm.client, sm.server.ServerNumber)
	if err != nil {
		if err.StatusCode == 404 {
			return ServerNotFound
		}
		fmt.Printf("Error getting rescue system status: %v\n", err)
		return RobotAPIUnavailable
	}

	if rescue.Rescue.Active {
		sm.retries = 0
		return RequiresReboot
	}
	return RescueModeInitiated
}

func (sm *StateMachine) checkSSH() ServerStatus {
	rescue, err := hetznerapi.GetRescueSystemDetails(sm.client, sm.server.ServerNumber)
	if err != nil {
		fmt.Printf("Error getting rescue system status: %v\n", err)
		return RobotAPIUnavailable
	}

	host := rescue.Rescue.ServerIP

	sshPassword := sm.lastSSHPassword
	sm.sshClient.SetTargetHost(host, "22")
	sshUser := "root"
	// Use rescue password if available
	if rescue.Rescue.Password != "" {
		sshPassword = rescue.Rescue.Password
	}
	// Override SSH password with environment variable if set
	sshPasswordFromEnv := os.Getenv("HETZNER_SSH_PASSWORD")
	if sshPasswordFromEnv != "" {
		sshPassword = sshPasswordFromEnv
	}

	sm.sshClient.Auth(sshUser, sshPassword)
	if err := sm.sshClient.EstablishSSHSession(); err == nil {
		sm.retries = 0
		return SSHAvailable
	} else {
		fmt.Printf("SSH not available: %v\n", err)
	}
	// Reboot if SSH is not available after several retries
	if rescue.Rescue.Active && sm.retries >= sm.maxRetries-1 {
		sm.retries = 0
		return RequiresReboot
	}
	return sm.state
}

func (sm *StateMachine) checkTalosAPI() ServerStatus {
	rescue, err := hetznerapi.GetRescueSystemDetails(sm.client, sm.server.ServerNumber)
	if err != nil {
		fmt.Printf("Error getting rescue system status: %v\n", err)
		return RobotAPIUnavailable
	}

	host := rescue.Rescue.ServerIP
	talosAPIAvailable, talosError := VerifyTalosAPIPort(host, 5)
	if talosAPIAvailable {
		sm.retries = 0
		return TalosAPIAvailable
	}
	if talosError != nil {
		fmt.Printf("Talos API not available: %v\n", talosError)
	}
	return sm.state
}

func installImage(sm *StateMachine) ServerStatus {
	version := sm.server.TalosVersion
	image := sm.server.TalosImage

	if version != "" && image != "" {
		fmt.Println("Warning: Both version and image are set. Using image definition.")
		version = ""
	}
	if image == "" {
		image = fmt.Sprintf("https://github.com/siderolabs/talos/releases/download/%s/metal-amd64.raw.zst", version)
	}

	output, sshErr := sm.sshClient.DownloadImage(image)
	if sshErr != nil {
		fmt.Printf("Failed to download image: %v, output %s\n", sshErr, output)
		return SSHAvailable
	}

	output, sshErr = sm.sshClient.InstallImage(sm.server.Disk)
	if sshErr != nil {
		fmt.Printf("Failed to install image: %v output %s\n", sshErr, output)
		output, sshErr = sm.sshClient.ListDisks()
		fmt.Printf("Failed list disks: %v output %s\n", sshErr, output)
		return SSHAvailable
	}

	hetznerapi.RebootServer(sm.client, sm.server.ServerNumber)
	sm.retries = 0
	return TalosImageInstalled
}
