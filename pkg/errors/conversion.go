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

package errors

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"os"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	errNetInvalidAddr = DefineInvalidArgument(
		"net_invalid_addr", "{message}", "timeout",
	)
	errNetAddr = DefineUnavailable(
		"net_addr", "{message}", "timeout",
	)
	errNetDNS = DefineUnavailable(
		"net_dns", "{message}", "timeout", "not_found",
	)
	errNetUnknownNetwork = DefineNotFound(
		"net_unknown_network", "{message}", "timeout",
	)
	errNetOperation = DefineUnavailable(
		"net_operation", "net operation failed", "op", "net", "source", "address", "timeout",
	)
	errNetTimeout = DefineUnavailable("net_timeout", "operation timed out")

	errRequest = Define("request", "request to `{url}` failed", "op")
	errURL     = DefineInvalidArgument("url", "invalid url `{url}`", "op")

	errSyscall = Define("syscall", "`{syscall}` failed", "syscall", "error", "timeout")

	errX509UnknownAuthority = DefineUnavailable(
		"x509_unknown_authority", "unknown certificate authority",
	)
	errX509Hostname = DefineUnavailable(
		"x509_hostname", "certificate authorized names do not match the requested name", "host",
	)
	errX509CertificateInvalid = DefineUnavailable(
		"x509_certificate_invalid", "certificate invalid", "detail", "reason",
	)

	// ErrContextCanceled is the Definition of the standard context.Cancelled error.
	// This variant exists in order to allow the error code to be properly propagated
	// over gRPC calls, otherwise the error code is unknown.
	ErrContextCanceled = DefineCanceled("context_canceled", "context canceled")
	// ErrContextDeadlineExceeded is the definition of the standard context.DeadlineExceeded error.
	// This variant exists in order to allow the error code to be properly propagated
	// over gRPC calls, otherwise the error code is unknown.
	ErrContextDeadlineExceeded = DefineDeadlineExceeded(
		"context_deadline_exceeded", "context deadline exceeded",
	)
)

