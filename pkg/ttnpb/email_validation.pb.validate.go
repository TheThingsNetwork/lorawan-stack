// Code generated by protoc-gen-fieldmask. DO NOT EDIT.

package ttnpb

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
)

// ValidateFields checks the field values on EmailValidation with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *EmailValidation) ValidateFields(paths ...string) error {
	if m == nil {
		return nil
	}

	if len(paths) == 0 {
		paths = EmailValidationFieldPathsNested
	}

	for name, subs := range _processPaths(append(paths[:0:0], paths...)) {
		_ = subs
		switch name {
		case "id":

			if l := utf8.RuneCountInString(m.GetId()); l < 1 || l > 64 {
				return EmailValidationValidationError{
					field:  "id",
					reason: "value length must be between 1 and 64 runes, inclusive",
				}
			}

		case "token":

			if l := utf8.RuneCountInString(m.GetToken()); l < 1 || l > 64 {
				return EmailValidationValidationError{
					field:  "token",
					reason: "value length must be between 1 and 64 runes, inclusive",
				}
			}

		case "address":

			if err := m._validateEmail(m.GetAddress()); err != nil {
				return EmailValidationValidationError{
					field:  "address",
					reason: "value must be a valid email address",
					cause:  err,
				}
			}

		case "created_at":

			if v, ok := interface{}(m.GetCreatedAt()).(interface{ ValidateFields(...string) error }); ok {
				if err := v.ValidateFields(subs...); err != nil {
					return EmailValidationValidationError{
						field:  "created_at",
						reason: "embedded message failed validation",
						cause:  err,
					}
				}
			}

		case "expires_at":

			if v, ok := interface{}(m.GetExpiresAt()).(interface{ ValidateFields(...string) error }); ok {
				if err := v.ValidateFields(subs...); err != nil {
					return EmailValidationValidationError{
						field:  "expires_at",
						reason: "embedded message failed validation",
						cause:  err,
					}
				}
			}

		case "updated_at":

			if v, ok := interface{}(m.GetUpdatedAt()).(interface{ ValidateFields(...string) error }); ok {
				if err := v.ValidateFields(subs...); err != nil {
					return EmailValidationValidationError{
						field:  "updated_at",
						reason: "embedded message failed validation",
						cause:  err,
					}
				}
			}

		default:
			return EmailValidationValidationError{
				field:  name,
				reason: "invalid field path",
			}
		}
	}
	return nil
}

func (m *EmailValidation) _validateHostname(host string) error {
	s := strings.ToLower(strings.TrimSuffix(host, "."))

	if len(host) > 253 {
		return errors.New("hostname cannot exceed 253 characters")
	}

	for _, part := range strings.Split(s, ".") {
		if l := len(part); l == 0 || l > 63 {
			return errors.New("hostname part must be non-empty and cannot exceed 63 characters")
		}

		if part[0] == '-' {
			return errors.New("hostname parts cannot begin with hyphens")
		}

		if part[len(part)-1] == '-' {
			return errors.New("hostname parts cannot end with hyphens")
		}

		for _, r := range part {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
				return fmt.Errorf("hostname parts can only contain alphanumeric characters or hyphens, got %q", string(r))
			}
		}
	}

	return nil
}

func (m *EmailValidation) _validateEmail(addr string) error {
	a, err := mail.ParseAddress(addr)
	if err != nil {
		return err
	}
	addr = a.Address

	if len(addr) > 254 {
		return errors.New("email addresses cannot exceed 254 characters")
	}

	parts := strings.SplitN(addr, "@", 2)

	if len(parts[0]) > 64 {
		return errors.New("email address local phrase cannot exceed 64 characters")
	}

	return m._validateHostname(parts[1])
}

// EmailValidationValidationError is the validation error returned by
// EmailValidation.ValidateFields if the designated constraints aren't met.
type EmailValidationValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e EmailValidationValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e EmailValidationValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e EmailValidationValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e EmailValidationValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e EmailValidationValidationError) ErrorName() string { return "EmailValidationValidationError" }

// Error satisfies the builtin error interface
func (e EmailValidationValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sEmailValidation.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = EmailValidationValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = EmailValidationValidationError{}

// ValidateFields checks the field values on ValidateEmailRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *ValidateEmailRequest) ValidateFields(paths ...string) error {
	if m == nil {
		return nil
	}

	if len(paths) == 0 {
		paths = ValidateEmailRequestFieldPathsNested
	}

	for name, subs := range _processPaths(append(paths[:0:0], paths...)) {
		_ = subs
		switch name {
		case "id":

			if l := utf8.RuneCountInString(m.GetId()); l < 1 || l > 64 {
				return ValidateEmailRequestValidationError{
					field:  "id",
					reason: "value length must be between 1 and 64 runes, inclusive",
				}
			}

		case "token":

			if l := utf8.RuneCountInString(m.GetToken()); l < 1 || l > 64 {
				return ValidateEmailRequestValidationError{
					field:  "token",
					reason: "value length must be between 1 and 64 runes, inclusive",
				}
			}

		default:
			return ValidateEmailRequestValidationError{
				field:  name,
				reason: "invalid field path",
			}
		}
	}
	return nil
}

// ValidateEmailRequestValidationError is the validation error returned by
// ValidateEmailRequest.ValidateFields if the designated constraints aren't met.
type ValidateEmailRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ValidateEmailRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ValidateEmailRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ValidateEmailRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ValidateEmailRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ValidateEmailRequestValidationError) ErrorName() string {
	return "ValidateEmailRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ValidateEmailRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sValidateEmailRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ValidateEmailRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ValidateEmailRequestValidationError{}
