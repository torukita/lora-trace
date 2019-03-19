package config

var C Config

type Config struct {
	Name          string `mapstructure:"name"`
	NetworkServer string `mapstructure:"network-server"`
	Server        string `mapstructure:"mqtt-server"`
	Username      string `mapstructure:"mqtt-username"`
	Password      string `mapstructure:"mqtt-password"`
	TraceTopic    string `mapstructure:"mqtt-trace-topic"`
}
