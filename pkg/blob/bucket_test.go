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

package blob_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/blob"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func testBucket(t *testing.T, config blob.Config) {
	a := assertions.New(t)
	ctx := test.Context()

	bucketName := os.Getenv("TEST_BUCKET")
	if bucketName == "" {
		bucketName = "bucket"
	}

	bucket, err := config.GetBucket(ctx, bucketName)
	a.So(err, should.BeNil)

	now := time.Now().Format(time.RFC3339)

	contents := []byte(now)
	err = bucket.WriteAll(ctx, "path/to/file", contents, blob.WriterOptions("text/plain", "key", "value"))
	a.So(err, should.BeNil)

	res, err := bucket.ReadAll(ctx, "path/to/file")
	a.So(err, should.BeNil)
	a.So(res, should.Resemble, contents)
}

func TestLocal(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("BlobTestLocal_%d", time.Now().UnixNano()/1000000))
	err := os.Mkdir(tmpDir, 0755)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	config := blob.Config{Provider: "local"}
	config.Local.Directory = tmpDir

	testBucket(t, config)
}

func TestAWS(t *testing.T) {
	config := blob.Config{Provider: "aws"}
	config.AWS.Endpoint = os.Getenv("AWS_ENDPOINT")
	config.AWS.Region = os.Getenv("AWS_REGION")
	config.AWS.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	config.AWS.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	if config.AWS.Region == "" || config.AWS.AccessKeyID == "" || config.AWS.SecretAccessKey == "" {
		t.Skip("Missing AWS credentials")
	}

	testBucket(t, config)
}

func TestGCP(t *testing.T) {
	config := blob.Config{Provider: "gcp"}
	config.GCP.CredentialsFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	config.GCP.Credentials = os.Getenv("GCP_CREDENTIALS")

	if config.GCP.CredentialsFile == "" && config.GCP.Credentials == "" {
		_, err := os.Stat("testdata/gcloud.json")
		if err != nil {
			t.Skip("Missing GCP credentials")
		}
		config.GCP.CredentialsFile = "testdata/gcloud.json"
	}

	testBucket(t, config)
}
