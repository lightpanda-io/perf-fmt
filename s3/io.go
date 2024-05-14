// Copyright 2023-2024 Lightpanda (Selecy SAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Puller interface {
	Pull(ctx context.Context) (io.ReadCloser, error)
}

type Pusher interface {
	Push(ctx context.Context, r io.Reader) error
}

type FileIO struct {
	Path string
}

func (fio *FileIO) Pull(ctx context.Context) (io.ReadCloser, error) {
	f, err := os.Open(fio.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return io.NopCloser(&bytes.Buffer{}), nil
		}

		return nil, fmt.Errorf("open file: %w", err)
	}

	return f, nil
}

func (fio *FileIO) Push(ctx context.Context, r io.Reader) error {
	f, err := os.Create(fio.Path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

type S3IO struct {
	svc      *s3.S3
	uploader *s3manager.Uploader
	bucket   string
	item     string
	acl      string
}

func NewS3IO(bucket, item string) (*S3IO, error) {
	session, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("aws session: %w", err)
	}

	svc := s3.New(session)

	return &S3IO{
		bucket: bucket,
		item:   item,
		svc:    svc,

		// set ACL to public-read.
		// https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl-overview.html#canned-acl
		acl: "public-read",

		uploader: s3manager.NewUploaderWithClient(svc),
	}, nil
}

func (s3io *S3IO) Pull(ctx context.Context) (io.ReadCloser, error) {
	obj, err := s3io.svc.GetObjectWithContext(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s3io.bucket),
			Key:    aws.String(s3io.item),
		})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return io.NopCloser(&bytes.Buffer{}), nil
			default:
				return nil, fmt.Errorf("awserr: get object: %w", err)
			}
		}
		return nil, fmt.Errorf("get object: %w", err)
	}

	return obj.Body, nil
}

func (s3io *S3IO) Push(ctx context.Context, r io.Reader) error {
	_, err := s3io.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		ACL:    aws.String(s3io.acl),
		Body:   r,
		Bucket: aws.String(s3io.bucket),
		Key:    aws.String(s3io.item),
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}
