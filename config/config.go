package config

import "time"

var Config *MapConfig

type PrometheusConfig struct {
	Enabled  bool   `mapstructure:"PROMETHEUS_ENABLED"`
	Endpoint string `mapstructure:"PROMETHEUS_ENDPOINT"`
	Port     string `mapstructure:"PROMETHEUS_PORT"`
}

type MapConfig struct {
	AppHost            string           `mapstructure:"APP_HOST"`
	DbConnectionString string           `mapstructure:"DB_CONNECTION_STRING"`
	JwtSecretKey       string           `mapstructure:"JWT_SECRET_KEY"`
	JwtExpiresIn       time.Duration    `mapstructure:"JWT_EXPIRE_DURATION"`
	YouTubeAPIKey      string           `mapstructure:"YOUTUBE_API_KEY"`
	Prometheus         PrometheusConfig `mapstructure:"PROMETHEUS"`
}
