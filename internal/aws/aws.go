package aws

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

// Client -
type Client struct {
	Session *session.Session
	Bucket  string
}

// New -
func New(id, secret, region, bucket string) (*Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		Session: sess,
		Bucket:  bucket,
	}, nil
}

// Upload -
func (c *Client) Upload(body io.Reader, filename string) (*s3manager.UploadOutput, error) {
	uploader := s3manager.NewUploader(c.Session)

	return uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(c.Bucket),
		Key:         aws.String(filename),
		Body:        body,
		ContentType: aws.String("application/json"),
	})
}

// Download -
func (c *Client) Download(filename string) (io.Reader, error) {
	downloader := s3manager.NewDownloader(c.Session)

	buf := aws.NewWriteAtBuffer([]byte{})

	if _, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(filename),
	}); err != nil {
		return nil, errors.Errorf("failed to download file, %v", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}
