package main

import (
	"fmt"
	"os"
)

type awsData struct {
	BucketName string
	Region     string
}

// FromEnv -
func (c *awsData) FromEnv() error {
	c.BucketName = os.Getenv("BCD_AWS_BUCKET_NAME")
	if c.BucketName == "" {
		return fmt.Errorf("Please, set BCD_AWS_BUCKET_NAME")
	}

	c.Region = os.Getenv("BCD_AWS_REGION")
	if c.Region == "" {
		return fmt.Errorf("Please, set BCD_AWS_REGION")
	}
	return nil
}
