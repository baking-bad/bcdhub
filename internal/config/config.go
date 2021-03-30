package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

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
	Storage      StorageConfig         `yaml:"storage"`
	RabbitMQ     RabbitConfig          `yaml:"rabbitmq"`
	DB           DatabaseConfig        `yaml:"db"`
	OAuth        OAuthConfig           `yaml:"oauth"`
	Sentry       SentryConfig          `yaml:"sentry"`
	SharePath    string                `yaml:"share_path"`
	BaseURL      string                `yaml:"base_url"`
	IPFSGateways []string              `yaml:"ipfs"`
	Domains      TezosDomainsConfig    `yaml:"domains"`

	API APIConfig `yaml:"api"`

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
		ProjectName         string   `yaml:"project_name"`
		SentryEnabled       bool     `yaml:"sentry_enabled"`
		CacheAliasesSeconds int      `yaml:"cache_aliases_seconds"`
		MQ                  MQConfig `yaml:"mq"`
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

// StorageConfig -
type StorageConfig struct {
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

// FrontendConfig -
type FrontendConfig struct {
	GaEnabled      bool `yaml:"ga_enabled"`
	MempoolEnabled bool `yaml:"mempool_enabled"`
	SandboxMode    bool `yaml:"sandbox_mode"`
}

// SeedConfig -
type SeedConfig struct {
	User struct {
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

type APIConfig struct {
	ProjectName   string         `yaml:"project_name"`
	Bind          string         `yaml:"bind"`
	SwaggerHost   string         `yaml:"swagger_host"`
	CorsEnabled   bool           `yaml:"cors_enabled"`
	OAuthEnabled  bool           `yaml:"oauth_enabled"`
	SentryEnabled bool           `yaml:"sentry_enabled"`
	SeedEnabled   bool           `yaml:"seed_enabled"`
	Frontend      FrontendConfig `yaml:"frontend"`
	Seed          SeedConfig     `yaml:"seed"`
	Networks      []string       `yaml:"networks"`
	MQ            MQConfig       `yaml:"mq"`
	Pinata        PinataConfig   `yaml:"pinata"`
	PageSize      uint64         `yaml:"page_size"`
}

// SentryConfig -
type SentryConfig struct {
	Environment string `yaml:"environment"`
	URI         string `yaml:"uri"`
	FrontURI    string `yaml:"front_uri"`
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
	TTLSeconds  uint `yaml:"ttl_seconds"`
}

// TezosDomainsConfig -
type TezosDomainsConfig map[string]string

// PinataConfig -
type PinataConfig struct {
	Key            string `yaml:"key"`
	SecretKey      string `yaml:"secret_key"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
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

	expanded := expandEnv(string(src))

	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return config, fmt.Errorf("unmarshaling configuration file %s error: %w", filename, err)
	}

	return config, nil
}

var defaultEnv = regexp.MustCompile(`\${(?P<name>[\w\.]{1,}):-(?P<value>[\w\.]*)}`)

func expandEnv(data string) string {
	vars := defaultEnv.FindAllStringSubmatch(data, -1)
	data = defaultEnv.ReplaceAllString(data, `${$name}`)

	for i := range vars {
		if _, ok := os.LookupEnv(vars[i][1]); !ok {
			os.Setenv(vars[i][1], vars[i][2])
		}
	}

	data = os.ExpandEnv(data)
	return data
}
