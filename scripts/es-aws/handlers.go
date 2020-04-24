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
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var mappingNames = []string{
	elastic.DocBigMapDiff, elastic.DocBlocks, elastic.DocContracts, elastic.DocMetadata, elastic.DocMigrations, elastic.DocOperations, elastic.DocProtocol,
}

func createRepository(es *elastic.Elastic, creds awsData) error {
	name, err := askQuestion("Please, enter new repository name:")
	if err != nil {
		return err
	}

	return es.CreateAWSRepository(name, creds.BucketName, creds.Region)
}

func snapshot(es *elastic.Elastic, creds awsData) error {
	if err := uploadMappings(es, creds); err != nil {
		return err
	}
	if err := listRepositories(es); err != nil {
		return err
	}
	name, err := askQuestion("Please, enter target repository name:")
	if err != nil {
		return err
	}
	snapshotName := fmt.Sprintf("snapshot_%s", strings.ToLower(time.Now().UTC().Format(time.RFC3339)))
	return es.CreateSnapshots(name, snapshotName, mappingNames)
}

func restore(es *elastic.Elastic, creds awsData) error {
	if err := listRepositories(es); err != nil {
		return err
	}
	name, err := askQuestion("Please, enter target repository name:")
	if err != nil {
		return err
	}

	if err := listSnapshots(es, name); err != nil {
		return err
	}
	snapshotName, err := askQuestion("Please, enter target snapshot name:")
	if err != nil {
		return err
	}
	return es.RestoreSnapshots(name, snapshotName, mappingNames)
}

func setPolicy(es *elastic.Elastic, creds awsData) error {
	if err := listPolicies(es); err != nil {
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
	return es.SetSnapshotPolicy(policyID, schedule, policyID, repository, iExpiredAfter)
}

func listPolicies(es *elastic.Elastic) error {
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

func listRepositories(es *elastic.Elastic) error {
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

func listSnapshots(es *elastic.Elastic, repository string) error {
	listSnaps, err := es.ListSnapshots(repository)
	if err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println(listSnaps)
	fmt.Println("")
	return nil
}

func uploadMappings(es *elastic.Elastic, creds awsData) error {
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
			return fmt.Errorf("failed to upload file, %v", err)
		}
	}
	return nil
}

func restoreMappings(es *elastic.Elastic, creds awsData) error {
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
			return fmt.Errorf("failed to upload file, %v", err)
		}
		data := bytes.NewReader(buf.Bytes())

		if err := es.CreateMapping(key, data); err != nil {
			return err
		}
	}
	return nil
}

func reloadSecureSettings(es *elastic.Elastic, creds awsData) error {
	resp, err := es.API.Nodes.ReloadSecureSettings()
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf(resp.Status())
	}

	return nil
}

func deleteIndices(es *elastic.Elastic, creds awsData) error {
	options := []func(*esapi.IndicesDeleteRequest){
		es.API.Indices.Delete.WithAllowNoIndices(true),
	}

	resp, err := es.API.Indices.Delete(mappingNames, options...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf(resp.Status())
	}

	return nil
}
