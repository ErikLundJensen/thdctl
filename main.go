package main

import (
	"os"

	"github.com/eriklundjensen/thdctl/cmd/thdctl"
	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
)
var debug bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "thdctl",
		Short: "Talos Hetzner Dedicate Servers CLI",
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	for _, cmd := range thdctl.Commands {
		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}