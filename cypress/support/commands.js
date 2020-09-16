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

import stringToHash from '../../pkg/webui/lib/string-to-hash'

// Helper function to quickly login to the oauth app programmatically.
Cypress.Commands.add('loginOAuth', credentials => {
  const baseUrl = Cypress.config('baseUrl')
  const oauthRootPath = Cypress.config('oauthRootPath')

  // Obtain csrf token.
  cy.request({
    method: 'GET',
    url: `${baseUrl}${oauthRootPath}/login`,
  }).then(({ headers }) => {
    cy.request({
      method: 'POST',
      url: `${baseUrl}${oauthRootPath}/api/auth/login`,
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
  const oauthRootPath = Cypress.config('oauthRootPath')
  const consoleRootPath = Cypress.config('consoleRootPath')

  // Obtain csrf token.
  cy.request({
    method: 'GET',
    url: `${baseUrl}${oauthRootPath}/login`,
  }).then(({ headers }) => {
    // Login to OAuth provider.
    cy.request({
      method: 'POST',
      url: `${baseUrl}${oauthRootPath}/api/auth/login`,
      body: { user_id: credentials.user_id, password: credentials.password },
      headers: {
        'X-CSRF-Token': headers['x-csrf-token'],
      },
    })

    // Do OAuth round trip.
    cy.request({
      method: 'GET',
      url: `${baseUrl}${consoleRootPath}/login/ttn-stack?next=/`,
    })

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

// Helper function to create a new user programmatically.
Cypress.Commands.add('createUser', user => {
  const baseUrl = Cypress.config('baseUrl')
  const oauthRootPath = Cypress.config('oauthRootPath')

  // Obtain csrf token.
  cy.request({
    method: 'GET',
    url: `${baseUrl}${oauthRootPath}/login`,
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

  // Reset cookies and local storage to avoid csrf and session state inconsitencies within tests.
  cy.clearCookies()
  cy.clearLocalStorage()
})

// Helper function to quickly seed the database to a fresh state using a
// previously generated sql dump.
Cypress.Commands.add('dropAndSeedDatabase', () => {
  cy.exec('tools/bin/mage dev:sqlRestore dev:redisFlush')
    .its('code')
    .should('eq', 0)
})

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

// Selectors.

const getFieldDescriptorByLabel = label => {
  cy.findByLabelText(label).as('field')
  return cy
    .get('@field')
    .invoke('attr', 'aria-describedby')
    .then(describedBy => {
      return cy.get(`[id=${describedBy}]`)
    })
}

// Helper function to select field error.
Cypress.Commands.add('findErrorByLabelText', label => {
  getFieldDescriptorByLabel(label).as('error')

  // Check for the error icon.
  cy.get('@error')
    .children()
    .first()
    .should('contain', 'error')
    .and('be.visible')

  return cy.get('@error')
})

// Helper function to select field warning.
Cypress.Commands.add('findWarningByLabelText', label => {
  getFieldDescriptorByLabel(label).as('warning')

  // Check for the warning icon.
  cy.get('@warning')
    .children()
    .first()
    .should('contain', 'warning')
    .and('be.visible')

  return cy.get('@warning')
})

// Helper function to select field description.
Cypress.Commands.add('findDescriptionByLabelText', label => {
  return getFieldDescriptorByLabel(label)
})
