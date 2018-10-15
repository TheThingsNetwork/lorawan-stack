// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package blob

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/go-cloud/blob"
	"github.com/google/go-cloud/blob/fileblob"
	"github.com/google/go-cloud/blob/gcsblob"
	"github.com/google/go-cloud/blob/s3blob"
	"github.com/google/go-cloud/gcp"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"golang.org/x/oauth2/google"
)

// Config of the blob store.
type Config config.Blob

var (
	errUnknownProvider = errors.DefineInternal("unknown_provider", "unknown blob store provider `{provider}`")
	errConfig          = errors.DefineInternal("missing_config", "missing blob store configuration")
)

// GetBucket returns the requested blob bucket using the config.
func (c Config) GetBucket(ctx context.Context, bucket string) (*blob.Bucket, error) {
	switch c.Provider {
	case "local":
		return c.getLocal(ctx, bucket)
	case "aws":
		return c.getAWS(ctx, bucket)
	case "gcp":
		return c.getGCP(ctx, bucket)
	default:
		return nil, errUnknownProvider.WithAttributes("provider", c.Provider)
	}
}

func (c Config) getLocal(_ context.Context, bucket string) (*blob.Bucket, error) {
	bucketPath, err := filepath.Abs(filepath.Join(c.Local.Directory, bucket))
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(bucketPath)
	if err != nil {
		_, err = os.Stat(c.Local.Directory)
		if err != nil {
			return nil, errConfig.WithCause(err)
		}
		err = os.Mkdir(bucketPath, 0755)
		if err != nil {
			return nil, errConfig.WithCause(err)
		}
	}
	return fileblob.NewBucket(bucketPath)
}

type awsCredentials Config

func (c awsCredentials) Retrieve() (value credentials.Value, err error) {
	value.ProviderName = "TTNConfigProvider"
	value.AccessKeyID, value.SecretAccessKey = c.AWS.AccessKeyID, c.AWS.SecretAccessKey
	if value.AccessKeyID == "" || value.SecretAccessKey == "" {
		return value, errConfig
	}
	return value, nil
}

func (c awsCredentials) IsExpired() bool { return false }

func (c Config) getAWS(ctx context.Context, bucket string) (*blob.Bucket, error) {
	s, err := session.NewSession(&aws.Config{
		Endpoint:    &c.AWS.Endpoint,
		Region:      &c.AWS.Region,
		Credentials: credentials.NewCredentials(awsCredentials(c)),
	})
	if err != nil {
		return nil, err
	}
	return s3blob.OpenBucket(ctx, s, bucket)
}

func (c Config) getGCP(ctx context.Context, bucket string) (*blob.Bucket, error) {
	var jsonData []byte
	if c.GCP.Credentials != "" {
		jsonData = []byte(c.GCP.Credentials)
	} else if c.GCP.CredentialsFile != "" {
		var err error
		jsonData, err = ioutil.ReadFile(c.GCP.CredentialsFile)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errConfig
	}
	creds, err := google.CredentialsFromJSON(ctx, jsonData)
	if err != nil {
		return nil, err
	}
	cli, err := gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
	if err != nil {
		return nil, err
	}
	return gcsblob.OpenBucket(ctx, bucket, cli)
}

// WriterOptions returns WriterOptions with the given content type and metadata
// from the given key-value pairs.
func WriterOptions(contentType string, kv ...string) *blob.WriterOptions {
	opts := &blob.WriterOptions{
		ContentType: contentType,
	}
	if len(kv) > 0 {
		if len(kv)%2 != 0 {
			panic("Odd number of key-value elements")
		}
		m := make(map[string]string, len(kv)/2)
		var key string
		for i, node := range kv {
			if i%2 == 0 {
				key = node
			} else {
				m[key] = node
			}
		}
		opts.Metadata = m
	}
	return opts
}
