package thdctl

import (
	"strconv"

	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var getServerCmd = &cobra.Command{
	Use:   "getServer",
	Short: "Get server details",
	Args:  cobra.RangeArgs(1, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverNumber, err := strconv.Atoi(args[0])
		if err != nil {
			logrus.WithError(err).Error("Error parsing server number")
			return err
		}

		err = getServerDetails(RobotClient, serverNumber)
		return err
	},
}

func init() {
	addCommand(getServerCmd)
}

func getServerDetails(client robot.Client, serverNumber int) error {
	serverDetails, err := hetznerapi.GetServerDetails(client, serverNumber)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Err,
			"msg":   err.Message,
		}).Error("Error getting server details")
		return err.Err
	}

	logrus.WithFields(logrus.Fields{
		"ID":         serverDetails.ServerNumber,
		"Name":       serverDetails.ServerName,
		"Product":    serverDetails.Product,
		"Datacenter": serverDetails.Datacenter,
		"IPv4":       serverDetails.ServerIP,
		"IPv6":       serverDetails.ServerIPv6Net,
	}).Info("Server details")

	return nil
}
