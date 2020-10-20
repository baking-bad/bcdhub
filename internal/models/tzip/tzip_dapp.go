package tzip

import (
	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/tidwall/gjson"
)

// DApp -
type DApp struct {
	Name              string   `json:"name"`
	ShortDescription  string   `json:"short_description"`
	FullDescription   string   `json:"full_description"`
	Version           string   `json:"version"`
	License           string   `json:"license"`
	WebSite           string   `json:"website"`
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
	t.Name = resp.Get("_source.name").String()
	t.ShortDescription = resp.Get("_source.short_description").String()
	t.FullDescription = resp.Get("_source.full_description").String()
	t.Version = resp.Get("_source.version").String()
	t.License = resp.Get("_source.license").Time().String()
	t.WebSite = resp.Get("_source.website").String()
	t.Slug = resp.Get("_source.slug").String()
	t.AgoraReviewPostID = resp.Get("_source.agora_review_post_id").Int()
	t.AgoraQAPostID = resp.Get("_source.agora_qa_post_id").Int()

	t.Authors = utils.StringArray(resp, "_source.authors")
	t.SocialLinks = utils.StringArray(resp, "_source.social_links")
	t.Interfaces = utils.StringArray(resp, "_source.interfaces")
	t.Categories = utils.StringArray(resp, "_source.categories")
	t.Contracts = utils.StringArray(resp, "_source.contracts")

	t.Order = resp.Get("_source.order").Int()
	t.Soon = resp.Get("_source.soon").Bool()

	tokens := resp.Get("_source.dex_tokens")
	if tokens.Exists() {
		t.DexTokens = make([]DexToken, 0)
		for _, hit := range tokens.Array() {
			var token DexToken
			token.ParseElasticJSON(hit)
			t.DexTokens = append(t.DexTokens, token)
		}
	}

	pictures := resp.Get("_source.pictures")
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
