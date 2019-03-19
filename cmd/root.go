package cmd

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/torukita/lora-trace/cmd/config"
	"github.com/torukita/lora-trace/trace"
	"os"
	"os/signal"
	"syscall"
)

var (
	version = "v0.0.2"
)

func Execute() error {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lora-trace",
		Short:   "",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			level, _ := log.ParseLevel(viper.GetString("log-level"))
			log.Info("start")
			nsConfig := &trace.Config{
				NetworkServer: config.C.NetworkServer,
				Debug:         level,
			}
			mqttConfig := &trace.MqttConfig{
				Server:   config.C.Server,
				Username: config.C.Username,
				Password: config.C.Password,
				TopicTop: config.C.TopicTop,
				Debug:    level,
			}

			mqttClient := trace.NewMQTTClient(mqttConfig)
			nsClient := trace.NewNSClient(nsConfig)

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			err := trace.NewTrace(nsClient, mqttClient).Start()
			if err != nil {
				log.Fatal(err)
			}

			<-c

			log.Info("finish")
		},
	}

	cobra.OnInitialize(initConfig)
	cmd.PersistentFlags().String("log-level", "info", "panic,fatal,error,warn,info,debug,trace")
	viper.BindPFlag("log-level", cmd.PersistentFlags().Lookup("log-level"))

	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newDelCmd())
	return cmd
}

func initConfig() {
	viper.SetConfigType("toml")
	viper.SetConfigName("lora-trace")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(errors.Wrap(err, "Fatal error toml config file"))
	}
	if err := viper.Unmarshal(&config.C); err != nil {
		log.Fatal(errors.Wrap(err, "Fatal unmarshal toml config file"))
	}
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.TraceLevel)
}
