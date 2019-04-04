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
  .map(() => ({
    ids: {
      application_id: faker.lorem.slug(),
    },
    name: faker.random.word(),
    description: faker.random.words(),
    created_at: faker.date.past(),
    updated_at: faker.date.recent(),
  }))

const rights = {
  applications: [
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
      application_id: curr.ids.application_id,
      name: faker.lorem.words(),
      key: faker.random.uuid(),
      rights: getRights(rights.applications, 1, 5),
    })
  }

  return acc
}, [])

const devices = [ ...new Array(DEVICES_COUNT).keys() ]
  .map(function () {

    return {
      ids: {
        device_id: faker.lorem.slug(),
        application_ids: {
          application_id: '',
        },
        dev_eui: faker.random.alphaNumeric(16),
        join_eui: faker.random.alphaNumeric(16),
        dev_addr: faker.random.alphaNumeric(8),
      },
      name: faker.random.word(),
      description: faker.random.words(),
      created_at: faker.date.past(),
      updated_at: faker.date.recent(),
      attributes: {
        [faker.lorem.slug()]: faker.lorem.slug(),
        [faker.lorem.slug()]: faker.lorem.slug(),
      },
      version_ids: {
        brand_id: faker.lorem.slug(),
        model_id: faker.lorem.slug(),
        hardware_version: faker.system.semver(),
        firmware_version: faker.system.semver(),
      },
      network_server_address: faker.internet.url(),
      application_server_address: faker.internet.url(),
      // JS
      root_keys: {
        root_key_id: faker.random.uuid(),
        app_key: {
          key: `ttn-account-v3.${faker.random.alphaNumeric(8)}`,
          kek_label: faker.lorem.slug(),
        },
        nwk_key: {
          key: `ttn-account-v3.${faker.random.alphaNumeric(8)}`,
          kek_label: faker.lorem.slug(),
        },
      },
      // NS and AS
      session: {
        // Known by Network Server, Application Server and Join Server. Owned by Network Server.
        dev_addr: `00${faker.random.number()}`,
        keys: {
          session_key_id: faker.random.uuid(), // JS
          // NS
          f_nwk_s_int_key: {
            key: faker.random.alphaNumeric(32),
            kek_label: faker.lorem.slug(),
          },
          // NS
          s_nwk_s_int_key: {
            key: faker.random.alphaNumeric(32),
            kek_label: faker.lorem.slug(),
          },
          // NS
          nwk_s_enc_key: {
            key: faker.random.alphaNumeric(32),
            kek_label: faker.lorem.slug(),
          },
          // AS
          app_s_key: {
            key: faker.random.alphaNumeric(32),
            kek_label: faker.lorem.slug(),
          },
        },
        next_f_cnt_up: faker.random.number(), // NS
        next_n_f_cnt_down: faker.random.number(), // NS
        next_a_f_cnt_down: faker.random.number(), // AS
        last_conf_f_cnt_down: faker.random.number(), // NS
        started_at: faker.date.recent(), // NS
      },
      // AS
      formatters: {
        up_formatter: 'FORMATTER_NONE',
        up_formatter_parameter: '',
        down_formatter: 'FORMATTER_NONE',
        down_formatter_parameter: '',
      },
    }
  })

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
    ids: {
      gateway_id: faker.random.uuid(),
      eui: faker.random.alphaNumeric(16),
    },
    name: faker.lorem.word(),
    description: faker.lorem.words(),
    created_at: faker.date.past(),
    updated_at: faker.date.recent(),
    frequency_plan_id: 'EU_863_870',
    antennas: [ gatewayAntennas[idx] ],
    version_ids: {
      brand_id: faker.lorem.slug(),
      model_id: faker.lorem.slug(),
      hardware_version: faker.system.semver(),
      firmware_version: faker.system.semver(),
    },
  }))

export default { devices, applications, gateways, applicationsApiKeys, rights }
