package core

import "fmt"

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	DBName   string `yaml:"dbname"`
	Password string `yaml:"password"`
	SslMode  string `yaml:"sslmode"`
}

// ConnectionString -
func (cfg Config) ConnectionString() string {
	database := cfg.DBName
	if database == "" {
		database = "postgres"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s dbname=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.SslMode, database)
}
