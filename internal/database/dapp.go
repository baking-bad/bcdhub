package database

import (
	"time"

	"github.com/lib/pq"
)

// DApp model
type DApp struct {
	ID                uint           `gorm:"primary_key" json:"-"`
	CreatedAt         time.Time      `json:"-"`
	UpdatedAt         time.Time      `json:"-"`
	DeletedAt         *time.Time     `json:"-"`
	Name              string         `json:"name"`
	ShortDescription  string         `json:"short_description"`
	FullDescription   string         `json:"full_description"`
	Version           string         `json:"version"`
	License           string         `json:"license"`
	WebSite           string         `json:"website"`
	Slug              string         `json:"slug,omitempty"`
	AgoraReviewPostID uint           `json:"agora_review_post_id,omitempty"`
	AgoraQAPostID     uint           `json:"agora_qa_post_id,omitempty"`
	Authors           pq.StringArray `gorm:"type:varchar(128)[]" json:"authors"`
	SocialLinks       pq.StringArray `gorm:"type:varchar(1024)[]" json:"social_links"`
	Interfaces        pq.StringArray `gorm:"type:varchar(64)[]" json:"interfaces"`
	Categories        pq.StringArray `gorm:"type:varchar(32)[]" json:"categories"`
	Order             uint           `json:"-"`
	Soon              bool           `gorm:"default:false" json:"soon"`

	Pictures  []Picture `json:"pictures,omitempty"`
	Contracts []Alias   `json:"contracts,omitempty"`
	Tokens    []Token   `json:"tokens,omitempty"`
	DexTokens []Token   `gorm:"many2many:dex_tokens;" json:"dex_tokens,omitempty"`
}

// Picture model
type Picture struct {
	ID     uint   `gorm:"primary_key,AUTO_INCREMENT" json:"-"`
	Link   string `json:"link"`
	Type   string `json:"type"`
	DAppID uint   `json:"-"`
}

// GetDApps -
func (d *db) GetDApps() (dapps []DApp, err error) {
	err = d.Preload("Pictures").Preload("Contracts").Preload("Tokens").Preload("DexTokens").Order(`order`).Find(&dapps).Error
	return
}

// GetDApp -
func (d *db) GetDApp(id uint) (dapp DApp, err error) {
	err = d.Preload("Pictures").Preload("Contracts").Preload("Tokens").Preload("DexTokens").Scopes(idScope(id)).First(&dapp).Error
	return
}

// GetDAppBySlug -
func (d *db) GetDAppBySlug(slug string) (dapp DApp, err error) {
	err = d.Preload("Pictures").Preload("Contracts").Preload("Tokens").Preload("DexTokens").Where("slug = ?", slug).First(&dapp).Error
	return
}
