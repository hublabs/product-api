package config

import (
	"os"

	configutil "github.com/pangpanglabs/goutils/config"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/pangpanglabs/goutils/jwtutil"
	"github.com/sirupsen/logrus"
)

var config C

func Init(appEnv string, options ...func(*C)) C {
	if err := configutil.Read(appEnv, &config); err != nil {
		logrus.WithError(err).Warn("Fail to load config file")
	}

	if s := os.Getenv("JWT_SECRET"); s != "" {
		config.JwtSecret = s
		jwtutil.SetJwtSecret(s)
	}

	for _, option := range options {
		option(&config)
	}

	return config
}

func Config() C {
	return config
}

type C struct {
	Database struct {
		Driver     string
		Connection string
		Logger     struct {
			Kafka echomiddleware.KafkaConfig
		}
	}
	RedisConn   string
	BehaviorLog struct {
		Kafka echomiddleware.KafkaConfig
	}
	EventBroker struct {
		Kafka echomiddleware.KafkaConfig
	}
	JwtSecret             string
	ServiceName, HttpPort string
	Debug                 bool
}
