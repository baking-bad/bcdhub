package database

import (
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
)

// Alias -
type Alias struct {
	ID      int64  `gorm:"primary_key,AUTO_INCREMENT" json:"-"`
	Alias   string `json:"alias" example:"Contract alias"`
	Network string `json:"network" example:"babylonnet"`
	Address string `json:"address" example:"KT1CeekjGVRc5ASmgWDc658NBExetoKNuiqz"`
	Slug    string `json:"slug" example:"contract_slug"`
}

// AfterUpdate -
func (a *Alias) AfterUpdate(tx *gorm.DB) (err error) {
	return setSlug(a, tx)
}

func (d *db) GetAliases(network string) ([]Alias, error) {
	var aliases []Alias

	if err := d.ORM.Where("network = ?", network).Find(&aliases).Error; err != nil {
		return nil, err
	}

	return aliases, nil
}

func (d *db) GetAlias(address, network string) (Alias, error) {
	var alias Alias
	if err := d.ORM.Where("address = ? AND network = ?", address, network).First(&alias).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return alias, err
		}
		return alias, nil
	}
	return alias, nil
}

func (d *db) GetBySlug(slug string) (Alias, error) {
	var a Alias
	if err := d.ORM.Where("slug = ?", slug).Find(&a).Error; err != nil {
		return a, err
	}
	return a, nil
}

func createSlug(alias string) (string, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}
	slug := strings.ToLower(reg.ReplaceAllString(alias, "_"))
	return slug, nil
}

func setSlug(a *Alias, tx *gorm.DB) error {
	if a.Slug != "" {
		return nil
	}
	slug, err := createSlug(a.Alias)
	if err != nil {
		return err
	}
	tx.Model(a).Update("slug", slug)
	return nil
}
