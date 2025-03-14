package main

import (
	"os"

	"github.com/eriklundjensen/thdctl/cmd/thdctl"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "thdctl",
		Short: "Talos Hetzner Dedicate Servers CLI",
	}

	for _, cmd := range thdctl.Commands {
		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
