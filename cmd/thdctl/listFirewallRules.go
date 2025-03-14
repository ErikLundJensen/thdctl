package thdctl

import (
	"fmt"
	"strconv"

	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/spf13/cobra"
)

var listFirewallRulesCmd = &cobra.Command{
	Use:   "listFirewallRules <serverNumber>",
	Short: "List all firewall rules for a server",
	Run: func(cmd *cobra.Command, args []string) {
		serverNumber, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Error parsing server number: %v\n", err)
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
		fmt.Printf("Error getting firewall rules: %v\n", err)
		return err
	}
	fmt.Println("Firewall status:")
	fmt.Printf("Server: %d, Status:%s\n", firewallRes.ServerNumber, firewallRes.Status)

	fmt.Println("Firewall rules:")
	for _, rule := range firewallRes.Rules {
		fmt.Printf("SrcIP: %s, DstIP: %s, Protocol: %s, SrcPort: %s, DstPort: %s, Action: %s, TCPFlags: %s\n",
			rule.SrcIP, rule.DstIP, rule.Protocol, rule.SrcPort, rule.DstPort, rule.Action, rule.TCPFlags)
	}
	return nil
}
