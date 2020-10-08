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
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/pkg/errors"
)

var mappingNames = []string{
	elastic.DocBigMapDiff,
	elastic.DocBlocks,
	elastic.DocContracts,
	elastic.DocMetadata,
	elastic.DocMigrations,
	elastic.DocOperations,
	elastic.DocProtocol,
	elastic.DocBigMapActions,
	elastic.DocTransfers,
	elastic.DocTokenMetadata,
	elastic.DocTZIP,
}

type snapshotCommand struct{}

var snapshotCmd snapshotCommand

// Execute
func (x *snapshotCommand) Execute(args []string) error {
	if err := uploadMappings(ctx.ES, creds); err != nil {
		return err
	}
	if err := listRepositories(ctx.ES); err != nil {
		return err
	}
	name, err := askQuestion("Please, enter target repository name:")
	if err != nil {
		return err
	}
	snapshotName := fmt.Sprintf("snapshot_%s", strings.ToLower(time.Now().UTC().Format(time.RFC3339)))
	return ctx.ES.CreateSnapshots(name, snapshotName, mappingNames)
}

type restoreCommand struct{}

var restoreCmd restoreCommand

// Execute
func (x *restoreCommand) Execute(args []string) error {
	if err := listRepositories(ctx.ES); err != nil {
		return err
	}
	name, err := askQuestion("Please, enter target repository name:")
	if err != nil {
		return err
	}

	if err := listSnapshots(ctx.ES, name); err != nil {
		return err
	}
	snapshotName, err := askQuestion("Please, enter target snapshot name:")
	if err != nil {
		return err
	}
	return ctx.ES.RestoreSnapshots(name, snapshotName, mappingNames)
}

type setPolicyCommand struct{}

var setPolicyCmd setPolicyCommand

// Execute
func (x *setPolicyCommand) Execute(args []string) error {
	if err := listPolicies(ctx.ES); err != nil {
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
	return ctx.ES.SetSnapshotPolicy(policyID, schedule, policyID, repository, iExpiredAfter)
}

type reloadSecureSettingsCommand struct{}

var reloadSecureSettingsCmd reloadSecureSettingsCommand

// Execute
func (x *reloadSecureSettingsCommand) Execute(args []string) error {
	return ctx.ES.ReloadSecureSettings()
}

func listPolicies(es elastic.IElastic) error {
	policies, err := es.GetAllPolicies()
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

func listRepositories(es elastic.IElastic) error {
	listRepos, err := es.ListRepositories()
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("Availiable repositories")
	fmt.Println("=======================================")
	for i := range listRepos {
		fmt.Println(listRepos[i])
	}
	fmt.Println("")
	return nil
}

func listSnapshots(es elastic.IElastic, repository string) error {
	listSnaps, err := es.ListSnapshots(repository)
	if err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println(listSnaps)
	fmt.Println("")
	return nil
}

func uploadMappings(es elastic.IElastic, creds awsData) error {
	mappings, err := es.GetMappings(mappingNames)
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
func restoreMappings(es elastic.IElastic, creds awsData) error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(creds.Region),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		return err
	}
	downloader := s3manager.NewDownloader(sess)

	for _, key := range mappingNames {
		fileName := fmt.Sprintf("mappings/%s.json", key)
		buf := aws.NewWriteAtBuffer([]byte{})

		if _, err := downloader.Download(buf, &s3.GetObjectInput{
			Bucket: aws.String(creds.BucketName),
			Key:    aws.String(fileName),
		}); err != nil {
			return errors.Errorf("failed to upload file, %v", err)
		}
		data := bytes.NewReader(buf.Bytes())

		if err := es.CreateMapping(key, data); err != nil {
			return err
		}
	}
	return nil
}
