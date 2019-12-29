package project

import (
	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
)

// Project - entity for project
type Project struct {
	ID    int64  `gorm:"AUTO_INCREMENT;unique_index;column:id" json:"id"`
	Alias string `gorm:"alias" json:"alias,omitempty"`

	Contracts []contract.Contract `gorm:"foreignkey:ProjectID" json:"contracts"`

	Tags []string `sql:"-" json:"tags,omitempty"`
}

// TableName - set table name
func (p *Project) TableName() string {
	return "projects"
}

// AfterFind -
func (p *Project) AfterFind() (err error) {
	tags := make([]string, 0)
	if len(p.Contracts) > 0 {
		for _, c := range p.Contracts {
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
	}
	p.Tags = tags
	return
}
