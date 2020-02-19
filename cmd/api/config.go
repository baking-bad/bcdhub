package main

type config struct {
	Address string `json:"address"`
	Search  struct {
		URI string `json:"uri"`
	} `json:"search"`
	RPCs map[string]string `json:"rpc"`
	Dir  string            `json:"dir"`
}
