package main

import (
	"encoding/json"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/db/contract"
	"github.com/aopoltorzhicky/bcdhub/internal/microservice"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/streadway/amqp"
)

var ms *microservice.Microservice

func handler(data amqp.Delivery) error {
	var contractID int64
	if err := json.Unmarshal(data.Body, &contractID); err != nil {
		return fmt.Errorf("Unmarshal message body error: %s", err)
	}

	var c contract.Contract
	if err := ms.DB.First(&c, contractID).Error; err != nil {
		return fmt.Errorf("Find contract error: %s", err)
	}

	if err := compute(c); err != nil {
		return fmt.Errorf("Compute error message : %s", err)
	}
	if err := data.Ack(false); err != nil {
		return fmt.Errorf("Error acknowledging message : %s", err)
	}
	return nil
}

func main() {
	var err error
	ms, err = microservice.New("config.json", handler)
	if err != nil {
		panic(err)
	}

	// initCompute()

	ms.Start()
}
