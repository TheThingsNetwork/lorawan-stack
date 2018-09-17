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

import Api from '.'

jest.mock('../../generated/api-definition.json', () => (
  {
    ListApplications: {
      file: 'lorawan-stack/api/application_services.proto',
      http: [
        {
          method: 'get',
          pattern: '/applications',
          parameters: [],
        },
        {
          method: 'get',
          pattern: '/users/{collaborator.user_ids.user_id}/applications',
          parameters: [
            'collaborator.user_ids.user_id',
          ],
        },
        {
          method: 'get',
          pattern: '/organizations/{collaborator.organization_ids.organization_id}/applications',
          parameters: [
            'collaborator.organization_ids.organization_id',
          ],
        },
      ],
    },
  }
))

describe('API', function () {
  let api
  beforeEach(function () {
    api = new Api('http', { baseURL: 'http://localhost:1885' })
    api.connector.get = jest.fn()
  })

  test('it applies api definitions correctly', function () {
    expect(api.ListApplications).toBeDefined()
    expect(typeof api.ListApplications).toBe('function')
  })

  test('it applies parameters correctly', function () {
    api.connector.get = jest.fn()

    api.ListApplications({ 'collaborator.user_ids.user_id': 'test' })

    expect(api.connector.get).toHaveBeenCalledTimes(1)
    expect(api.connector.get).toHaveBeenCalledWith('/users/test/applications', undefined)
  })

  test('it throws when parameters mismatch', function () {
    expect(function () {
      api.ListApplications({ 'some.other.param': 'test' })
    }).toThrow()
  })
})
