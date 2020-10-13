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
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func getStoredUpFlags() *pflag.FlagSet {
	flags := &pflag.FlagSet{}

	flags.Bool("stream-output", false, "print output as JSON stream")

	flags.Uint32("f-port", 0, "query upstream messages with specific FPort")
	flags.String("order", "", "order results (received_at|-received_at)")
	flags.Uint32("limit", 0, "limit number of upstream messages to fetch")
	flags.AddFlagSet(timestampFlags("after", "query upstream messages after specified timestamp"))
	flags.AddFlagSet(timestampFlags("before", "query upstream messages before specified timestamp"))

	types := make([]string, 0, len(ttnpb.StoredApplicationUpTypes))
	for k := range ttnpb.StoredApplicationUpTypes {
		types = append(types, k)
	}
	flags.String("type", "", fmt.Sprintf("message type (%s)", strings.Join(types, "|")))

	return flags
}

func getStoredUpRequest(flags *pflag.FlagSet) (*ttnpb.GetStoredApplicationUpRequest, error) {
	var err error
	req := &ttnpb.GetStoredApplicationUpRequest{}

	req.After, err = getTimestampFlags(flags, "after")
	if err != nil {
		return nil, err
	}
	req.Before, err = getTimestampFlags(flags, "before")
	if err != nil {
		return nil, err
	}
	req.Order, _ = flags.GetString("order")
	req.Type, _ = flags.GetString("type")

	if flags.Changed("f-port") {
		fport, _ := flags.GetUint32("f-port")
		req.FPort = &pbtypes.UInt32Value{
			Value: fport,
		}
	}

	if flags.Changed("limit") {
		limit, _ := flags.GetUint32("limit")
		req.Limit = &pbtypes.UInt32Value{
			Value: limit,
		}
	}
	return req, nil
}
