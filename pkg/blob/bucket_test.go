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
	. "go.thethings.network/lorawan-stack/pkg/blob"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func bucketName() string {
	if name := os.Getenv("TEST_BUCKET"); name != "" {
		return name
	}
	return "bucket"
}

func testBucket(t *testing.T, conf config.BlobConfig) {
	a := assertions.New(t)
	ctx := test.Context()

	bucket, err := conf.Bucket(ctx, bucketName())
	if !a.So(err, should.BeNil) {
		t.Errorf("Failed to create bucket: %v", err)
		return
	}

	now := time.Now().Format(time.RFC3339)

	contents := []byte(now)
	err = bucket.WriteAll(ctx, "path/to/file", contents, WriterOptions("text/plain", "key", "value"))
	if !a.So(err, should.BeNil) {
		t.Errorf("Failed to write contents: %v", err)
		return
	}

	res, err := bucket.ReadAll(ctx, "path/to/file")
	if !a.So(err, should.BeNil) {
		t.Errorf("Failed to read contents: %v", err)
		return
	}
	a.So(res, should.Resemble, contents)
}

func TestLocal(t *testing.T) {
	a := assertions.New(t)

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("BlobTestLocal_%d", time.Now().UnixNano()/1000000))
	if err := os.Mkdir(tmpDir, 0755); !a.So(err, should.BeNil) {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	conf := config.BlobConfig{Provider: "local"}
	conf.Local.Directory = tmpDir

	testBucket(t, conf)
}

func TestAWS(t *testing.T) {
	conf := config.BlobConfig{Provider: "aws"}
	conf.AWS.Endpoint = os.Getenv("AWS_ENDPOINT")
	conf.AWS.Region = os.Getenv("AWS_REGION")
	conf.AWS.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	conf.AWS.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	if conf.AWS.Region == "" || conf.AWS.AccessKeyID == "" || conf.AWS.SecretAccessKey == "" {
		t.Skip("Missing AWS credentials")
	}

	testBucket(t, conf)
}

func TestGCP(t *testing.T) {
	conf := config.BlobConfig{Provider: "gcp"}
	conf.GCP.CredentialsFile = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	conf.GCP.Credentials = os.Getenv("GCP_CREDENTIALS")

	if conf.GCP.CredentialsFile == "" && conf.GCP.Credentials == "" {
		_, err := os.Stat("testdata/gcloud.json")
		if err != nil {
			t.Skip("Missing GCP credentials")
		}
		conf.GCP.CredentialsFile = "testdata/gcloud.json"
	}

	testBucket(t, conf)
}
