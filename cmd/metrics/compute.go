package main

import (
	"fmt"
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
)

func compute(c contract.Contract) error {
	log.Printf("Found %s contract in %s", c.Address.Address, c.Network)
	rpc, ok := ms.RPCs[c.Network]
	if !ok {
		return fmt.Errorf("Unknown network: %s", c.Network)
	}
	upd, err := computeMetrics(rpc, ms.DB, c)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := ms.DB.Model(&c).Updates(upd).Error; err != nil {
		return err
	}
	return nil
}
