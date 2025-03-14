package main

import (
	"os"

	"github.com/eriklundjensen/thdctl/cmd/thdctl"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug     bool
	logFormat string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "thdctl",
		Short: "Talos Hetzner Dedicate Servers CLI",
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log", "txt", "set log format (txt|json)")

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
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}
