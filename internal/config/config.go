package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Environments
const (
	EnvironmentVar  = "BCD_ENV"
	EnvironmentDev  = "development"
	EnvironmentProd = "production"
	EnvironmentYou  = "you"
	EnvironmentBox  = "sandbox"
)

// Config -
type Config struct {
	RPC          map[string]RPCConfig  `yaml:"rpc"`
	TzKT         map[string]TzKTConfig `yaml:"tzkt"`
	Elastic      ElasticSearchConfig   `yaml:"elastic"`
	RabbitMQ     RabbitConfig          `yaml:"rabbitmq"`
	DB           DatabaseConfig        `yaml:"db"`
	OAuth        OAuthConfig           `yaml:"oauth"`
	Sentry       SentryConfig          `yaml:"sentry"`
	SharePath    string                `yaml:"share_path"`
	IPFSGateways []string              `yaml:"ipfs"`

	API struct {
		ProjectName   string     `yaml:"project_name"`
		Bind          string     `yaml:"bind"`
		SwaggerHost   string     `yaml:"swagger_host"`
		CorsEnabled   bool       `yaml:"cors_enabled"`
		OAuthEnabled  bool       `yaml:"oauth_enabled"`
		SentryEnabled bool       `yaml:"sentry_enabled"`
		Seed          SeedConfig `yaml:"seed"`
		Networks      []string   `yaml:"networks"`
		MQ            MQConfig   `yaml:"mq"`
	} `yaml:"api"`

	Compiler struct {
		ProjectName   string    `yaml:"project_name"`
		SentryEnabled bool      `yaml:"sentry_enabled"`
		AWS           AWSConfig `yaml:"aws"`
		MQ            MQConfig  `yaml:"mq"`
	} `yaml:"compiler"`

	Indexer struct {
		Networks map[string]struct {
			Boost string `yaml:"boost"`
		} `yaml:"networks"`
		ProjectName   string `yaml:"project_name"`
		SentryEnabled bool   `yaml:"sentry_enabled"`

		SkipDelegatorBlocks bool     `yaml:"skip_delegator_blocks"`
		MQ                  MQConfig `yaml:"mq"`
	} `yaml:"indexer"`

	Metrics struct {
		ProjectName   string   `yaml:"project_name"`
		SentryEnabled bool     `yaml:"sentry_enabled"`
		MQ            MQConfig `yaml:"mq"`
	} `yaml:"metrics"`

	Scripts struct {
		AWS      AWSConfig `yaml:"aws"`
		Networks []string  `yaml:"networks"`
		MQ       MQConfig  `yaml:"mq"`
	} `yaml:"scripts"`
}

// RPCConfig -
type RPCConfig struct {
	URI     string `yaml:"uri"`
	Timeout int    `yaml:"timeout"`
}

// TzKTConfig -
type TzKTConfig struct {
	URI         string `yaml:"uri"`
	ServicesURI string `yaml:"services_uri"`
	BaseURI     string `yaml:"base_uri"`
	Timeout     int    `yaml:"timeout"`
}

// ElasticSearchConfig -
type ElasticSearchConfig struct {
	URI     []string `yaml:"uri"`
	Timeout int      `yaml:"timeout"`
}

// RabbitConfig -
type RabbitConfig struct {
	URI     string `yaml:"uri"`
	Timeout int    `yaml:"timeout"`
}

// DatabaseConfig -
type DatabaseConfig struct {
	ConnString string `yaml:"conn_string"`
	Timeout    int    `yaml:"timeout"`
}

// AWSConfig -
type AWSConfig struct {
	BucketName      string `yaml:"bucket_name"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
}

// OAuthConfig -
type OAuthConfig struct {
	State string `yaml:"state"`
	JWT   struct {
		Secret      string `yaml:"secret"`
		RedirectURL string `yaml:"redirect_url"`
	} `yaml:"jwt"`
	Github struct {
		ClientID    string `yaml:"client_id"`
		Secret      string `yaml:"secret"`
		CallbackURL string `yaml:"callback_url"`
	} `yaml:"github"`
	Gitlab struct {
		ClientID    string `yaml:"client_id"`
		Secret      string `yaml:"secret"`
		CallbackURL string `yaml:"callback_url"`
	} `yaml:"gitlab"`
}

// SeedConfig -
type SeedConfig struct {
	Enabled bool `yaml:"enabled"`
	User    struct {
		Login     string `yaml:"login"`
		Name      string `yaml:"name"`
		AvatarURL string `yaml:"avatar_url"`
		Token     string `yaml:"token"`
	} `yaml:"user"`
	Subscriptions []struct {
		Address   string `yaml:"address"`
		Network   string `yaml:"network"`
		Alias     string `yaml:"alias"`
		WatchMask uint   `yaml:"watch_mask"`
	} `yaml:"subscriptions"`
	Aliases []struct {
		Alias   string `yaml:"alias"`
		Network string `yaml:"network"`
		Address string `yaml:"address"`
	} `yaml:"aliases"`
	Accounts []struct {
		PrivateKey    string `yaml:"private_key"`
		PublicKeyHash string `yaml:"public_key_hash"`
		Network       string `yaml:"network"`
	} `yaml:"accounts"`
}

// SentryConfig -
type SentryConfig struct {
	Environment string `yaml:"environment"`
	URI         string `yaml:"uri"`
	Debug       bool   `yaml:"debug"`
}

// MQConfig -
type MQConfig struct {
	NeedPublisher bool                   `yaml:"publisher"`
	Queues        map[string]QueueParams `yaml:"queues"`
}

// QueueParams -
type QueueParams struct {
	NonDurable  bool `yaml:"non_durable"`
	AutoDeleted bool `yaml:"auto_deleted"`
}

// LoadDefaultConfig -
func LoadDefaultConfig() (Config, error) {
	configurations := map[string]string{
		EnvironmentProd: "production.yml",
		EnvironmentYou:  "you.yml",
		EnvironmentBox:  "sandbox.yml",
		EnvironmentDev:  "../../configs/development.yml",
	}

	env := os.Getenv(EnvironmentVar)

	config, ok := configurations[env]
	if !ok {
		return Config{}, fmt.Errorf("Unknown configuration for %s variable %s", EnvironmentVar, env)
	}

	return LoadConfig(config)
}

// LoadConfig -
func LoadConfig(filename string) (Config, error) {
	var config Config
	if filename == "" {
		return config, fmt.Errorf("you have to provide configuration filename")
	}

	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("reading file %s error: %w", filename, err)
	}

	expanded := os.ExpandEnv(string(src))

	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return config, fmt.Errorf("unmarshaling configuration file %s error: %w", filename, err)
	}

	return config, nil
}
