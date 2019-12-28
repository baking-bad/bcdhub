package main

type config struct {
	Address string `json:"address"`
	Db      struct {
		URI string `json:"uri"`
		Log bool   `json:"log"`
	} `json:"db"`
}
