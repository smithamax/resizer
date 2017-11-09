package resizer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Source interface {
	Get(path string) (io.ReadCloser, error)
	Put(path string, r io.Reader, contentType string) error
}

type FileSource struct {
	Root string
}

func (s FileSource) Get(p string) (io.ReadCloser, error) {
	path := filepath.Join(s.Root, p)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return file, err
}

func (s FileSource) Put(p string, r io.Reader, contentType string) error {
	path := filepath.Join(s.Root, p)
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("could not create file: %s", err)
	}
	w := bufio.NewWriter(f)

	_, err = io.Copy(w, r)

	return err
}

type S3Source struct {
	bucket   string
	prefix   string
	client   *s3.S3
	uploader *s3manager.Uploader
}

func NewS3Source(bucket, region, prefix string) (*S3Source, error) {
	sess, err := session.NewSession(&aws.Config{Region: &region})

	if err != nil {
		return nil, err
	}

	creds := credentials.NewChainCredentials([]credentials.Provider{
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
		&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(sess), ExpiryWindow: 5 * time.Minute},
	})

	sess.Config.Credentials = creds

	return &S3Source{
		bucket,
		prefix,
		s3.New(sess),
		s3manager.NewUploader(sess),
	}, nil
}

func (s *S3Source) Get(p string) (io.ReadCloser, error) {
	path := filepath.Join(s.prefix, p)
	input := &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &path,
	}

	result, err := s.client.GetObject(input)
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchKey:
			return nil, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (s *S3Source) Put(p string, r io.Reader, contentType string) error {
	path := filepath.Join(s.prefix, p)
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      &s.bucket,
		Key:         &path,
		Body:        r,
		ContentType: &contentType,
	})
	return err
}
