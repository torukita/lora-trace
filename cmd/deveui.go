package cmd

import (
	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/torukita/lora-trace/cmd/config"
	"github.com/torukita/lora-trace/trace"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [deveui]",
		Short: "add deveui to trace framce",
		Run: func(cmd *cobra.Command, args []string) {
			level, _ := log.ParseLevel(viper.GetString("log-level"))
			deveui := args[0]
			var devEUI lorawan.EUI64
			if err := devEUI.UnmarshalText([]byte(deveui)); err != nil {
				log.Fatal(errors.Wrap(err, "wrong format"))
			}
			mqttConfig := &trace.MqttConfig{
				Server:   config.C.Server,
				Username: config.C.Username,
				Password: config.C.Password,
				TopicTop: config.C.TopicTop,
				Debug:    level,
			}

			mqttClient := trace.NewMQTTClient(mqttConfig)
			if err := mqttClient.Connect(); err != nil {
				log.Fatal(err)
			}
			if err := mqttClient.TraceOn(devEUI); err != nil {
				log.Fatal(err)
			}
			log.Trace("added deveui to trace")
		},
	}
	return cmd
}

func newDelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [deveui]",
		Short: "remove deveui from trace list",
		Run: func(cmd *cobra.Command, args []string) {
			level, _ := log.ParseLevel(viper.GetString("log-level"))
			deveui := args[0]
			var devEUI lorawan.EUI64
			if err := devEUI.UnmarshalText([]byte(deveui)); err != nil {
				log.Fatal(errors.Wrap(err, "wrong format"))
			}
			mqttConfig := &trace.MqttConfig{
				Server:   config.C.Server,
				Username: config.C.Username,
				Password: config.C.Password,
				TopicTop: config.C.TopicTop,
				Debug:    level,
			}

			mqttClient := trace.NewMQTTClient(mqttConfig)
			if err := mqttClient.Connect(); err != nil {
				log.Fatal(err)
			}
			if err := mqttClient.TraceOff(devEUI); err != nil {
				log.Fatal(err)
			}
			log.Trace("removed deveui to trace")
		},
	}
	return cmd
}
