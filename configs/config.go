package configs

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

const appName = "duty_bot"

type Config struct {
	ApiToken                string `required:"true"`
	Port                    string `default:"80"`
	Address                 string `required:"true"`
	Debug                   bool   `default:"true"`
	Tls                     bool   `default:"false"`
	CertFile                string `required:"true"`
	KeyFile                 string `required:"true"`
	IntervalTime            int    `default:"1"`
	BaseTimeForNotification int    `default:"11"`
	ChatIDForNotification   int64
}

func NewConfig() *Config {

	var c Config
	err := envconfig.Process(appName, &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &c
}
