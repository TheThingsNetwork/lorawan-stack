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
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/rpcmiddleware/validator"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
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
	fieldMask types.FieldMask
}

func (m *msgWithFieldMask) GetFieldMask() types.FieldMask { return m.fieldMask }

func (m *msgWithFieldMask) ValidateFields(...string) error { return nil }

func handler(ctx context.Context, req interface{}) (interface{}, error) {
	res := req.(interface{ getResult() *testSubject }).getResult()
	res.handlerCalled = true
	return res, nil
}

func TestUnaryServerInterceptor(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testErr := errors.New("test")

	RegisterAllowedFieldMaskPaths("/ttn.lorawan.v3.Test/Unary", "foo")

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

	res, err = intercept(ctx, &msgWithValidate{testSubject{
		err: testErr,
	}}, info, handler)
	if a.So(err, should.BeError) {
		a.So(err, should.Resemble, &testErr)
	}

	res, err = intercept(ctx, &msgWithValidateContext{}, info, handler)
	if a.So(err, should.BeNil) {
		a.So(res.(*testSubject).validateCalled, should.BeTrue)
		a.So(res.(*testSubject).handlerCalled, should.BeTrue)
		a.So(res.(*testSubject).ctx, should.Equal, ctx)
	}

	res, err = intercept(ctx, &msgWithValidateContext{testSubject{
		err: testErr,
	}}, info, handler)
	if a.So(err, should.BeError) {
		a.So(err, should.Resemble, &testErr)
	}

	res, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: types.FieldMask{Paths: []string{"foo"}},
	}, info, handler)
	a.So(err, should.BeNil)

	res, err = intercept(ctx, &msgWithFieldMask{
		fieldMask: types.FieldMask{Paths: []string{"bar"}},
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

func (ss *ss) RecvMsg(_ interface{}) error {
	return nil
}

func TestStreamServerInterceptor(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testErr := errors.New("test")

	RegisterAllowedFieldMaskPaths("/ttn.lorawan.v3.Test/Stream", "foo")

	info := &grpc.StreamServerInfo{FullMethod: "/ttn.lorawan.v3.Test/Stream"}

	intercept := StreamServerInterceptor()

	err := intercept(nil, &ss{ctx: ctx}, info, func(_ interface{}, stream grpc.ServerStream) error {

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
			a.So(err, should.Resemble, &testErr)
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
			a.So(err, should.Resemble, &testErr)
		}

		subject = &msgWithFieldMask{
			fieldMask: types.FieldMask{Paths: []string{"foo"}},
		}
		err = stream.RecvMsg(subject)
		a.So(err, should.BeNil)

		subject = &msgWithFieldMask{
			fieldMask: types.FieldMask{Paths: []string{"bar"}},
		}
		err = stream.RecvMsg(subject)
		if a.So(err, should.BeError) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		return nil
	})
	a.So(err, should.BeNil)
}
