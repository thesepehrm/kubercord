package config

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	Config = &Configuration{
		Discord: &DiscordConfiguration{
			Webhook: "",
		},
		K8s: &K8sConfiguration{
			Namespace: "",
			ConfigDir: "",
			Interval:  15,
		},
	}
)

type Configuration struct {
	Discord *DiscordConfiguration `mapstructure:"discord"`
	K8s     *K8sConfiguration     `mapstructure:"k8s"`
}

type DiscordConfiguration struct {
	Webhook string `mapstructure:"webhook"`
}

type K8sConfiguration struct {
	Namespace string `mapstructure:"namespace"`
	ConfigDir string `mapstructure:"config_dir"`
	Interval  int    `mapstructure:"interval"`
}

func Init(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
	}

	viper.SetEnvPrefix("KC")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.ReadInConfig(); err != nil {
		logrus.WithError(err).Fatal("Failed to read config file")
	}
	var cfg Configuration

	if err := viper.UnmarshalExact(&cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to parse config file")
	}
	Config = &cfg

}
