// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package store

type EUIBlock struct {
	Model

	Type           string `gorm:"unique_index:eui_block_index;type:VARCHAR(10);"`
	StartEUI       EUI64  `gorm:"type:VARCHAR(16);column:start_eui"`
	MaxCounter     int64  `gorm:"type:BIGINT;column:end_counter"`
	CurrentCounter int64  `gorm:"type:BIGINT;column:current_counter"`
}

func init() {
	registerModel(&EUIBlock{})
}

func (EUIBlock) TableName() string {
	return "eui_blocks"
}
