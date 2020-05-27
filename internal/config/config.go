package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

// Config -
type Config struct {
	RPC  map[string]RPCConfig  `yaml:"rpc"`
	TzKT map[string]TzKTConfig `yaml:"tzkt"`

	Elastic  ElasticSearchConfig `yaml:"elastic"`
	RabbitMQ RabbitConfig        `yaml:"rabbitmq"`
	DB       DatabaseConfig      `yaml:"db"`
	OAuth    OAuthConfig         `yaml:"oauth"`
	Seed     SeedConfig          `yaml:"seed"`

	Share struct {
		Path string `yaml:"path"`
	} `yaml:"share"`

	Sentry struct {
		Environment string `yaml:"environment"`
		URI         string `yaml:"uri"`
		Debug       bool   `yaml:"debug"`
	} `yaml:"sentry"`

	API struct {
		Bind        string `yaml:"bind"`
		SwaggerHost string `yaml:"swagger_host"`
		OAuth       struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"oauth"`
		Sentry struct {
			Enabled bool   `yaml:"enabled"`
			Project string `yaml:"project"`
		} `yaml:"sentry"`
		Networks []string `yaml:"networks"`
		Seed     struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"seed"`
	} `yaml:"api"`

	Indexer struct {
		Sentry struct {
			Enabled bool   `yaml:"enabled"`
			Project string `yaml:"project"`
		} `yaml:"sentry"`
		Networks map[string]struct {
			Boost string `yaml:"boost"`
		} `yaml:"networks"`
	} `yaml:"indexer"`

	Metrics struct {
		Sentry struct {
			Enabled bool   `yaml:"enabled"`
			Project string `yaml:"project"`
		} `yaml:"sentry"`
	} `yaml:"metrics"`

	Migrations struct {
		Networks []string `yaml:"networks"`
	} `yaml:"migrations"`

	AWS struct {
		BucketName      string `yaml:"bucket_name"`
		Region          string `yaml:"region"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
	} `yaml:"aws"`
}

// RPCConfig -
type RPCConfig struct {
	URI     string `yaml:"uri"`
	Timeout int    `yaml:"timeout"`
}

// ElasticSearchConfig -
type ElasticSearchConfig struct {
	URI     string `yaml:"uri"`
	Timeout int    `yaml:"timeout"`
}

// DatabaseConfig -
type DatabaseConfig struct {
	ConnString string `yaml:"conn_string"`
}

// RabbitConfig -
type RabbitConfig struct {
	URI    string   `yaml:"uri"`
	Queues []string `yaml:"queues"`
}

// TzKTConfig -
type TzKTConfig struct {
	URI         string `yaml:"uri"`
	ServicesURI string `yaml:"services_uri"`
	Timeout     int    `yaml:"timeout"`
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
	User struct {
		Login     string `yaml:"login"`
		Name      string `yaml:"name"`
		AvatarURL string `yaml:"avatar_url"`
		Token     string `yaml:"token"`
	} `yaml:"user"`
	Subscriptions []struct {
		Address string `yaml:"address"`
		Network string `yaml:"network"`
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

// LoadConfig -
func LoadConfig(filenames ...string) (Config, error) {
	var config Config
	if len(filenames) <= 0 {
		return config, fmt.Errorf("You have to provide at least one filename")
	}

	var sections map[string]interface{}
	for _, filename := range filenames {

		var override map[string]interface{}
		src, err := ioutil.ReadFile(filename)
		if err != nil {
			return config, err
		}
		if err := yaml.Unmarshal(src, &override); err != nil {
			return config, err
		}

		if sections == nil {
			sections = override
		} else {
			for k, v := range override {
				sections[k] = v
			}
		}
	}

	res, err := yaml.Marshal(sections)
	if err != nil {
		return config, err
	}

	// log.Println(string(res))

	res = []byte(os.ExpandEnv(string(res)))
	if err := yaml.Unmarshal(res, &config); err != nil {
		return config, err
	}

	return config, nil
}

// LoadDefaultConfig -
func LoadDefaultConfig() (Config, error) {
	var options struct {
		ConfigFiles []string `short:"f" default:"config.yml" description:"Config filename.yml"`
	}

	_, err := flags.Parse(&options)
	if err != nil {
		return Config{}, err
	}

	return LoadConfig(options.ConfigFiles...)
}
