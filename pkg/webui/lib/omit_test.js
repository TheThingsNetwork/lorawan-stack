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

import omitDeep from './omit'

const stateObject = {
  user: {
    user: {
      ids: {
        user_id: 'admin',
      },
      created_at: '2020-05-07T11:50:32.549Z',
      updated_at: '2020-05-07T11:50:32.549Z',
      primary_email_address_validated_at: '2020-05-07T11:50:32.532582Z',
      state: 'STATE_APPROVED',
      isAdmin: true,
    },
    rights: ['RIGHT_APPLICATION_DEVICES_WRITE_KEYS'],
  },
  users: {
    entities: {},
    selectedUser: null,
  },
  init: {
    initialized: true,
  },
  applications: {
    entities: {
      app2: {
        ids: {
          application_id: 'app2',
        },
        created_at: '2020-05-13T10:25:39.249Z',
        updated_at: '2020-05-13T10:25:39.249Z',
      },
    },
    selectedApplication: null,
  },
  link: {},
  devices: {
    entities: {},
  },
  gateways: {
    entities: {
      gateway1: {
        ids: {
          gateway_id: '243',
        },
        created_at: '2020-05-12T08:14:45.667Z',
        updated_at: '2020-05-12T09:48:45.844Z',
        version_ids: {},
      },
    },
    selectedGateway: null,
    statistics: {},
  },
  webhooks: {
    selectedWebhook: null,
    entities: {},
  },
  webhookFormats: {},
  webhookTemplates: {},
  deviceTemplateFormats: {},
  pubsubs: {
    selectedPubsub: null,
    entities: {},
  },
  pubsubFormats: {},
  configuration: {},
  organizations: {
    entities: {},
    selectedOrganization: null,
  },
  apiKeys: {
    entities: {},
    selectedApiKey: null,
  },
  collaborators: {
    entities: {},
    selectedCollaborator: null,
  },
  rights: {
    applications: {
      rights: [],
    },
    gateways: {
      rights: [],
    },
    organizations: {
      rights: [],
    },
  },
  events: {
    applications: {},
    devices: {},
    gateways: {},
    organizations: {},
  },
  ui: {
    fetching: {
      INITIALIZE: false,
      GET_USER_RIGHTS: false,
      GET_USER_ME: false,
      GET_APPLICATION_LIST: false,
      GET_GATEWAY_LIST: false,
    },
    error: {},
  },
  pagination: {
    applications: {
      ids: ['app2', 'new-app'],
      totalCount: 2,
    },
    apiKeys: {
      ids: [],
    },
    gateways: {
      ids: ['243', 'my-new-gateway', 'new-gtw'],
      totalCount: 3,
    },
  },
  router: {
    location: {
      pathname: '/',
      search: '',
      hash: '',
    },
    action: 'POP',
  },
  js: {
    prefixes: [],
  },
  gatewayStatus: {},
}

describe('Omit utils', function () {
  describe('when object is empty', function () {
    const object = {}
    it('returns same object', function () {
      const result = omitDeep(object, ['value'])

      expect(result).toStrictEqual(object)
    })
  })

  describe('when array of values is empty', function () {
    const values = []
    it('returns same object', function () {
      const result = omitDeep(stateObject, values)

      expect(result).toStrictEqual(stateObject)
    })
  })

  describe('when object and values are empty', function () {
    const object = {}
    const values = []
    it('returns same object', function () {
      const result = omitDeep(object, values)

      expect(result).toStrictEqual(object)
    })
  })

  describe('when excluding the single top level property', function () {
    const values = ['gateways']

    it('omits object `gateways` property', function () {
      const result = omitDeep(stateObject, values)

      expect(result.gateways).toBeUndefined()
    })
  })

  describe('when excluding multiple top level and nested properties', function () {
    const values = ['apiKeys', 'ids']

    it('omits all occurrences of `apiKeys` and `ids`', function () {
      const result = omitDeep(stateObject, values)

      expect(result.pagination.gateways.totalCount).toStrictEqual(
        stateObject.pagination.gateways.totalCount,
      )
      expect(result.pubsubs).toStrictEqual(stateObject.pubsubs)
      expect(result.user.user.ids).toBeUndefined()
      expect(result.applications.entities.app2.ids).toBeUndefined()
      expect(result.gateways.entities.gateway1.ids).toBeUndefined()
      expect(result.apiKeys).toBeUndefined()
      expect(result.pagination.applications.ids).toBeUndefined()
      expect(result.pagination.gateways.ids).toBeUndefined()
      expect(result.pagination.apiKeys).toBeUndefined()
    })
  })
})
