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

import Applications from './applications'

const mockApplicationData = {
  ids: {
    application_id: 'test',
  },
  created_at: '2018-08-29T14:00:20.793Z',
  updated_at: '2018-08-29T14:00:20.793Z',
  name: 'string',
  description: 'string',
  attributes: {
    additionalProp1: 'string',
    additionalProp2: 'string',
    additionalProp3: 'string',
  },
  contact_info: [
    {
      contact_type: 'CONTACT_TYPE_OTHER',
      contact_method: 'CONTACT_METHOD_OTHER',
      value: 'string',
      public: true,
      validated_at: '2018-08-29T14:00:20.793Z',
    },
  ],
}

jest.mock('../api', function() {
  return jest.fn().mockImplementation(function() {
    return {
      ApplicationRegistry: {
        Get: jest.fn().mockResolvedValue({ data: mockApplicationData }),
        List: jest.fn().mockResolvedValue({
          data: { applications: [mockApplicationData] },
          headers: { 'x-total-count': 1 },
        }),
      },
    }
  })
})

describe('Applications', function() {
  let applications
  beforeEach(function() {
    const Api = require('../api')

    const Applications = require('./applications').default
    applications = new Applications(new Api(), { defaultUserId: 'testuser' })
  })

  describe('Proxied results', function() {
    test('instance initializes correctly', function() {
      jest.resetModules()

      expect(applications).toBeInstanceOf(Applications)
      expect(applications._api).toBeDefined()
    })

    test('instance returns an application instance on getById()', async function() {
      jest.resetModules()

      const app = await applications.getById('test')
      expect(app).toBeDefined()
      expect(app.constructor.name).toBe('Application')
      expect(app.ids.application_id).toBe('test')
    })

    test('instance returns an application list on getAll()', async function() {
      jest.resetModules()

      const result = await applications.getAll()
      expect(result).toBeDefined()

      const { applications: apps, totalCount } = result
      expect(apps.constructor.name).toBe('Array')
      expect(apps).toHaveLength(1)
      expect(apps[0].constructor.name).toBe('Application')
      expect(totalCount).toBe(1)
    })
  })
})
