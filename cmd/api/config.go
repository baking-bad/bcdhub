package main

type config struct {
	Address string `json:"address"`
	Search  struct {
		URI string `json:"uri"`
	} `json:"search"`
	RPCs map[string]string `json:"rpc"`
	Dir  string            `json:"dir"`
	DB   struct {
		URI string `json:"uri"`
	} `json:"db"`
	Sentry struct {
		Project string `json:"project"`
		Env     string `json:"env"`
		DSN     string `json:"dsn"`
		Debug   bool   `json:"debug"`
	} `json:"sentry"`
}
