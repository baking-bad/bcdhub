package database

import (
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// Alias model
type Alias struct {
	ID          int64      `gorm:"primary_key,AUTO_INCREMENT" json:"-"`
	Alias       string     `json:"alias"`
	Network     string     `json:"network"`
	Address     string     `json:"address"`
	Slug        string     `json:"slug,omitempty"`
	ReleaseDate *time.Time `json:"release_date"`
	DAppID      uint       `json:"-"`
}

// OperationAlises -
type OperationAlises struct {
	Source      string
	Destination string
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

func (d *db) GetOperationAliases(src, dst, network string) (OperationAlises, error) {
	var ret OperationAlises
	if err := d.ORM.Raw(`
	SELECT
	COALESCE(
		(SELECT a.alias
		FROM aliases a
		WHERE a.address = ? AND a.network = ?
		), '') AS source,
	COALESCE(
		(SELECT a.alias
		FROM aliases a
		WHERE a.address = ? AND a.network = ?
		), '') AS destination;`, src, network, dst, network).Scan(&ret).Error; err != nil {
		return ret, err
	}

	return ret, nil
}

func (d *db) GetAliasesMap(network string) (map[string]string, error) {
	var aliases []Alias

	if err := d.ORM.Where("network = ?", network).Find(&aliases).Error; err != nil {
		return nil, err
	}

	ret := make(map[string]string, len(aliases))
	for _, a := range aliases {
		ret[a.Address] = a.Alias
	}

	return ret, nil
}

func (d *db) CreateAlias(alias, address, network string) error {
	return d.ORM.Create(&Alias{
		Alias:   alias,
		Address: address,
		Network: network,
	}).Error
}

func (d *db) CreateOrUpdateAlias(a *Alias) error {
	return d.ORM.Where("network = ? AND address = ?", a.Network, a.Address).Assign(Alias{Alias: a.Alias}).FirstOrCreate(a).Error
}
