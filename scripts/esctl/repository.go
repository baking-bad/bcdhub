package main

import "github.com/baking-bad/bcdhub/internal/models"

type createRepoCommand struct{}

var createRepoCmd createRepoCommand

// Execute
func (x *createRepoCommand) Execute(_ []string) error {
	name, err := askQuestion("Please, enter new repository name:")
	if err != nil {
		return err
	}

	opts := []models.CreateRepositoryOption{
		models.WithCompress(),
		models.WithMaxRetries(3),
	}

	readOnlyAnswer, err := askQuestion("Read-only (yes/no):")
	if err != nil {
		return err
	}
	if readOnlyAnswer == "yes" {
		opts = append(opts, models.WithReadOnly())
	}

	return ctx.Storage.CreateAWSRepository(name, creds.BucketName, creds.Region, opts...)
}
