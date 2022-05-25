// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func tableExists(ctx context.Context, db *bun.DB, tableName string) (bool, error) {
	c, err := db.NewSelect().
		TableExpr("INFORMATION_SCHEMA.tables").
		Where("table_name = ?", tableName).
		Where("table_type = 'BASE TABLE'").
		Where("table_schema = CURRENT_SCHEMA()").
		Count(ctx)
	if err != nil {
		return false, err
	}
	return c == 1, nil
}
