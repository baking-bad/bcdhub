package main

import (
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
)

func compute(c contract.Contract) error {
	log.Printf("Found %s contract in %s", c.Address, c.Network)
	upd, err := computeMetrics(ms.DB, c)
	if err != nil {
		return err
	}

	if err := ms.DB.Model(&c).Updates(upd).Error; err != nil {
		return err
	}
	return nil
}
