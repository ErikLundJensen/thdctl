package thdctl

import (
	"strconv"

	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listFirewallRulesCmd = &cobra.Command{
	Use:   "listFirewallRules <serverNumber>",
	Short: "List all firewall rules for a server",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		serverNumber, err := strconv.Atoi(args[0])
		if err != nil {
			logrus.WithError(err).Error("Error parsing server number")
			return
		}

		listFirewallRules(RobotClient, serverNumber)
	},
}

func init() {
	addCommand(listFirewallRulesCmd)
}

func listFirewallRules(client robot.Client, serverNumber int) error {
	firewallRes, err := hetznerapi.GetFirewallRules(client, serverNumber)
	if err != nil {
		logrus.WithError(err).Error("Error getting firewall rules")
		return err
	}
	logrus.Info("Firewall status:")
	logrus.WithFields(logrus.Fields{
		"Server": serverNumber,
		"Status": firewallRes.Status,
	}).Info("Firewall details")

	logrus.Info("Firewall rules:")
	for _, rule := range firewallRes.Rules {
		logrus.WithFields(logrus.Fields{
			"SrcIP":    rule.SrcIP,
			"DstIP":    rule.DstIP,
			"Protocol": rule.Protocol,
			"SrcPort":  rule.SrcPort,
			"DstPort":  rule.DstPort,
			"Action":   rule.Action,
			"TCPFlags": rule.TCPFlags,
		}).Info("Firewall rule")
	}
	return nil
}
