// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import faker from 'faker'

const APPLICATIONS_COUNT = 10
const DEVICES_COUNT = 80

const applications = [ ...new Array(APPLICATIONS_COUNT).keys() ]
  .map(_ => ({
    application_id: faker.random.uuid(),
    name: faker.random.word(),
    description: faker.random.words(),
    created_at: faker.date.past(),
    updated_at: faker.date.recent(),
  }))


const devices = [ ...new Array(DEVICES_COUNT).keys() ]
  .map(function () {
    const app = applications[Math.floor(Math.random() * APPLICATIONS_COUNT)]

    return {
      device_id: faker.random.uuid(),
      application_id: app.application_id,
      name: faker.random.word(),
      description: faker.random.words(),
      created_at: faker.date.past(),
      updated_at: faker.date.recent(),
    }
  })

export default { devices, applications }
