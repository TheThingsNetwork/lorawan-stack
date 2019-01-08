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

import faker from 'faker'

const APPLICATIONS_COUNT = 10
const DEVICES_COUNT = 80
const GATEWAYS_COUNT = 15

const applications = [ ...new Array(APPLICATIONS_COUNT).keys() ]
  .map(_ => ({
    application_id: faker.random.uuid(),
    name: faker.random.word(),
    description: faker.random.words(),
    created_at: faker.date.past(),
    updated_at: faker.date.recent(),
  }))

const rights = {
  application: [
    'RIGHT_APPLICATION_INFO',
    'RIGHT_APPLICATION_SETTINGS_BASIC',
    'RIGHT_APPLICATION_SETTINGS_API_KEYS',
    'RIGHT_APPLICATION_SETTINGS_COLLABORATORS',
    'RIGHT_APPLICATION_DELETE',
    'RIGHT_APPLICATION_DEVICES_READ',
    'RIGHT_APPLICATION_DEVICES_WRITE',
    'RIGHT_APPLICATION_DEVICES_READ_KEYS',
    'RIGHT_APPLICATION_DEVICES_WRITE_KEYS',
    'RIGHT_APPLICATION_TRAFFIC_READ',
    'RIGHT_APPLICATION_TRAFFIC_UP_WRITE',
    'RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE',
    'RIGHT_APPLICATION_LINK',
  ],
}

const getRights = function (from, min, max) {
  const count = Math.floor(Math.random() * (max - min)) + min

  const res = []
  for (let i = 0; i < count; i++) {
    res.push(from[i])
  }

  return res
}

const applicationsApiKeys = applications.reduce(function (acc, curr, idx, apps) {
  const keysCount = Math.floor(Math.random() * 5) + 1

  for (let i = 0; i < keysCount; i++) {
    acc.push({
      id: faker.random.uuid(),
      application_id: curr.application_id,
      name: faker.lorem.words(),
      key: faker.random.uuid(),
      rights: getRights(rights.application, 1, 5),
    })
  }

  return acc
}, [])

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

const generateGatewayEUI = function () {
  let res = 'eui-'
  for (let i = 0; i < 16; i++) {
    res += faker.random.alphaNumeric()
  }

  return res
}

const gatewayAntennas = [ ...new Array(GATEWAYS_COUNT).keys() ]
  .map(_ => ({
    gain: Math.floor(Math.random() * 5),
    location: {
      latitude: faker.address.latitude(),
      longitude: faker.address.longitude(),
    },
  }))

const gateways = [ ...new Array(GATEWAYS_COUNT).keys() ]
  .map((_, idx) => ({
    gateway_id: faker.random.uuid(),
    eui: generateGatewayEUI(),
    name: faker.lorem.word(),
    description: faker.lorem.words(),
    created_at: faker.date.past(),
    updated_at: faker.date.recent(),
    frequency_plan: 'EU_863_870',
    antennas: [ gatewayAntennas[idx] ],
  }))

export default { devices, applications, gateways, applicationsApiKeys }
