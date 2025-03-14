package thdctl

import (
	"fmt"
	"os"
	"strconv"

	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/eriklundjensen/thdctl/pkg/validation"
	"github.com/sirupsen/logrus"
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
			logrus.WithError(err).Error("Error parsing server number")
			return err
		}
		sshClient := &hetznerapi.SSHClient{}
		err = initializeServer(RobotClient, sshClient, serverNumber, initCmdFlags)
		return err
	},
}

// TODO: validate disk parameter does not include special characters (prevent injection of commands in shell)
func init() {
	initCmd.Flags().BoolVarP(&initCmdFlags.skipReboot, "skipReboot", "n", false, "skip reboot of server after enabling rescue system.")
	initCmd.Flags().BoolVarP(&initCmdFlags.enableRescueSystem, "enable-rescue-system", "r", false, "entering rescue system even if rescue system already enabled. This will generate a new password.")
	initCmd.Flags().StringVarP(&initCmdFlags.disk, "disk", "d", "nvme0n1", "disk to use for installation of image.")
	initCmd.Flags().StringVarP(&initCmdFlags.version, "version", "v", defaultTalosVersion, "Talos version.")
	initCmd.Flags().StringVarP(&initCmdFlags.image, "image", "i", "", "Talos image URL. Don't use hcloud-amd64 image target Hetzner Cloud, use Talos 'metal' image instead.")
	initCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return validation.ValidateDiskName(initCmdFlags.disk)
	}
	addCommand(initCmd)
}

func initializeServer(client robot.ClientInterface, sshClient hetznerapi.SSHClientInterface, serverNumber int, f cmdFlags) error {
	sshPassword := os.Getenv("HETZNER_SSH_PASSWORD") // Set your Hetzner password in environment variable

	rescue, err := hetznerapi.GetRescueSystemDetails(client, serverNumber)
	if err != nil {
		logrus.WithError(err).Error("Error getting rescue system status")
		return err
	}

	if !rescue.Rescue.Active || f.enableRescueSystem {
		rescue, err = hetznerapi.EnableRescueSystem(client, serverNumber)
		if err != nil {
			logrus.WithError(err).Error("Error enabling rescue system")
			return err
		}
	}

	if !f.skipReboot {
		err = hetznerapi.RebootServer(client, serverNumber)
	}
	if err != nil || rescue == nil {
		logrus.WithError(err).Error("Rescue system state is not available")
		return err
	}
	sshClient.SetTargetHost(rescue.Rescue.ServerIP, "22")

	sshUser := "root"
	if rescue.Rescue.Password != "" {
		sshPassword = rescue.Rescue.Password
	}
	sshClient.Auth(sshUser, sshPassword)

	sshClient.WaitForReboot()
	logrus.Info("Server rebooted in rescue system mode")

	version := defaultTalosVersion
	if f.version != "" {
		if f.image != "" {
			logrus.Warn("Warning: Both version and image flags are set. Using image flag.")
		}
		version = f.version
	}
	imageUrl := fmt.Sprintf("https://github.com/siderolabs/talos/releases/download/%s/metal-amd64.raw.zst", version)
	if f.image != "" {
		imageUrl = f.image
	}

	output, sshErr := sshClient.DownloadImage(imageUrl)
	if sshErr != nil {
		logrus.WithFields(logrus.Fields{
			"error":  sshErr,
			"output": output,
		}).Error("Failed to download image")
		return sshErr
	}

	output, sshErr = sshClient.InstallImage(f.disk)
	if sshErr != nil {
		logrus.WithFields(logrus.Fields{
			"error":  sshErr,
			"output": output,
		}).Error("Failed to install image")
		_, sshErr = sshClient.ListDisks()
		return sshErr
	}

	hetznerapi.RebootServer(client, serverNumber)
	return nil
}
