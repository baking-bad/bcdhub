package teztnets

// Info -
type Info map[string]NetworkInfo

// NetworkInfo -
type NetworkInfo struct {
	Aliases            []string  `json:"aliases"`
	Category           string    `json:"category"`
	ChainName          string    `json:"chain_name"`
	Description        string    `json:"description"`
	DockerBuild        string    `json:"docker_build"`
	FaucetURL          string    `json:"faucet_url"`
	GitRef             string    `json:"git_ref"`
	HumanName          string    `json:"human_name"`
	Indexers           []Indexer `json:"indexers"`
	LastBakingDaemon   string    `json:"last_baking_daemon"`
	MaskedFromMainPage bool      `json:"masked_from_main_page"`
	NetworkURL         string    `json:"network_url"`
	RPCURL             string    `json:"rpc_url"`
	RPCUrls            []string  `json:"rpc_urls"`
	ActivatedOn        string    `json:"activated_on"`
}

// Indexer -
type Indexer struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
