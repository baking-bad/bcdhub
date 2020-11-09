package tzip

import (
	"github.com/tidwall/gjson"
)

// DAppsTZIP -
type DAppsTZIP struct {
	DApps []DApp `json:"dapps,omitempty"`
}

// DApp -
type DApp struct {
	Name              string   `json:"name"`
	ShortDescription  string   `json:"short_description"`
	FullDescription   string   `json:"full_description"`
	WebSite           string   `json:"web_site"`
	Slug              string   `json:"slug,omitempty"`
	AgoraReviewPostID int64    `json:"agora_review_post_id,omitempty"`
	AgoraQAPostID     int64    `json:"agora_qa_post_id,omitempty"`
	Authors           []string `json:"authors"`
	SocialLinks       []string `json:"social_links"`
	Interfaces        []string `json:"interfaces"`
	Categories        []string `json:"categories"`
	Contracts         []string `json:"contracts"`
	Order             int64    `json:"order"`
	Soon              bool     `json:"soon"`

	Pictures  []Picture  `json:"pictures,omitempty"`
	DexTokens []DexToken `json:"dex_tokens,omitempty"`
}

// Picture -
type Picture struct {
	Link string `json:"link"`
	Type string `json:"type"`
}

// ParseElasticJSON -
func (t *Picture) ParseElasticJSON(resp gjson.Result) {
	t.Link = resp.Get("link").String()
	t.Type = resp.Get("type").String()
}

// DexToken -
type DexToken struct {
	TokenID  int64  `json:"token_id"`
	Contract string `json:"contract"`
}

// ParseElasticJSON -
func (t *DexToken) ParseElasticJSON(resp gjson.Result) {
	t.Contract = resp.Get("contract").String()
	t.TokenID = resp.Get("token_id").Int()
}
