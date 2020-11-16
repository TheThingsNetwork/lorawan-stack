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

describe('API keys', () => {
  const apiKeyName = 'api-test-key'
  const apiKey = {
    name: apiKeyName,
    rights: ['RIGHT_APPLICATION_ALL'],
  }
  const userId = 'main-api-key-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'create-api-key-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()

    cy.createUser(user)
    cy.loginConsole({
      user_id: userId,
      password: user.password,
    })

    cy.clearLocalStorage()
    cy.clearCookies()
  })

  describe('Application', () => {
    const applicationId = 'api-keys-test-app'
    const application = { ids: { application_id: applicationId } }

    before(() => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.createApplication(application, userId)
      cy.createApiKey('applications', applicationId, apiKey)
      cy.clearLocalStorage()
      cy.clearCookies()
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('succeeds editing api key', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/api-keys`)

      cy.findByRole('cell', { name: apiKeyName }).click()
      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`API key updated`)
        .should('be.visible')
    })

    it('succeeds deleting api key', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/api-keys`)

      cy.findByRole('cell', { name: apiKeyName }).click()
      cy.findByRole('button', { name: /Delete key/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Delete key', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Delete key/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${applicationId}/api-keys`,
      )

      cy.findByRole('cell', { name: apiKeyName }).should('not.exist')
    })
  })

  describe('Gateway', () => {
    const gatewayId = 'api-keys-test-gateway'
    const gateway = { ids: { gateway_id: gatewayId } }

    before(() => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.createGateway(gateway, userId)
      cy.createApiKey('gateways', gatewayId, apiKey)
      cy.clearLocalStorage()
      cy.clearCookies()
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('succeeds editing api key', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/api-keys`)

      cy.findByRole('cell', { name: apiKeyName }).click()
      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`API key updated`)
        .should('be.visible')
    })

    it('succeeds deleting api key', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/api-keys`)

      cy.findByRole('cell', { name: apiKeyName }).click()
      cy.findByRole('button', { name: /Delete key/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Delete key', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Delete key/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/api-keys`,
      )

      cy.findByRole('cell', { name: apiKeyName }).should('not.exist')
    })
  })

  describe('Organization', () => {
    const organizationId = 'api-keys-test-org'
    const organization = {
      ids: { organization_id: organizationId },
    }

    before(() => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.createOrganization(organization, userId)
      cy.createApiKey('organizations', organizationId, apiKey)
      cy.clearLocalStorage()
      cy.clearCookies()
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    })

    it('succeeds editing api key', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/organizations/${organizationId}/api-keys`)

      cy.findByRole('cell', { name: apiKeyName }).click()
      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`API key updated`)
        .should('be.visible')
    })

    it('succeeds deleting api key', () => {
      cy.visit(`${Cypress.config('consoleRootPath')}/organizations/${organizationId}/api-keys`)

      cy.findByRole('cell', { name: apiKeyName }).click()
      cy.findByRole('button', { name: /Delete key/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Delete key', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Delete key/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/organizations/${organizationId}/api-keys`,
      )

      cy.findByRole('cell', { name: apiKeyName }).should('not.exist')
    })
  })
})
