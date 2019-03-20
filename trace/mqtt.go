package trace

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brocaar/lorawan"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	trace_topic = "%s/trace"
	frame_topic = "%s/frame/%s"
)

type MqttConfig struct {
	Server     string
	Username   string
	Password   string
	TraceTopic string
	Debug      log.Level
}

type MqttClient struct {
	cl              MQTT.Client // interface
	config          *MqttConfig
	traceOnHandler  func(lorawan.EUI64) error
	traceOffHandler func(lorawan.EUI64) error
}

type traceConfig struct {
	Trace  bool          `json:"trace"`
	DevEUI lorawan.EUI64 `json:"deveui"`
}

var defaultOnOffHandler = func(deveui lorawan.EUI64) error {
	log.WithField("deveui", deveui).Trace("called defaultOnffHandler")
	return nil
}

func NewMQTTClient(config *MqttConfig) *MqttClient {
	log.SetLevel(config.Debug)
	opts := MQTT.NewClientOptions().AddBroker(config.Server).SetCleanSession(true)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetProtocolVersion(4)
	opts.SetConnectTimeout(3 * time.Second)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts.SetTLSConfig(tlsConfig)
	return &MqttClient{
		cl:              MQTT.NewClient(opts),
		config:          config,
		traceOnHandler:  defaultOnOffHandler,
		traceOffHandler: defaultOnOffHandler,
	}
}

func (c *MqttClient) SetTraceOnHandler(handler func(lorawan.EUI64) error) {
	c.traceOnHandler = handler
}

func (c *MqttClient) SetTraceOffHandler(handler func(lorawan.EUI64) error) {
	c.traceOffHandler = handler
}

func (c *MqttClient) Connect() error {
	if !c.cl.IsConnected() {
		if token := c.cl.Connect(); token.Wait() && token.Error() != nil {
			return errors.Wrap(token.Error(), "failed to connect")
		}
	}
	return nil
}

func (c *MqttClient) TraceOn(deveui lorawan.EUI64) error {
	trace := traceConfig{
		Trace:  true,
		DevEUI: deveui,
	}
	b, _ := json.Marshal(trace)

	if token := c.cl.Publish(fmt.Sprintf(trace_topic, c.config.TraceTopic), 0, false, b); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *MqttClient) TraceOff(deveui lorawan.EUI64) error {
	trace := traceConfig{
		Trace:  false,
		DevEUI: deveui,
	}
	b, _ := json.Marshal(trace)

	if token := c.cl.Publish(fmt.Sprintf(trace_topic, c.config.TraceTopic), 0, false, b); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *MqttClient) TraceFrame(deveui lorawan.EUI64, data []byte) error {
	if token := c.cl.Publish(fmt.Sprintf(frame_topic, c.config.TraceTopic, deveui.String()), 0, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *MqttClient) Run() error {
	topic := fmt.Sprintf(trace_topic, c.config.TraceTopic)

	var onReceived = func(mc MQTT.Client, m MQTT.Message) {
		if m.Topic() == topic {
			var conf traceConfig
			if err := json.Unmarshal(m.Payload(), &conf); err != nil {
				log.WithField("topic", topic).Warn(errors.Wrap(err, "invalid data format"))
				return
			}
			if conf.Trace {
				c.traceOnHandler(conf.DevEUI)
			} else {
				c.traceOffHandler(conf.DevEUI)
			}
		}
	}

	if token := c.cl.Subscribe(topic, 0, onReceived); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	log.Trace("mqtt run finish")
	return nil
}

func init() {
	log.SetLevel(log.TraceLevel)
}
