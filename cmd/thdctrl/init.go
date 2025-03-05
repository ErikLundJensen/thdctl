package thdctrl

import (
	"fmt"
	"os"
	"strconv"

	"github.com/eriklundjensen/thdctrl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctrl/pkg/robot"
	"github.com/spf13/cobra"
)

const defaultTalosVersion = "v1.9.2"

type cmdFlags struct {
	skipReboot         bool
	enableRescueSystem bool
	disk               string
	version            string
	image              string
}

var initCmdFlags cmdFlags

var initCmd = &cobra.Command{
	Use:   "init <serverNumber>",
	Short: "Initialize the application",
	Args:  cobra.RangeArgs(1, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverNumber, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Error parsing server number: %v\n", err)
			return err
		}
		sshClient := &hetznerapi.SSHClient{}
		err = initializeServer(RobotClient, sshClient, serverNumber, initCmdFlags)
		return err
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initCmdFlags.skipReboot, "skipReboot", "n", false, "skip reboot of server after enabling rescue system.")
	initCmd.Flags().BoolVarP(&initCmdFlags.enableRescueSystem, "enable-rescue-system", "r", false, "entering rescue system even if rescue system already enabled. This will generate a new password.")
	initCmd.Flags().StringVarP(&initCmdFlags.disk, "disk", "d", "nvme0n1", "disk to use for installation of image.")
	initCmd.Flags().StringVarP(&initCmdFlags.version, "version", "v", defaultTalosVersion, "Talos version.")
	initCmd.Flags().StringVarP(&initCmdFlags.image, "image", "i", "", "Talos image URL. Don't use hcloud-amd64 image target Hetzner Cloud, use Talos 'metal' image instead.")
	addCommand(initCmd)
}

func initializeServer(client robot.ClientInterface, sshClient hetznerapi.SSHClientInterface, serverNumber int, f cmdFlags) error {
	sshPassword := os.Getenv("HETZNER_SSH_PASSWORD") // Set your Hetzner password in environment variable

	rescue, err := hetznerapi.GetRescueSystemDetails(client, serverNumber)
	if err != nil {
		fmt.Printf("Error getting rescue system status: %v\n", err)
		return err
	}

	if !rescue.Rescue.Active || f.enableRescueSystem {
		rescue, err = hetznerapi.EnableRescueSystem(client, serverNumber)
		if err != nil {
			fmt.Printf("Error enabling rescue system: %v\n", err)
			return err
		}
	}

	if !f.skipReboot {
		err = hetznerapi.RebootServer(client, serverNumber)
	}
	if err != nil || rescue == nil {
		fmt.Printf("Rescue system state is not available: %v\n", err)
		return err
	}
	sshClient.SetTargetHost(rescue.Rescue.ServerIP, "22")
	
	sshUser := "root"
	if rescue.Rescue.Password != "" {
		sshPassword = rescue.Rescue.Password
	}
	sshClient.Auth(sshUser, sshPassword)

	sshClient.WaitForReboot()
	fmt.Printf("Server rebooted in rescue system mode\n")

	version := defaultTalosVersion
	if f.version != "" {
		if f.image != "" {
			fmt.Println("Warning: Both version and image flags are set. Using image flag.")
		}
		version = f.version
	}
	imageUrl := fmt.Sprintf("https://github.com/siderolabs/talos/releases/download/%s/metal-amd64.raw.zst", version)
	if f.image != "" {
		imageUrl = f.image
	}

	output, sshErr := sshClient.DownloadImage(imageUrl)
	if sshErr != nil {
		fmt.Printf("Failed to download image: %v, output %s\n", err, output)
		return sshErr
	}

	output, sshErr = sshClient.InstallImage(f.disk)
	if sshErr != nil {
		fmt.Printf("Failed to install image: %v output %s\n", err, output)
		_, sshErr = sshClient.ListDisks()
		return sshErr
	}

	hetznerapi.RebootServer(client, serverNumber)
	return nil
}
