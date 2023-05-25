// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var applicationUpFlags = util.NormalizedFlagSet()

func getStoredUpFlags() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	flags.Bool("stream-output", false, "print output as JSON stream")

	flags.Uint32("f-port", 0, "query upstream messages with specific FPort")
	flags.String("order", "", "order results (received_at|-received_at)")
	flags.Uint32("limit", 0, "limit number of upstream messages to fetch")
	flags.AddFlagSet(timestampFlags("after", "query upstream messages after specified timestamp"))
	flags.AddFlagSet(timestampFlags("before", "query upstream messages before specified timestamp"))
	flags.Duration("last", 0, "query upstream messages in the last hours or minutes")
	flags.String(
		"continuation-token", "",
		"continuation token for pagination (if used additional flags other than the type are ignored)",
	)

	ttnpb.AddSelectFlagsForApplicationUp(flags, "", false)

	types := make([]string, 0, len(ttnpb.StoredApplicationUpTypes))
	for k := range ttnpb.StoredApplicationUpTypes {
		types = append(types, k)
	}
	sort.Strings(types)
	flags.String("type", "", fmt.Sprintf("message type (allowed values: %s)", strings.Join(types, ", ")))

	return flags
}

func getStoredUpRequest(flags *pflag.FlagSet) (*ttnpb.GetStoredApplicationUpRequest, error) {
	req := &ttnpb.GetStoredApplicationUpRequest{}

	req.Type, _ = flags.GetString("type")
	req.ContinuationToken, _ = flags.GetString("continuation-token")
	if req.ContinuationToken != "" {
		return req, nil
	}

	before, after, last, err := timeRangeFromFlags(flags)
	if err != nil {
		return nil, err
	}
	req.Before = before
	req.After = after
	req.Last = last //nolint
	req.Order, _ = flags.GetString("order")

	if flags.Changed("f-port") {
		fport, _ := flags.GetUint32("f-port")
		req.FPort = wrapperspb.UInt32(fport)
	}
	req.FieldMask = ttnpb.FieldMask(
		ttnpb.AllowedFields(
			util.SelectFieldMask(flags, applicationUpFlags),
			ttnpb.RPCFieldMaskPaths["/ttn.lorawan.v3.ApplicationUpStorage/GetStoredApplicationUp"].Allowed,
		)...,
	)

	if flags.Changed("limit") {
		limit, _ := flags.GetUint32("limit")
		req.Limit = &wrapperspb.UInt32Value{
			Value: limit,
		}
	}
	return req, nil
}

func countStoredUpFlags() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	flags.Uint32("f-port", 0, "query upstream messages with specific FPort")
	flags.AddFlagSet(timestampFlags("after", "query upstream messages after specified timestamp"))
	flags.AddFlagSet(timestampFlags("before", "query upstream messages before specified timestamp"))
	flags.Duration("last", 0, "query upstream messages in the last hours or minutes")

	types := make([]string, 0, len(ttnpb.StoredApplicationUpTypes))
	for k := range ttnpb.StoredApplicationUpTypes {
		types = append(types, k)
	}
	sort.Strings(types)
	flags.String("type", "", fmt.Sprintf("message type (allowed values: %s)", strings.Join(types, ", ")))

	return flags
}

func countStoredUpRequest(flags *pflag.FlagSet) (*ttnpb.GetStoredApplicationUpCountRequest, error) {
	before, after, last, err := timeRangeFromFlags(flags)
	if err != nil {
		return nil, err
	}
	req := &ttnpb.GetStoredApplicationUpCountRequest{
		Before: before,
		After:  after,
		Last:   last,
	}
	if flags.Changed("f-port") {
		fport, _ := flags.GetUint32("f-port")
		req.FPort = &wrapperspb.UInt32Value{
			Value: fport,
		}
	}
	req.Type, _ = flags.GetString("type")

	return req, nil
}

func timeRangeFromFlags(flags *pflag.FlagSet) (beforePB *timestamppb.Timestamp, afterPB *timestamppb.Timestamp, lastPB *durationpb.Duration, err error) {
	if flags.Changed("last") && (hasTimestampFlags(flags, "after") || hasTimestampFlags(flags, "before")) {
		return nil, nil, nil, fmt.Errorf("--last cannot be used with --after or --before flags")
	}
	after, err := getTimestampFlags(flags, "after")
	if err != nil {
		return nil, nil, nil, err
	}
	if after != nil {
		afterPB = timestamppb.New(*after)
	}
	before, err := getTimestampFlags(flags, "before")
	if err != nil {
		return nil, nil, nil, err
	}
	if before != nil {
		beforePB = timestamppb.New(*before)
	}

	if flags.Changed("last") {
		d, err := flags.GetDuration("last")
		if err != nil {
			return nil, nil, nil, err
		}
		lastPB = durationpb.New(d)
	}
	return
}

type continuationToken struct {
	ContinuationToken string `json:"continuation_token,omitempty"`
}

var errNoContinuationToken = errors.DefineUnavailable("no_continuation_token", "no continuation token")

func newContinuationTokenFromMD(md metadata.MD) (*continuationToken, error) {
	continuationTokenHeaderValues := md.Get("x-continuation-token")
	if len(continuationTokenHeaderValues) == 1 {
		return &continuationToken{
			ContinuationToken: continuationTokenHeaderValues[0],
		}, nil
	}
	return nil, errNoContinuationToken
}
