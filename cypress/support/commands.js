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

import 'cypress-file-upload'

import { noop, merge } from 'lodash'

import stringToHash from '../../pkg/webui/lib/string-to-hash'

before(() => {
  cy.readFile('.env/admin_api_key.txt').then(adminKey => {
    Cypress.config('adminApiKey', adminKey)
  })
})

// Helper function to quickly login to the Account App app programmatically.
Cypress.Commands.add('loginAccountApp', credentials => {
  const baseUrl = Cypress.config('baseUrl')
  const accountAppRootPath = Cypress.config('accountAppRootPath')

  // Obtain csrf token.
  cy.request({
    method: 'GET',
    url: `${baseUrl}${accountAppRootPath}/login`,
  }).then(({ headers }) => {
    cy.request({
      method: 'POST',
      url: `${baseUrl}${accountAppRootPath}/api/auth/login`,
      body: { user_id: credentials.user_id, password: credentials.password },
      headers: {
        'X-CSRF-Token': headers['x-csrf-token'],
      },
    })
  })
})

// Helper function to quickly login to the console programmatically.
Cypress.Commands.add('loginConsole', credentials => {
  const baseUrl = Cypress.config('baseUrl')
  const accountAppRootPath = Cypress.config('accountAppRootPath')
  const consoleRootPath = Cypress.config('consoleRootPath')

  // Obtain csrf token.
  cy.request({
    method: 'GET',
    url: `${baseUrl}${accountAppRootPath}/login`,
  }).then(({ headers }) => {
    // Login to Account App provider.
    cy.request({
      method: 'POST',
      url: `${baseUrl}${accountAppRootPath}/api/auth/login`,
      body: { user_id: credentials.user_id, password: credentials.password },
      headers: {
        'X-CSRF-Token': headers['x-csrf-token'],
      },
    }).then(() => {
      // Do Account App round trip.
      cy.request({
        method: 'GET',
        url: `${baseUrl}${consoleRootPath}/login/ttn-stack?next=/`,
      }).then(() => {
        // Obtain access token.
        cy.request({
          method: 'GET',
          url: `${baseUrl}${consoleRootPath}/api/auth/token`,
        }).then(resp => {
          window.localStorage.setItem(
            // We store local storage values with a hashed key based on the mount path
            // to prevent clashes with other apps on the same domain.
            `accessToken-${stringToHash('/console')}`,
            JSON.stringify(resp.body),
          )
        })
      })
    })
  })
})

// Helper function to visit the authorization URI for the Console.
Cypress.Commands.add('visitConsoleAuthorizationScreen', credentials => {
  const baseUrl = Cypress.config('baseUrl')
  const accountAppRootPath = Cypress.config('accountAppRootPath')
  const consoleRootPath = Cypress.config('consoleRootPath')

  cy.task('execSql', `UPDATE clients SET skip_authorization=false WHERE client_id='console';`)
  // Obtain csrf token.
  cy.request({
    method: 'GET',
    url: `${baseUrl}${accountAppRootPath}/login`,
  }).then(({ headers }) => {
    // Login to Account App provider.
    cy.request({
      method: 'POST',
      url: `${baseUrl}${accountAppRootPath}/api/auth/login`,
      body: { user_id: credentials.user_id, password: credentials.password },
      headers: {
        'X-CSRF-Token': headers['x-csrf-token'],
      },
    }).then(() => {
      // Do Account App round trip.
      cy.request({
        method: 'GET',
        url: `${baseUrl}${consoleRootPath}/login/ttn-stack?next=/`,
      }).then(({ allRequestResponses }) => {
        cy.visit(allRequestResponses[allRequestResponses.length - 1]['Request URL'])
      })
    })
  })
})

// Helper function to obtain the currently active access token.
Cypress.Commands.add('getAccessToken', callback => {
  const tokenString = window.localStorage.getItem(`accessToken-${stringToHash('/console')}`)
  const accessToken = JSON.parse(tokenString).access_token
  callback(accessToken)
})

// Helper function to create a new user programmatically.
Cypress.Commands.add('createUser', user => {
  const baseUrl = Cypress.config('baseUrl')
  const accountAppRootPath = Cypress.config('accountAppRootPath')

  // Obtain csrf token.
  cy.request({
    method: 'GET',
    url: `${baseUrl}${accountAppRootPath}/login`,
  }).then(({ headers }) => {
    // Register user.
    cy.request({
      method: 'POST',
      url: `${baseUrl}/api/v3/users`,
      body: { user },
      headers: {
        'X-CSRF-Token': headers['x-csrf-token'],
      },
    })
  })

  // Reset cookies and local storage to avoid csrf and session state inconsistencies within tests.
  cy.clearCookies()
  cy.clearLocalStorage()
})

