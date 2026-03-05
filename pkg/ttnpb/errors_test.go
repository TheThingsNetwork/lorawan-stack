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

package ttnpb_test

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/goproto"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestErrorDetailsToProtoXSSSanitization(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	errDef := errors.Define(
		"test_xss_sanitization",
		"XSS Test Error",
		"path", "paths", "count",
	)

	xssPayload := `<script>alert("xss")</script>`

	errWithXSS := errDef.WithAttributes(
		"path", xssPayload,
		"paths", []string{xssPayload, "safe"},
		"count", 42,
	)

	pb := ttnpb.ErrorDetailsToProto(errWithXSS)
	a.So(pb, should.NotBeNil)

	// Recover attributes from the proto to verify sanitization.
	attrs, err := goproto.Map(pb.GetAttributes())
	a.So(err, should.BeNil)

	// String attribute must be escaped.
	a.So(attrs["path"], should.NotContainSubstring, "<script>")
	a.So(attrs["path"], should.ContainSubstring, "&lt;script&gt;")

	// []string attribute values must be escaped.
	if paths, ok := attrs["paths"].([]any); ok && len(paths) >= 2 {
		a.So(paths[0], should.NotContainSubstring, "<script>")
		a.So(paths[0], should.ContainSubstring, "&lt;script&gt;")
		a.So(paths[1], should.Equal, "safe")
	} else {
		t.Fatal("expected paths attribute to be a list with 2 elements")
	}

	// Numeric attribute must pass through unchanged.
	a.So(attrs["count"], should.Equal, float64(42))
}

func TestGRPCConversion(t *testing.T) {
	a := assertions.New(t)

	errDef := errors.Define("test_grpc_conversion_err_def", "gRPC Conversion Error", "foo")
	a.So(errors.FromGRPCStatus(errDef.GRPCStatus()).Definition, should.EqualErrorOrDefinition, errDef)

	errHello := errDef.WithAttributes("foo", "bar", "baz", "qux")
	errHelloExpected := errDef.WithAttributes("foo", "bar")
	a.So(errors.FromGRPCStatus(errHello.GRPCStatus()), should.EqualErrorOrDefinition, errHelloExpected)
}
