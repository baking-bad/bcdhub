package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

type snapshotCommand struct{}

var snapshotCmd snapshotCommand

// Execute
func (x *snapshotCommand) Execute(_ []string) error {
	if err := uploadMappings(ctx.Storage, creds); err != nil {
		return err
	}
	if err := listRepositories(ctx.Storage); err != nil {
		return err
	}
	name, err := askQuestion("Please, enter target repository name:")
	if err != nil {
		return err
	}
	snapshotName := fmt.Sprintf("snapshot_%s", strings.ToLower(time.Now().UTC().Format(time.RFC3339)))
	return ctx.Storage.CreateSnapshots(name, snapshotName, models.AllDocuments())
}

type restoreCommand struct{}

var restoreCmd restoreCommand

// Execute
func (x *restoreCommand) Execute(_ []string) error {
	if err := listRepositories(ctx.Storage); err != nil {
		return err
	}
	name, err := askQuestion("Please, enter target repository name:")
	if err != nil {
		return err
	}

	if err := listSnapshots(ctx.Storage, name); err != nil {
		return err
	}
	snapshotName, err := askQuestion("Please, enter target snapshot name:")
	if err != nil {
		return err
	}
	return ctx.Storage.RestoreSnapshots(name, snapshotName, models.AllDocuments())
}

type setPolicyCommand struct{}

var setPolicyCmd setPolicyCommand

// Execute
func (x *setPolicyCommand) Execute(_ []string) error {
	if err := listPolicies(ctx.Storage); err != nil {
		return err
	}
	policyID, err := askQuestion("Please, enter target new or existing policy ID:")
	if err != nil {
		return err
	}
	repository, err := askQuestion("Please, enter target repository name:")
	if err != nil {
		return err
	}
	schedule, err := askQuestion("Please, enter schedule in cron format (https://www.elastic.co/guide/en/elasticsearch/reference/current/trigger-schedule.html#schedule-cron):")
	if err != nil {
		return err
	}
	expiredAfter, err := askQuestion("Please, enter expiration in days:")
	if err != nil {
		return err
	}
	iExpiredAfter, err := strconv.ParseInt(expiredAfter, 10, 64)
	if err != nil {
		return err
	}
	return ctx.Storage.SetSnapshotPolicy(policyID, schedule, policyID, repository, iExpiredAfter)
}

type reloadSecureSettingsCommand struct{}

var reloadSecureSettingsCmd reloadSecureSettingsCommand

// Execute
func (x *reloadSecureSettingsCommand) Execute(_ []string) error {
	return ctx.Storage.ReloadSecureSettings()
}

func listPolicies(storage models.GeneralRepository) error {
	policies, err := storage.GetAllPolicies()
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("Availiable snapshot policies")
	fmt.Println("=======================================")
	for i := range policies {
		fmt.Println(policies[i])
	}
	fmt.Println("")
	return nil
}

func listRepositories(storage models.GeneralRepository) error {
	listRepos, err := storage.ListRepositories()
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("Availiable repositories")
	fmt.Println("=======================================")
	for i := range listRepos {
		fmt.Print(listRepos[i].String())
	}
	fmt.Println("")
	return nil
}

func listSnapshots(storage models.GeneralRepository, repository string) error {
	listSnaps, err := storage.ListSnapshots(repository)
	if err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println(listSnaps)
	fmt.Println("")
	return nil
}

func uploadMappings(storage models.GeneralRepository, creds awsData) error {
	mappings, err := storage.GetMappings(models.AllDocuments())
	if err != nil {
		return err
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(creds.Region),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(sess)

	for key, value := range mappings {
		fileName := fmt.Sprintf("mappings/%s.json", key)
		body := strings.NewReader(value)

		if _, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(creds.BucketName),
			Key:         aws.String(fileName),
			Body:        body,
			ContentType: aws.String("application/json"),
		}); err != nil {
			return errors.Errorf("failed to upload file, %v", err)
		}
	}
	return nil
}

// nolint
func restoreMappings(storage models.GeneralRepository, creds awsData) error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(creds.Region),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		return err
	}
	downloader := s3manager.NewDownloader(sess)

	for _, key := range models.AllDocuments() {
		fileName := fmt.Sprintf("mappings/%s.json", key)
		buf := aws.NewWriteAtBuffer([]byte{})

		if _, err := downloader.Download(buf, &s3.GetObjectInput{
			Bucket: aws.String(creds.BucketName),
			Key:    aws.String(fileName),
		}); err != nil {
			return errors.Errorf("failed to upload file, %v", err)
		}
		data := bytes.NewReader(buf.Bytes())

		if err := storage.CreateMapping(key, data); err != nil {
			return err
		}
	}
	return nil
}
