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

package commands

import (
	"strconv"

	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func paginationFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Uint32("limit", 50, "maximum number of results to get")
	flagSet.Uint32("page", 1, "results page number")
	return flagSet
}

func withPagination(flagSet *pflag.FlagSet) (limit, page uint32, opt grpc.CallOption, getTotal func() uint64) {
	limit, _ = flagSet.GetUint32("limit")
	page, _ = flagSet.GetUint32("page")
	responseHeaders := metadata.MD{}
	opt = grpc.Header(&responseHeaders)
	getTotal = func() uint64 {
		totalHeader := responseHeaders.Get("x-total-count")
		if len(totalHeader) > 0 {
			total, _ := strconv.ParseUint(totalHeader[len(totalHeader)-1], 10, 64)
			if total != 0 && total > uint64(limit)*uint64(page) {
				logger.WithField("total", total).Infof("Use the flags \"--limit=%d --page=%d\" to get the next page of results", limit, page+1)
			} else {
				logger.Debugf("Total results: %d", total)
			}
			return total
		}
		return 0
	}
	return
}