// From returns an *Error if it can be derived from the given input.
// For a nil error, false will be returned.
func From(err error) (out *Error, ok bool) { //nolint:gocyclo
	if err == nil {
		return nil, false
	}
	defer func() {
		if out != nil {
			clone := *out
			out = &clone
		}
	}()
	if errors.Is(err, context.Canceled) {
		return build(ErrContextCanceled, 0), true
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return build(ErrContextDeadlineExceeded, 0), true
	}
	if matched := (*Error)(nil); errors.As(err, &matched) {
		if matched == nil {
			return nil, false
		}
		return matched, true
	}
	if matched := (*Definition)(nil); errors.As(err, &matched) {
		if matched == nil {
			return nil, false
		}
		return build(matched, 0), true
	}
	if matched := (ErrorDetails)(nil); errors.As(err, &matched) {
		var e Error
		setErrorDetails(&e, matched)
		return &e, true
	}
	if matched := (interface{ GRPCStatus() *status.Status })(nil); errors.As(err, &matched) {
		return FromGRPCStatus(matched.GRPCStatus()), true
	}
	if matched := (validationError)(nil); errors.As(err, &matched) {
		e := build(errValidation, 0).WithAttributes(
			"field", matched.Field(),
			"reason", matched.Reason(),
			"name", matched.ErrorName(),
		)
		if cause := matched.Cause(); cause != nil {
			e = e.WithCause(cause)
		}
		return e, true
	}
	if matched := (*net.DNSError)(nil); errors.As(err, &matched) {
		e := build(errNetDNS, 0).WithAttributes(
			"not_found", matched.IsNotFound,
		).WithAttributes(
			netErrorDetails(matched)...,
		)
		return e, true
	}
	if matched := (*net.AddrError)(nil); errors.As(err, &matched) {
		return build(errNetAddr, 0).WithAttributes(netErrorDetails(matched)...), true
	}
	if matched := (net.InvalidAddrError)(""); errors.As(err, &matched) {
		return build(errNetInvalidAddr, 0).WithAttributes(netErrorDetails(matched)...), true
	}
	if matched := (net.UnknownNetworkError)(""); errors.As(err, &matched) {
		return build(errNetUnknownNetwork, 0).WithAttributes(netErrorDetails(matched)...), true
	}
	if matched := (*net.OpError)(nil); errors.As(err, &matched) {
		// Do not use netErrorDetails(err) as err.Error() will panic if err.Err is nil.
		e := build(errNetOperation, 0).WithAttributes(
			"op", matched.Op,
			"net", matched.Net,
			"timeout", matched.Timeout(),
		)
		if matched.Addr != nil {
			e = e.WithAttributes("address", matched.Addr.String())
		}
		if matched.Source != nil {
			e = e.WithAttributes("source", matched.Source.String())
		}
		if matched.Err != nil {
			e = e.WithCause(matched.Err)
		}
		return e, true
	}
	if matched := (*url.Error)(nil); errors.As(err, &matched) {
		definition := errRequest
		if matched.Op == "parse" {
			definition = errURL
		}
		e := build(definition, 0).WithAttributes(
			"url", matched.URL,
			"op", matched.Op,
		)
		if matched.Err != nil {
			e = e.WithCause(matched.Err)
		}
		return e, true
	}
	if matched := (net.Error)(nil); errors.As(err, &matched) && matched.Timeout() {
		return build(errNetTimeout, 0), true
	}
	if matched := (*os.SyscallError)(nil); errors.As(err, &matched) {
		e := build(errSyscall, 0).WithAttributes("syscall", matched.Syscall)
		if err := matched.Err; err != nil {
			e = e.WithAttributes(syscallErrorAttributes(err)...)
			e = e.WithCause(err)
		}
		return e, true
	}
	if matched := (x509.CertificateInvalidError{}); errors.As(err, &matched) {
		return build(errX509CertificateInvalid, 0).WithAttributes(
			"detail", matched.Detail,
			"reason", matched.Reason,
		), true
	}
	if matched := (x509.UnknownAuthorityError{}); errors.As(err, &matched) {
		return build(errX509UnknownAuthority, 0), true
	}
	if matched := (x509.HostnameError{}); errors.As(err, &matched) {
		return build(errX509Hostname, 0).WithAttributes(
			"host", matched.Host,
		), true
	}
	return nil, false
}

// ErrorDetails that can be carried over API.
type ErrorDetails interface {
	Error() string
	Namespace() string
	Name() string
	MessageFormat() string
	PublicAttributes() map[string]any
	CorrelationID() string
	Cause() error
	Code() uint32
	Details() []proto.Message
}

func setErrorDetails(err *Error, details ErrorDetails) {
	if err.Definition == nil {
		err.Definition = &Definition{}
	}
	if namespace := details.Namespace(); namespace != "" {
		err.namespace = namespace
	}
	if name := details.Name(); name != "" {
		err.name = name
	}
	if messageFormat := details.MessageFormat(); messageFormat != "" {
		err.messageFormat = messageFormat
		err.messageFormatArguments = messageFormatArguments(messageFormat)
		err.parsedMessageFormat, _ = formatter.Parse(messageFormat)
	}
	if attributes := details.PublicAttributes(); len(attributes) != 0 {
		err.attributes = attributes
		for attr := range attributes {
			err.publicAttributes = append(err.publicAttributes, attr)
		}
	}
	if correlationID := details.CorrelationID(); correlationID != "" {
		err.correlationID = correlationID
	}
	if cause := details.Cause(); cause != nil {
		err.cause, _ = From(cause)
	}
	if code := details.Code(); code != 0 {
		err.code = code
	}
	err.details = append(err.details, details.Details()...)
}

func netErrorDetails(err net.Error) []any {
	return []any{
		"message", err.Error(),
		"timeout", err.Timeout(),
	}
}
