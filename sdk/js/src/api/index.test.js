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

import StackConfiguration from '../util/stack-configuration'
import Api from '.'

jest.mock('../../generated/api-definition.json', () => ({
  ApplicationRegistry: {
    Create: {
      file: 'lorawan-stack/api/application_services.proto',
      http: [
        {
          method: 'post',
          pattern: '/users/{collaborator.user_ids.user_id}/applications',
          body: '*',
          parameters: ['collaborator.user_ids.user_id'],
        },
        {
          method: 'post',
          pattern: '/organizations/{collaborator.organization_ids.organization_id}/applications',
          body: '*',
          parameters: ['collaborator.organization_ids.organization_id'],
        },
      ],
    },
    List: {
      file: 'lorawan-stack/api/application_services.proto',
      http: [
        {
          method: 'get',
          pattern: '/applications',
          parameters: [],
        },
      ],
    },
    Events: {
      file: 'lorawan-stack/api/application_services.proto',
      http: [
        {
          method: 'get',
          pattern: '/events',
          parameters: [],
          stream: true,
        },
      ],
    },
  },
}))

describe('API', function() {
  let api
  beforeEach(function() {
    api = new Api(
      'http',
      new StackConfiguration({
        is: 'http://localhost:1885',
        as: 'http://localhost:1885',
        ns: 'http://localhost:1885',
        js: 'http://localhost:1885',
      }),
    )
    api._connector.handleRequest = jest.fn()
  })

  test('it applies api definitions correctly', function() {
    expect(api.ApplicationRegistry.Create).toBeDefined()
    expect(typeof api.ApplicationRegistry.Create).toBe('function')
  })

  test('it applies parameters correctly', function() {
    api.ApplicationRegistry.Create(
      { routeParams: { 'collaborator.user_ids.user_id': 'test' } },
      { name: 'test-name' },
    )

    expect(api._connector.handleRequest).toHaveBeenCalledTimes(1)
    expect(api._connector.handleRequest).toHaveBeenCalledWith(
      'post',
      '/users/test/applications',
      undefined,
      { name: 'test-name' },
      false,
    )
  })

  test('it throws when parameters mismatch', function() {
    expect(function() {
      api.ApplicationRegistry.Create({ 'some.other.param': 'test' })
    }).toThrow()
  })

  test('it respects the search query', function() {
    api.ApplicationRegistry.List(undefined, { limit: 2, page: 1 })

    expect(api._connector.handleRequest).toHaveBeenCalledTimes(1)
    expect(api._connector.handleRequest).toHaveBeenCalledWith(
      'get',
      '/applications',
      undefined,
      { limit: 2, page: 1 },
      false,
    )
  })

  test('it sets stream value to true for streaming endpoints', function() {
    api.ApplicationRegistry.Events()

    expect(api._connector.handleRequest).toHaveBeenCalledTimes(1)
    expect(api._connector.handleRequest).toHaveBeenCalledWith(
      'get',
      '/events',
      undefined,
      undefined,
      true,
    )
  })
})