// Helper function to create a new client programmatically.
Cypress.Commands.add('createClient', (client, userId) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/users/${userId}/clients`,
    body: { client },
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to create a new application programmatically.
Cypress.Commands.add('createApplication', (application, userId) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/users/${userId}/applications`,
    body: { application },
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to create a new api key programmatically
Cypress.Commands.add('createApiKey', (entity, entityId, apiKey, cb = noop) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/${entity}/${entityId}/api-keys`,
    body: { ...apiKey },
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  }).then(response => cb(response.body))
})

// Helper function to create a new collaborator programmatically
Cypress.Commands.add('createCollaborator', (entity, entityId, collaborator) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'PUT',
    url: `${baseUrl}/api/v3/${entity}/${entityId}/collaborators`,
    body: collaborator,
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to set an application collaborator programmatically.
Cypress.Commands.add('setApplicationCollaborator', (applicationId, collaboratorId, rights) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  const body = {
    collaborator: {
      ids: { user_ids: { user_id: collaboratorId } },
      rights,
    },
  }
  cy.request({
    method: 'PUT',
    url: `${baseUrl}/api/v3/applications/${applicationId}/collaborators`,
    body,
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to create a new application payload formatter programmatically.
Cypress.Commands.add('setApplicationPayloadFormatter', (appId, formatter) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    url: `${baseUrl}/api/v3/as/applications/${appId}/link`,
    method: 'PUT',
    body: {
      link: {
        default_formatters: {
          down_formatter: 'FORMATTER_JAVASCRIPT',
          down_formatter_parameter:
            formatter ||
            'function encodeDownlink(input) {\n  return {\n    bytes: [],\n    fPort: 1,\n    warnings: [],\n    errors: []\n  };\n}\n\nfunction decodeDownlink(input) {\n  return {\n    data: {\n      bytes: input.bytes\n    },\n    warnings: [],\n    errors: []\n  }\n}',
          up_formatter: 'FORMATTER_JAVASCRIPT',
          up_formatter_parameter:
            formatter ||
            'function decodeUplink(input) {\n  return {\n    data: {\n      bytes: input.bytes\n    },\n    warnings: [],\n    errors: []\n  };\n}',
        },
      },
      field_mask: {
        paths: [
          'default_formatters.down_formatter',
          'default_formatters.down_formatter_parameter',
          'default_formatters.up_formatter',
          'default_formatters.up_formatter_parameter',
        ],
      },
    },
    headers: { Authorization: `Bearer ${adminApiKey}` },
  })
})

// Helper function to create a new gateway programmatically.
Cypress.Commands.add('createGateway', (gateway, userId) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/users/${userId}/gateways`,
    body: { gateway },
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to create a new organization programmatically.
Cypress.Commands.add('createOrganization', (organization, userId) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/users/${userId}/organizations`,
    body: { organization },
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to create a new end device programmatically.
Cypress.Commands.add('createEndDeviceIsOnly', (applicationId, endDevice) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/applications/${applicationId}/devices`,
    body: endDevice,
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to create a mock device in all components.
Cypress.Commands.add(
  'createMockDeviceAllComponents',
  (
    applicationId,
    fixture = 'console/devices/device.*.json',
    overwrites = { ns: {}, as: {}, js: {}, is: {} },
    injectHost = true,
  ) => {
    const baseUrl = Cypress.config('baseUrl')
    const adminApiKey = Cypress.config('adminApiKey')
    const interpolateFixture = (fixtureString, component) => fixtureString.replace('*', component)
    const headers = {
      Authorization: `Bearer ${adminApiKey}`,
    }
    cy.fixture(interpolateFixture(fixture, 'is')).then(body => {
      if (injectHost && body && 'end_device' in body) {
        if ('network_server_address' in body.end_device) {
          body.end_device.network_server_address = window.location.hostname
        }
        if ('join_server_address' in body.end_device) {
          body.end_device.join_server_address = window.location.hostname
        }
        if ('application_server_address' in body.end_device) {
          body.end_device.application_server_address = window.location.hostname
        }
      }
      cy.request({
        method: 'POST',
        url: `${baseUrl}/api/v3/applications/${applicationId}/devices`,
        body: { ...body, ...overwrites.is },
        headers,
      })
    })
    cy.fixture(interpolateFixture(fixture, 'ns')).then(body => {
      cy.request({
        method: 'PUT',
        url: `${baseUrl}/api/v3/ns/applications/${applicationId}/devices/${body.end_device.ids.device_id}`,
        body: { ...body, ...overwrites.ns },
        headers,
      })
    })
    cy.fixture(interpolateFixture(fixture, 'as')).then(body => {
      cy.request({
        method: 'PUT',
        url: `${baseUrl}/api/v3/as/applications/${applicationId}/devices/${body.end_device.ids.device_id}`,
        body: { ...body, ...overwrites.as },
        headers,
      })
    })
    cy.fixture(interpolateFixture(fixture, 'js')).then(body => {
      if (injectHost && body && 'end_device' in body) {
        if ('network_server_address' in body.end_device) {
          body.end_device.network_server_address = window.location.hostname
        }
        if ('application_server_address' in body.end_device) {
          body.end_device.application_server_address = window.location.hostname
        }
      }
      cy.request({
        method: 'PUT',
        url: `${baseUrl}/api/v3/js/applications/${applicationId}/devices/${body.end_device.ids.device_id}`,
        body: { ...body, ...overwrites.js },
        headers,
      })
    })
    return cy.fixture(interpolateFixture(fixture, 'is'))
  },
)

// Helper function to create a new pub sub programmatically.
Cypress.Commands.add('createPubSub', (applicationId, pubSub) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/as/pubsub/${applicationId}`,
    body: pubSub,
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to create a new pub sub programmatically.
Cypress.Commands.add('createWebhook', (applicationId, webhook) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'POST',
    url: `${baseUrl}/api/v3/as/webhooks/${applicationId}`,
    body: webhook,
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Helper function to update gateway programmatically.
Cypress.Commands.add('updateGateway', (gatewayId, gateway) => {
  const baseUrl = Cypress.config('baseUrl')
  const adminApiKey = Cypress.config('adminApiKey')
  cy.request({
    method: 'PUT',
    url: `${baseUrl}/api/v3/gateways/${gatewayId}`,
    body: gateway,
    headers: {
      Authorization: `Bearer ${adminApiKey}`,
    },
  })
})

// Overwrite the default `type` to make sure that subject is resolved and focused before simulating typing. This is helpful
// when:
// 1. The action is forced via the `forced` option for inputs that are visually hidden for styling purposes.
// 2. The action is performed during minor layout shifts.
Cypress.Commands.overwrite('type', (originalFn, subject, ...args) => {
  subject.focus()

  return originalFn(subject, ...args)
})

// Overwrite the default `click` to make sure that subject is resolved and focused before simulating clicks. This is helpful
// when:
// 1. The action is forced via the `forced` option for elements that are visually hidden for styling purposes.
// 2. The action is performed during minor layout shifts.
Cypress.Commands.overwrite('click', (originalFn, subject, ...args) => {
  subject.focus()

  return originalFn(subject, ...args)
})

// Helper function to quickly seed the database to a fresh state using a
// previously generated sql dump.
Cypress.Commands.add('dropAndSeedDatabase', () => cy.task('dropAndSeedDatabase'))

// Helper function to augment the stack configuration object. See support/utils.js for utility
// functions to modify the configuration object.
// Note: make sure to call this function before the first call to `cy.visit`, otherwise there will
// be no effect and the configuration object will be left unchanged. For more details see
// https://docs.cypress.io/api/events/catalog-of-events.html#App-Events.
Cypress.Commands.add('augmentStackConfig', fns => {
  const fnArray = Array.isArray(fns) ? fns : [fns]

  cy.on('window:before:load', win => {
    Object.defineProperty(win, '__initStackConfig', {
      value: () => {
        fnArray.forEach(fn => fn(win.__ttn_config__))
      },
    })
  })
})

// Helper function to augment the configuration that is returned from the IS.
// Note: `cy.server()` needs to run before this command.
Cypress.Commands.add('augmentIsConfig', config => {
  const baseUrl = Cypress.config('baseUrl')
  cy.request({
    method: 'GET',
    url: `${baseUrl}/api/v3/is/configuration`,
  }).then(({ body }) => {
    cy.intercept('GET', `${baseUrl}/api/v3/is/configuration`, {
      configuration: merge({}, body.configuration, config),
    })
  })
})

// Selectors.

// Helper function to select an option. Use this function instead of `cy.select` as it allows
// interacting with `react-select` in both interactive and headless modes of cypress.
Cypress.Commands.add('selectOption', { prevSubject: true }, (subject, option) => {
  cy.wrap(subject).type(option, { force: true })

  cy.get('.select__option')
    .first()
    .then($option => {
      // Native `cy.click` even with the `force` option doesnt work properly in headless electron
      // environment causing issues when dealing with `react-select` (in interactive mode this works fine).
      Cypress.$($option).trigger('click')
    })
})

const getFieldDescriptorByLabel = label => {
  cy.findByLabelText(label).as('field')
  return cy
    .get('@field')
    .invoke('attr', 'aria-describedby')
    .then(describedBy => cy.get(`[id="${describedBy}"]`))
}

// Helper function to select field error.
Cypress.Commands.add('findErrorByLabelText', label => {
  getFieldDescriptorByLabel(label).as('error')

  // Check for the error icon.
  cy.get('@error').children().first().should('contain', 'error').and('be.visible')

  return cy.get('@error')
})

// Helper function to select field warning.
Cypress.Commands.add('findWarningByLabelText', label => {
  getFieldDescriptorByLabel(label).as('warning')

  // Check for the warning icon.
  cy.get('@warning').children().first().should('contain', 'warning').and('be.visible')

  return cy.get('@warning')
})

// Helper function to select field description.
Cypress.Commands.add('findDescriptionByLabelText', label => getFieldDescriptorByLabel(label))
