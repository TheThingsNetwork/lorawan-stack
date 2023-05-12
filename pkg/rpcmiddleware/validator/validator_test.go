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

package validator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/validator"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type testSubject struct {
	err            error
	validateCalled bool
	ctx            context.Context
	handlerCalled  bool
}

func (t testSubject) getResult() *testSubject {
	return &t
}

func (t *testSubject) ValidateFields(...string) error { return nil }

type msgWithValidate struct {
	testSubject
}

func (m *msgWithValidate) Validate() error {
	m.validateCalled = true
	return m.err
}

type msgWithValidateContext struct {
	testSubject
}

func (m *msgWithValidateContext) ValidateContext(ctx context.Context) error {
	m.validateCalled = true
	m.ctx = ctx
	return m.err
}

type msgWithFieldMask struct {
	testSubject
	fieldMask *fieldmaskpb.FieldMask

	fieldIsZero map[string]bool
}

func (m *msgWithFieldMask) GetFieldMask() *fieldmaskpb.FieldMask { return m.fieldMask }

func (m *msgWithFieldMask) ValidateFields(...string) error { return nil }

func (m *msgWithFieldMask) FieldIsZero(s string) bool {
	v, ok := m.fieldIsZero[s]
	if !ok {
		panic(fmt.Sprintf("FieldIsZero called with unexpected field: '%s'", s))
	}
	return v
}

func handler(ctx context.Context, req any) (any, error) {
	res := req.(interface{ getResult() *testSubject }).getResult()
	res.handlerCalled = true
	return res, nil
}

func TestUnaryServerInterceptor(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testErr := errors.New("test")

	RegisterAllowedFieldMaskPaths("/ttn.lorawan.v3.Test/Unary", true, []string{
		"foo",
		"foo.a",
		"foo.a.a",
		"foo.a.b",
		"foo.b",
	},
		"foo",
		"foo.a",
		"foo.a.b",
	)

	info := &grpc.UnaryServerInfo{FullMethod: "/ttn.lorawan.v3.Test/Unary"}

	intercept := UnaryServerInterceptor()

	res, err := intercept(ctx, &testSubject{}, info, handler)
	if a.So(err, should.BeNil) {
		a.So(res.(*testSubject).handlerCalled, should.BeTrue)
	}

	res, err = intercept(ctx, &msgWithValidate{}, info, handler)
	if a.So(err, should.BeNil) {
		a.So(res.(*testSubject).validateCalled, should.BeTrue)
		a.So(res.(*testSubject).handlerCalled, should.BeTrue)
	}

	_, err = intercept(ctx, &msgWithValidate{testSubject{
		err: testErr,
	}}, info, handler)
	if a.So(err, should.BeError) {
		a.So(err, should.EqualErrorOrDefinition, testErr)
	}

	res, err = intercept(ctx, &msgWithValidateContext{}, info, handler)
	if a.So(err, should.BeNil) {
		a.So(res.(*testSubject).validateCalled, should.BeTrue)
		a.So(res.(*testSubject).handlerCalled, should.BeTrue)
		a.So(res.(*testSubject).ctx, should.Equal, ctx)
	}

	_, err = intercept(ctx, &msgWithValidateContext{testSubject{
		err: testErr,
	}}, info, handler)
	if a.So(err, should.BeError) {
		a.So(err, should.EqualErrorOrDefinition, testErr)
	}

	_, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: ttnpb.FieldMask("foo"),
		fieldIsZero: map[string]bool{
			"foo.a.a": true,
			"foo.b":   true,
		},
	}, info, handler)
	a.So(err, should.BeNil)

	_, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: ttnpb.FieldMask("foo.a"),
		fieldIsZero: map[string]bool{
			"foo.a.a": true,
		},
	}, info, handler)
	a.So(err, should.BeNil)

	_, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: ttnpb.FieldMask("foo.a.b"),
	}, info, handler)
	a.So(err, should.BeNil)

	_, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: ttnpb.FieldMask("foo"),
		fieldIsZero: map[string]bool{
			"foo.a.a": false,
			"foo.b":   true,
		},
	}, info, handler)
	if a.So(err, should.BeError) {
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
	}

	_, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: ttnpb.FieldMask("foo"),
		fieldIsZero: map[string]bool{
			"foo.a.a": false,
			"foo.b":   true,
		},
	}, info, handler)
	if a.So(err, should.BeError) {
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
	}

	_, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: ttnpb.FieldMask("bar"),
	}, info, handler)
	if a.So(err, should.BeError) {
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
	}
}

type ss struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *ss) Context() context.Context {
	return ss.ctx
}

func (ss *ss) RecvMsg(_ any) error {
	return nil
}

func TestStreamServerInterceptor(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testErr := errors.New("test")

	RegisterAllowedFieldMaskPaths("/ttn.lorawan.v3.Test/Stream", true, []string{
		"foo",
		"foo.a",
		"foo.a.a",
		"foo.a.b",
		"foo.b",
	},
		"foo",
		"foo.a",
		"foo.a.b",
	)

	info := &grpc.StreamServerInfo{FullMethod: "/ttn.lorawan.v3.Test/Stream"}

	intercept := StreamServerInterceptor()

	err := intercept(nil, &ss{ctx: ctx}, info, func(_ any, stream grpc.ServerStream) error {
		var subject interface{ getResult() *testSubject }

		subject = &testSubject{}
		err := stream.RecvMsg(subject)
		a.So(err, should.BeNil)

		subject = &msgWithValidate{}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeNil) {
			a.So(subject.getResult().validateCalled, should.BeTrue)
		}

		subject = &msgWithValidate{testSubject{
			err: testErr,
		}}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeError) {
			a.So(err, should.EqualErrorOrDefinition, testErr)
		}

		subject = &msgWithValidateContext{}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeNil) {
			a.So(subject.getResult().validateCalled, should.BeTrue)
			a.So(subject.getResult().ctx, should.Equal, ctx)
		}

		subject = &msgWithValidateContext{testSubject{
			err: testErr,
		}}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeError) {
			a.So(err, should.EqualErrorOrDefinition, testErr)
		}

		subject = &msgWithFieldMask{
			fieldMask: ttnpb.FieldMask("foo"),
			fieldIsZero: map[string]bool{
				"foo.a.a": true,
				"foo.b":   true,
			},
		}
		err = stream.RecvMsg(subject)
		a.So(err, should.BeNil)

		subject = &msgWithFieldMask{
			fieldMask: ttnpb.FieldMask("foo.a"),
			fieldIsZero: map[string]bool{
				"foo.a.a": true,
			},
		}
		err = stream.RecvMsg(subject)
		a.So(err, should.BeNil)

		subject = &msgWithFieldMask{
			fieldMask: ttnpb.FieldMask("foo.a.b"),
		}
		err = stream.RecvMsg(subject)
		a.So(err, should.BeNil)

		subject = &msgWithFieldMask{
			fieldMask: ttnpb.FieldMask("foo"),
			fieldIsZero: map[string]bool{
				"foo.a.a": false,
				"foo.b":   true,
			},
		}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeError) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		subject = &msgWithFieldMask{
			fieldMask: ttnpb.FieldMask("foo"),
			fieldIsZero: map[string]bool{
				"foo.a.a": false,
				"foo.b":   true,
			},
		}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeError) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		subject = &msgWithFieldMask{
			fieldMask: ttnpb.FieldMask("bar"),
		}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeError) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		return nil
	})
	a.So(err, should.BeNil)
}
