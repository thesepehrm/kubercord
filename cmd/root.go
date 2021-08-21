package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thesepehrm/kubercord/config"
	"github.com/thesepehrm/kubercord/core"
)

var cfgFile string
var logFormat string
var kubeConfig *string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kubercord",
	Short: "kubercord",
	Long:  `KuberCord Application`,

	Run: func(cmd *cobra.Command, args []string) {
		kc, err := core.Init()
		if err != nil {
			logrus.WithError(err).Panic("Failed to initialize kubercord")
		}
		logrus.Info("Kubercord initialized")
		kc.Start()
		logrus.Info("Kubercord started")

		// Handle SIGINT
		quitChannel := make(chan os.Signal, 1)
		signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
		<-quitChannel
		println()

		logrus.Warn("Stopping kubercord")
		kc.Stop()
		logrus.Info("Exited")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("Failed to execute root command")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		// Here you will define your flags and configuration settings.
		if logFormat == "text" {
			logrus.SetFormatter(&logrus.TextFormatter{})
		} else {
			logrus.SetFormatter(&logrus.JSONFormatter{})
		}

		config.Init(cfgFile)
		logrus.Info("Configurations initialized")
	})

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yml", "config file (default is config.yml)")
	RootCmd.PersistentFlags().StringVar(&logFormat, "logfmt", "text", "log format (text or json)")

}
