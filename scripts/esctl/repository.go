package main

type createRepoCommand struct{}

var createRepoCmd createRepoCommand

// Execute
func (x *createRepoCommand) Execute(_ []string) error {
	name, err := askQuestion("Please, enter new repository name:")
	if err != nil {
		return err
	}

	return ctx.Storage.CreateAWSRepository(name, creds.BucketName, creds.Region)
}
