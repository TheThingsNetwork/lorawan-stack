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

package redis_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestPagination(t *testing.T) {
	a := assertions.New(t)

	for _, tc := range []struct {
		limit          int64
		page           int64
		expectedLimit  uint64
		expectedOffset uint64
	}{
		{
			limit:          0,
			page:           0,
			expectedLimit:  0,
			expectedOffset: 0,
		},
		{
			limit:          10,
			page:           0,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			limit:          10,
			page:           1,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			limit:          10,
			page:           2,
			expectedLimit:  10,
			expectedOffset: 10,
		},
	} {
		t.Run(fmt.Sprintf("limitAndOffsetFromContext, limit:%v, offset:%v", tc.limit, tc.page),
			func(t *testing.T) {
				ctx := redis.NewContextWithPagination(test.Context(), tc.limit, tc.page, nil)

				limit, offset := redis.PaginationLimitAndOffsetFromContext(ctx)

				a.So(limit, should.Equal, tc.expectedLimit)
				a.So(offset, should.Equal, tc.expectedOffset)
			},
		)
	}

	t.Run("SetTotalCount", func(t *testing.T) {
		var totalCount int64
		ctx := test.Context()
		total := int64(10)

		redis.SetPaginationTotal(ctx, total)
		a.So(totalCount, should.BeZeroValue)

		ctx = redis.NewContextWithPagination(ctx, 5, 1, &totalCount)
		a.So(totalCount, should.BeZeroValue)

		redis.SetPaginationTotal(ctx, total)
		a.So(totalCount, should.Equal, total)
	})
}
