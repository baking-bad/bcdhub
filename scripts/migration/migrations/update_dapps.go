package migrations

import (
	"fmt"
	"io/ioutil"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
)

// UpdateDapps - migration that updates dapps and tokens from d_apps.yml and tokens.yml files
type UpdateDapps struct{}

// Key -
func (m *UpdateDapps) Key() string {
	return "update_dapps"
}

// Description -
func (m *UpdateDapps) Description() string {
	return "update dapps and tokens from d_apps.yml and tokens.yml files. store them in scripts/migration/data folder."
}

// DappsData -
type DappsData struct {
	Dapps []database.DApp `yaml:"dapps"`
}

// TokensData -
type TokensData struct {
	Tokens []struct {
		DappSlug string         `yaml:"dapp_slug"`
		Token    database.Token `yaml:"token"`
	} `yaml:"tokens"`
}

// Do - migrate function
func (m *UpdateDapps) Do(ctx *config.Context) error {
	src, err := ioutil.ReadFile("data/d_apps.yml")
	if err != nil {
		return err
	}

	var data DappsData
	if err := yaml.Unmarshal(src, &data); err != nil {
		return err
	}

	tokensSrc, err := ioutil.ReadFile("data/tokens.yml")
	if err != nil {
		return err
	}

	var tokenData TokensData
	if err := yaml.Unmarshal(tokensSrc, &tokenData); err != nil {
		return err
	}

	tokens := make(map[string]database.Token)
	for _, token := range tokenData.Tokens {
		tokens[token.DappSlug] = token.Token
	}

	if err := ctx.DB.DeleteDapps(); err != nil {
		return err
	}

	if err := ctx.DB.DeleteTokens(); err != nil {
		return err
	}

	for i := range data.Dapps {
		if err := ctx.DB.CreateDapp(&(data.Dapps[i])); err != nil {
			return err
		}
	}

	for slug, token := range tokens {
		dapp, err := ctx.DB.GetDAppBySlug(slug)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				logger.Info("no dapp with slug %s", slug)
				continue
			}
			return err
		}

		token.DAppID = dapp.ID

		if err := ctx.DB.CreateToken(&token); err != nil {
			fmt.Println("CreateToken err", err)
			return err
		}
	}

	return nil
}
