// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/s3blob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

var (
	errInvalidConfig = errors.DefineInvalidArgument("invalid_config", "invalid blob store configuration")
)

func Local(_ context.Context, bucket, path string) (*blob.Bucket, error) {
	bucketPath, err := filepath.Abs(filepath.Join(path, bucket))
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(bucketPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(bucketPath, 0755)
		if err != nil {
			return nil, errInvalidConfig.WithCause(err)
		}
	} else if err != nil {
		return nil, err
	}
	return fileblob.OpenBucket(bucketPath, nil)
}

func AWS(ctx context.Context, bucket string, conf *aws.Config) (*blob.Bucket, error) {
	s, err := session.NewSession(conf)
	if err != nil {
		return nil, err
	}
	return s3blob.OpenBucket(ctx, s, bucket, nil)
}

func GCP(ctx context.Context, bucket string, jsonCredentials []byte) (*blob.Bucket, error) {
	creds, err := google.CredentialsFromJSON(ctx, jsonCredentials, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}
	cli, err := gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
	if err != nil {
		return nil, err
	}
	return gcsblob.OpenBucket(ctx, cli, bucket, nil)
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
