package thdctl

import (
	"fmt"
	"strconv"

	"github.com/eriklundjensen/thdctl/pkg/hetznerapi"
	"github.com/eriklundjensen/thdctl/pkg/robot"
	"github.com/spf13/cobra"
)

var getServerCmd = &cobra.Command{
	Use:   "getServer",
	Short: "Get server details",
	Args:  cobra.RangeArgs(1, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverNumber, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Error parsing server number: %v\n", err)
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
		fmt.Printf("Error getting server details: %v\n", err.Message)
		return err.Err
	}

	fmt.Printf("ID: %d, Name: %s, Product: %s, Datacenter: %s, IPv4: %s, IPv6: %s\n",
		serverDetails.ServerNumber, serverDetails.ServerName, serverDetails.Product, serverDetails.Datacenter, serverDetails.ServerIP, serverDetails.ServerIPv6Net)

	return nil
}
