package thdctl

import (
	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listServersCmd = &cobra.Command{
	Use:   "listServers",
	Short: "List all servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := listServers(RobotClient)
		return err
	},
}

func init() {
	addCommand(listServersCmd)
}

func listServers(client robot.Client) error {
	servers, err := hetznerapi.ListServers(client)
	if err != nil {
		logrus.WithError(err).Error("Error listing servers")
		return err
	}
	logrus.Info("List of servers:")
	for _, server := range servers {
		serverDetails := server.Server
		logrus.WithFields(logrus.Fields{
			"ID":         serverDetails.ServerNumber,
			"Name":       serverDetails.ServerName,
			"Product":    serverDetails.Product,
			"Datacenter": serverDetails.Datacenter,
			"IPv4":       serverDetails.ServerIP,
			"IPv6":       serverDetails.ServerIPv6Net,
		}).Info("Server details")
	}
	return nil
}
