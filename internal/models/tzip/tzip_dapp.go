package tzip

import (
	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/tidwall/gjson"
)

// DApps -
type DApps struct {
	DApps []DApp `json:"dapps,omitempty"`
}

// ParseElasticJSON -
func (t *DApps) ParseElasticJSON(resp gjson.Result) {
	dapps := resp.Get("_source.dapps")
	if dapps.Exists() && dapps.IsArray() {
		t.DApps = make([]DApp, 0)
		for _, dapp := range dapps.Array() {
			var app DApp
			app.ParseElasticJSON(dapp)
			t.DApps = append(t.DApps, app)
		}
	}
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

// ParseElasticJSON -
func (t *DApp) ParseElasticJSON(resp gjson.Result) {
	t.Name = resp.Get("name").String()
	t.ShortDescription = resp.Get("short_description").String()
	t.FullDescription = resp.Get("full_description").String()
	t.WebSite = resp.Get("web_site").String()
	t.Slug = resp.Get("slug").String()
	t.AgoraReviewPostID = resp.Get("agora_review_post_id").Int()
	t.AgoraQAPostID = resp.Get("agora_qa_post_id").Int()

	t.Authors = utils.StringArray(resp, "authors")
	t.SocialLinks = utils.StringArray(resp, "social_links")
	t.Interfaces = utils.StringArray(resp, "interfaces")
	t.Categories = utils.StringArray(resp, "categories")
	t.Contracts = utils.StringArray(resp, "contracts")

	t.Order = resp.Get("order").Int()
	t.Soon = resp.Get("soon").Bool()

	tokens := resp.Get("dex_tokens")
	if tokens.Exists() {
		t.DexTokens = make([]DexToken, 0)
		for _, hit := range tokens.Array() {
			var token DexToken
			token.ParseElasticJSON(hit)
			t.DexTokens = append(t.DexTokens, token)
		}
	}

	pictures := resp.Get("pictures")
	if pictures.Exists() {
		t.Pictures = make([]Picture, 0)
		for _, hit := range pictures.Array() {
			var pic Picture
			pic.ParseElasticJSON(hit)
			t.Pictures = append(t.Pictures, pic)
		}
	}
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
