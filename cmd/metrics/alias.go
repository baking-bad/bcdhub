package main

import "github.com/aopoltorzhicky/bcdhub/internal/models"

func setAlias(c *models.Contract) error {
	alias, err := ctx.DB.GetAlias(c.Address, c.Network)
	if err != nil {
		return err
	}

	if alias != "" {
		c.Alias = alias
	}
	return nil
}

func setOperationAliases(op *models.Operation) error {
	aliasSource, err := ctx.DB.GetAlias(op.Source, op.Network)
	if err != nil {
		return err
	}

	if aliasSource != "" {
		op.SourceAlias = aliasSource
	}

	aliasDest, err := ctx.DB.GetAlias(op.DestinationAlias, op.Network)
	if err != nil {
		return err
	}

	if aliasDest != "" {
		op.DestinationAlias = aliasDest
	}
	return nil
}
