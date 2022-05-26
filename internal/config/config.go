package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
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
	RPC          map[string]RPCConfig     `yaml:"rpc"`
	TzKT         map[string]TzKTConfig    `yaml:"tzkt"`
	Services     map[string]ServiceConfig `yaml:"services"`
	Storage      StorageConfig            `yaml:"storage"`
	Sentry       SentryConfig             `yaml:"sentry"`
	SharePath    string                   `yaml:"share_path"`
	BaseURL      string                   `yaml:"base_url"`
	IPFSGateways []string                 `yaml:"ipfs"`
	Domains      TezosDomainsConfig       `yaml:"domains"`

	API APIConfig `yaml:"api"`

	Indexer struct {
		Networks        map[string]IndexerConfig `yaml:"networks"`
		ProjectName     string                   `yaml:"project_name"`
		SentryEnabled   bool                     `yaml:"sentry_enabled"`
		Connections     Connections              `yaml:"connections"`
		OffchainBaseURL string                   `yaml:"offchain_base_url"`
	} `yaml:"indexer"`

	Metrics struct {
		ProjectName         string      `yaml:"project_name"`
		SentryEnabled       bool        `yaml:"sentry_enabled"`
		CacheAliasesSeconds int         `yaml:"cache_aliases_seconds"`
		Connections         Connections `yaml:"connections"`
		Networks            []string    `yaml:"networks"`
	} `yaml:"metrics"`

	Scripts struct {
		AWS         AWSConfig   `yaml:"aws"`
		Networks    []string    `yaml:"networks"`
		Connections Connections `yaml:"connections"`
	} `yaml:"scripts"`
}

// IndexerConfig -
type IndexerConfig struct {
	ReceiverThreads int64 `yaml:"receiver_threads"`
}

// RPCConfig -
type RPCConfig struct {
	URI               string `yaml:"uri"`
	Timeout           int    `yaml:"timeout"`
	Cache             string `yaml:"cache"`
	RequestsPerSecond int    `yaml:"requests_per_second"`
}

// TzKTConfig -
type TzKTConfig struct {
	URI     string `yaml:"uri"`
	BaseURI string `yaml:"base_uri"`
	Timeout int    `yaml:"timeout"`
}

// ServiceConfig -
type ServiceConfig struct {
	MempoolURI string `yaml:"mempool"`
}

// StorageConfig -
type StorageConfig struct {
	Postgres PostgresConfig `yaml:"pg"`
	Elastic  []string       `yaml:"elastic"`
	Timeout  int            `yaml:"timeout"`
}

// PostgresConfig -
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
	Password string `yaml:"password"`
	SslMode  string `yaml:"sslmode"`
}

// ConnectionString -
func (p PostgresConfig) ConnectionString() string {
	database := p.DBName
	if database == "" {
		database = "postgres"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s dbname=%s", p.Host, p.Port, p.User, p.Password, p.SslMode, database)
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
	GaEnabled      bool              `yaml:"ga_enabled"`
	MempoolEnabled bool              `yaml:"mempool_enabled"`
	SandboxMode    bool              `yaml:"sandbox_mode"`
	RPC            map[string]string `yaml:"rpc"`
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
	SentryEnabled bool           `yaml:"sentry_enabled"`
	SeedEnabled   bool           `yaml:"seed_enabled"`
	Frontend      FrontendConfig `yaml:"frontend"`
	Seed          SeedConfig     `yaml:"seed"`
	Networks      []string       `yaml:"networks"`
	PageSize      uint64         `yaml:"page_size"`
	Connections   Connections    `yaml:"connections"`
}

// SentryConfig -
type SentryConfig struct {
	Environment string `yaml:"environment"`
	URI         string `yaml:"uri"`
	FrontURI    string `yaml:"front_uri"`
	Debug       bool   `yaml:"debug"`
}

// TezosDomainsConfig -
type TezosDomainsConfig map[string]string

// Connections -
type Connections struct {
	Open int `yaml:"open"`
	Idle int `yaml:"idle"`
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
		return Config{}, fmt.Errorf("unknown configuration for %s variable %s", EnvironmentVar, env)
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

var defaultEnv = regexp.MustCompile(`\${(?P<name>[\w\.]{1,}):-(?P<value>[\w\.:/-]*)}`)

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
