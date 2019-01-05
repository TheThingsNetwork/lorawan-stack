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

package validator

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
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

func handler(ctx context.Context, req interface{}) (interface{}, error) {
	res := req.(interface{ getResult() *testSubject }).getResult()
	res.handlerCalled = true
	return res, nil
}

func TestUnaryServerInterceptor(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	intercept := UnaryServerInterceptor()

	res, err := intercept(ctx, &testSubject{}, nil, handler)
	a.So(err, should.BeNil)
	a.So(res.(*testSubject).handlerCalled, should.BeTrue)

	res, err = intercept(ctx, &msgWithValidate{}, nil, handler)
	a.So(err, should.BeNil)
	a.So(res.(*testSubject).validateCalled, should.BeTrue)
	a.So(res.(*testSubject).handlerCalled, should.BeTrue)

	res, err = intercept(ctx, &msgWithValidate{testSubject{
		err: errors.New("foo"),
	}}, nil, handler)
	a.So(err, should.NotBeNil)

	res, err = intercept(ctx, &msgWithValidateContext{}, nil, handler)
	a.So(err, should.BeNil)
	a.So(res.(*testSubject).validateCalled, should.BeTrue)
	a.So(res.(*testSubject).ctx, should.Equal, ctx)
	a.So(res.(*testSubject).handlerCalled, should.BeTrue)

	res, err = intercept(ctx, &msgWithValidateContext{testSubject{
		err: errors.New("foo"),
	}}, nil, handler)
	a.So(err, should.NotBeNil)
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

	intercept := StreamServerInterceptor()

	err := intercept(nil, &ss{ctx: ctx}, nil, func(_ interface{}, stream grpc.ServerStream) error {

		var subject interface{ getResult() *testSubject }

		subject = &testSubject{}
		err := stream.RecvMsg(subject)
		a.So(err, should.BeNil)

		subject = &msgWithValidate{}
		err = stream.RecvMsg(subject)
		a.So(err, should.BeNil)
		a.So(subject.getResult().validateCalled, should.BeTrue)

		subject = &msgWithValidate{testSubject{
			err: errors.New("foo"),
		}}
		err = stream.RecvMsg(subject)
		a.So(err, should.NotBeNil)

		subject = &msgWithValidateContext{}
		err = stream.RecvMsg(subject)
		a.So(err, should.BeNil)
		a.So(subject.getResult().validateCalled, should.BeTrue)
		a.So(subject.getResult().ctx, should.Equal, ctx)

		subject = &msgWithValidateContext{testSubject{
			err: errors.New("foo"),
		}}
		err = stream.RecvMsg(subject)
		a.So(err, should.NotBeNil)

		return nil
	})

	a.So(err, should.BeNil)

}
