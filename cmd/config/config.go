package config

var C Config

type Config struct {
	Name     string `mapstructure:"name"`
	NetworkServer string `mapstructure:"network-server"`
	Server   string `mapstructure:"mqtt-server"`
	Username string `mapstructure:"mqtt-username"`
	Password string `mapstructure:"mqtt-password"`
	TopicTop string `mapstructure:"topic-top"`
}

