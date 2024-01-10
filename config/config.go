package config

import "time"

type Config struct {
	LogLevel      string        `default:"info" usage:"Log level."`
	LogFormat     string        `default:"text" usage:"Log format. It's text or json."`
	ListenAddress string        `default:":2055" usage:"Network address to accept packets."`
	MetricAddress string        `default:":9438" usage:"Network address to expose metrics."`
	MetricPath    string        `default:"/metrics" usage:"Network path to expose metrics."`
	Include       string        `default:"Count$" usage:"Types to include in collect process. It's regex."`
	Exclude       string        `default:"Time" usage:"Types to exclude in collect process. It's regex."`
	SampleExpire  time.Duration `default:"5m" usage:"How long a sample is valid for."`
}
