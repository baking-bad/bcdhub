package main

type config struct {
	Address string `json:"address"`
	Search  struct {
		URI string `json:"uri"`
	} `json:"search"`
	NodeRPC map[string][]string `json:"nodes"`
	Dir     string              `json:"dir"`
	DB      struct {
		URI string `json:"uri"`
	} `json:"db"`
	Sentry struct {
		Project string `json:"project"`
		Debug   bool   `json:"debug"`
	} `json:"sentry"`
}
