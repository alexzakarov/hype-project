package config

import (
	"flag"
	"fmt"
	"main/pkg/constants"
	"main/pkg/elasticsearch"
	"main/pkg/eventstroredb"
	"main/pkg/logger"
	"main/pkg/probes"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "ES microservice config path")
}

type Config struct {
	ServiceName         string                         `mapstructure:"serviceName"`
	AppEnv              string                         `mapstructure:"appEnv"`
	ApiVersion          string                         `mapstructure:"apiVersion"`
	Logger              *logger.Config                 `mapstructure:"logger"`
	GRPC                GRPC                           `mapstructure:"grpc"`
	Postgres            Postgres                       `mapstructure:"postgres"`
	PostgresCollections PostgresCollections            `mapstructure:"postgresCollections"`
	Probes              probes.Config                  `mapstructure:"probes"`
	EventStoreConfig    eventstroredb.EventStoreConfig `mapstructure:"eventStoreConfig"`
	Subscriptions       Subscriptions                  `mapstructure:"subscriptions"`
	Elastic             elasticsearch.Config           `mapstructure:"elastic"`
	ElasticIndexes      ElasticIndexes                 `mapstructure:"elasticIndexes"`
	Http                Http                           `mapstructure:"http"`
	OtelCollector       OtelCollector                  `mapstructure:"otelCollector"`
}

type GRPC struct {
	Port        string `mapstructure:"port"`
	Development bool   `mapstructure:"development"`
}

type PostgresCollections struct {
	Orders string `mapstructure:"orders" validate:"required"`
}

type Subscriptions struct {
	PoolSize                    int    `mapstructure:"poolSize" validate:"required,gte=0"`
	OrderPrefix                 string `mapstructure:"orderPrefix" validate:"required,gte=0"`
	PostgresProjectionGroupName string `mapstructure:"postgresProjectionGroupName" validate:"required,gte=0"`
	ElasticProjectionGroupName  string `mapstructure:"elasticProjectionGroupName" validate:"required,gte=0"`
}

type ElasticIndexes struct {
	Orders string `mapstructure:"orders" validate:"required"`
}

type Http struct {
	Port                string   `mapstructure:"port" validate:"required"`
	Development         bool     `mapstructure:"development"`
	BasePath            string   `mapstructure:"basePath" validate:"required"`
	OrdersPath          string   `mapstructure:"ordersPath" validate:"required"`
	DebugErrorsResponse bool     `mapstructure:"debugErrorsResponse"`
	IgnoreLogUrls       []string `mapstructure:"ignoreLogUrls"`
	SslCertPath         string   `mapstructure:"sslCertPath" validate:"required"`
	SslKeyPath          string   `mapstructure:"sslKeyPath" validate:"required"`
}

type Postgres struct {
	Host           string `mapstructure:"host" validate:"required"`
	Port           string `mapstructure:"port" validate:"required"`
	User           string `mapstructure:"user" validate:"required"`
	Password       string `mapstructure:"password" validate:"required"`
	DefaultDb      string `mapstructure:"defaultDb" validate:"required"`
	MaxConnections int    `mapstructure:"maxConnections" validate:"gte=0"`
}

type OtelCollector struct {
	Host string `mapstructure:"host" validate:"required"`
	Port string `mapstructure:"port" validate:"required"`
}

func InitConfig(env string) (*Config, error) {
	if env == "" {
		return nil, errors.New("env required")
	}
	if configPath == "" {
		configPathFromEnv := os.Getenv(constants.ConfigPath)
		if configPathFromEnv != "" {
			configPath = configPathFromEnv
		} else {
			getwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrap(err, "os.Getwd")
			}
			configPath = fmt.Sprintf("%s/config/config.%s.yaml", getwd, env)
			fmt.Println(configPath)
		}
	}

	cfg := &Config{}

	viper.SetConfigType(constants.Yaml)
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}

	grpcPort := os.Getenv(constants.GrpcPort)
	if grpcPort != "" {
		cfg.GRPC.Port = grpcPort
	}
	eventStoreConnectionString := os.Getenv(constants.EventStoreConnectionString)
	if eventStoreConnectionString != "" {
		cfg.EventStoreConfig.ConnectionString = eventStoreConnectionString
	}
	elasticUrl := os.Getenv(constants.ElasticUrl)
	if elasticUrl != "" {
		cfg.Elastic.URL = elasticUrl
	}

	return cfg, nil
}
