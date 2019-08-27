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

package store

import "github.com/jinzhu/gorm"

// AfterDelete deletes the Account of an Organization after it is deleted.
func (org *Organization) AfterDelete(db *gorm.DB) error {
	return db.Where(Account{
		AccountType: "organization",
		AccountID:   org.PrimaryKey(),
	}).Delete(Account{}).Error
}

// AfterDelete deletes the Account of a User after it is deleted.
func (usr *User) AfterDelete(db *gorm.DB) error {
	return db.Where(Account{
		AccountType: "user",
		AccountID:   usr.PrimaryKey(),
	}).Delete(Account{}).Error
}
