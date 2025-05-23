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
	initCmd.Flags().BoolVarP(&initCmdFlags.skipReboot, "skipReboot", "n", false, "skip reboot of server to enable rescue system. HETZNER_SSH_PASSWORD must be set in environment variable when skipReet")
	initCmd.Flags().BoolVarP(&initCmdFlags.enableRescueSystem, "enable-rescue-system", "r", false, "entering rescue system even if rescue system already enabled. This will generate a new password.")
	initCmd.Flags().StringVarP(&initCmdFlags.disk, "disk", "d", "sda", "disk to use for installation of image.")
	initCmd.Flags().StringVarP(&initCmdFlags.version, "version", "v", defaultTalosVersion, "Talos version.")
	initCmd.Flags().StringVarP(&initCmdFlags.image, "image", "i", "", "Talos image URL. Don't use hcloud-amd64 image target Hetzner Cloud, use Talos 'metal' image instead.")
	initCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return validation.ValidateDiskName(initCmdFlags.disk)
	}
	addCommand(initCmd)
}

func initializeServer(client robot.ClientInterface, sshClient hetznerapi.SSHClientInterface, serverNumber int, f cmdFlags) error {
	sshPassword := ""
	if f.skipReboot { 
		if (f.enableRescueSystem){
			return fmt.Errorf("can not enable rescue system and skip reboot at the same time")
		}
		sshPassword = os.Getenv("HETZNER_SSH_PASSWORD") // Set your Hetzner password in environment variable instead of initiating rescue system
		if (sshPassword	== "") {
			return fmt.Errorf("can not skip reboot without setting HETZNER_SSH_PASSWORD")
		}
	}

	rescue, err := hetznerapi.GetRescueSystemDetails(client, serverNumber)
	if err != nil {
		if err.StatusCode == 401 {
			logrus.WithFields(logrus.Fields{
				"username": client.(robot.Client).Username,
			}).Warn("Failed to authenticate with Hetzner API. Please check your credentials.")
		}
		logrus.WithError(err).Error("Error getting rescue system status")
		return err
	}

	if (!rescue.Rescue.Active && sshPassword =="") || f.enableRescueSystem {
		rescue, err = hetznerapi.EnableRescueSystem(client, serverNumber)
		if err != nil {
			logrus.WithError(err).Error("Error enabling rescue system")
			return err
		}
	}

	if !f.skipReboot || sshPassword == ""{
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

	output, sshErr := sshClient.VerifyDiskExists(f.disk)
	if sshErr != nil {
		logrus.WithFields(logrus.Fields{
			"error":  sshErr,
			"output": output,
		}).Error("Disk not found")
		listOutput, listDiskErr := sshClient.ListDisks()
		if listDiskErr != nil {
			logrus.WithError(listDiskErr).Error("Failed to list disks")
		} else {
			disks, _ := hetznerapi.ParseLSBLKOutput(listOutput)
			logrus.Info("Available disks")
			hetznerapi.LogAsJSON(disks)
		}
		return sshErr
	}

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

	output, sshErr = sshClient.DownloadImage(imageUrl)
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
		_, _ = sshClient.ListDisks()
		return sshErr
	}

	hetznerapi.RebootServer(client, serverNumber)
	return nil
}
