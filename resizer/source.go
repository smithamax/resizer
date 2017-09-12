package resizer

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
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
	bucket string
	prefix string
	client *s3.S3
}

func NewS3Source(bucket, region, prefix string) (*S3Source, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		return nil, err
	}
	return &S3Source{
		bucket,
		prefix,
		s3.New(auth, aws.Regions[region]),
	}, nil
}

func (s *S3Source) Get(p string) (io.ReadCloser, error) {
	path := filepath.Join(s.prefix, p)
	r, err := s.client.Bucket(s.bucket).GetReader(path)

	if aerr, ok := err.(*s3.Error); ok {
		switch aerr.StatusCode {
		case http.StatusNotFound:
			return nil, nil
		}
	}
	return r, err
}

func (s *S3Source) Put(p string, r io.Reader, contentType string) error {
	path := filepath.Join(s.prefix, p)
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return s.client.Bucket(s.bucket).Put(path, data, contentType, s3.Private)
}
