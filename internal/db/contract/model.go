package contract

import (
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/db/account"
	"github.com/aopoltorzhicky/bcdhub/internal/db/relation"
	"github.com/aopoltorzhicky/bcdhub/internal/db/tags"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// Contract - entity for contract
type Contract struct {
	ID         int64          `gorm:"AUTO_INCREMENT;unique_index;column:id" json:"id"`
	Network    string         `gorm:"column:network;index:idx_network_address" json:"network"`
	AddressID  int64          `gorm:"column:address_id;index:idx_network_address" json:"-"`
	Level      int64          `gorm:"level" json:"level"`
	Timestamp  time.Time      `gorm:"timestamp" json:"timestamp"`
	Balance    int64          `gorm:"balance" json:"balance"`
	ManagerID  int64          `gorm:"column:manager_id" json:"-"`
	DelegateID int64          `gorm:"column:delegate_id" json:"-"`
	Kind       string         `gorm:"kind" json:"kind"`
	Script     postgres.Jsonb `gorm:"script" json:"-"`
	HashCode   string         `gorm:"hash_code" json:"-"`
	Language   string         `gorm:"language" json:"language"`
	ProjectID  int64          `gorm:"project_id" json:"-"`

	Tags       []tags.Tag `gorm:"foreignkey:ContractID" json:"-"`
	StringTags []string   `sql:"-" json:"tags"`

	Relations []relation.Relation `gorm:"foreignkey:Address;association_foreignkey:Address" json:"relations"`

	Address  account.Account `gorm:"foreignkey:AddressID;association_foreignkey:ID" json:"address"`
	Manager  account.Account `gorm:"foreignkey:ManagerID;association_foreignkey:ID" json:"manager"`
	Delegate account.Account `gorm:"foreignkey:DelegateID;association_foreignkey:ID" json:"delegate"`
}

// TableName - set table name
func (c *Contract) TableName() string {
	return "contracts"
}

// AfterFind -
func (c *Contract) AfterFind() (err error) {
	tags := make([]string, 0)
	if len(c.Tags) > 0 {
		for _, t := range c.Tags {
			found := false
			for _, e := range tags {
				if t.Name == e {
					found = true
					break
				}
			}
			if !found {
				tags = append(tags, t.Name)
			}
		}
	}
	c.StringTags = tags
	return
}
