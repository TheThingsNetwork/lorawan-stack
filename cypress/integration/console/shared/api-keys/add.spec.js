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
  })

  describe('Application', () => {
    const applicationId = 'api-keys-test-app'
    const application = { ids: { application_id: applicationId } }

    before(() => {
      cy.createApplication(application, userId)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/api-keys/add`)
    })

    it('succeeds adding new api key', () => {
      cy.findByLabelText('Name').type(apiKeyName)
      cy.findByLabelText('Grant all current and future rights').check()
      cy.findByLabelText('Expiry date').type('2056-01-01')
      cy.findByRole('button', { name: 'Create API key' }).click()

      cy.findByTestId('error-notification').should('not.exist')

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Please copy newly created API key', { selector: 'h1' }).should(
            'be.visible',
          )
          cy.findByRole('button', { name: /I have copied the key/ }).click()
        })

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${applicationId}/api-keys`,
      )
    })
  })

  describe('Gateway', () => {
    const gatewayId = 'api-keys-test-gateway'
    const gateway = { ids: { gateway_id: gatewayId } }

    before(() => {
      cy.createGateway(gateway, userId)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/api-keys/add`)
    })

    it('succeeds adding new api key', () => {
      cy.findByLabelText('Name').type(apiKeyName)
      cy.findByLabelText('Grant all current and future rights').check()
      cy.findByLabelText('Expiry date').type('2056-01-01')
      cy.findByRole('button', { name: 'Create API key' }).click()

      cy.findByTestId('error-notification').should('not.exist')

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Please copy newly created API key', { selector: 'h1' }).should(
            'be.visible',
          )
          cy.findByRole('button', { name: /I have copied the key/ }).click()
        })

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/api-keys`,
      )
    })
  })

  describe('Organization', () => {
    const organizationId = 'api-keys-test-org'
    const organization = {
      ids: { organization_id: organizationId },
    }

    before(() => {
      cy.createOrganization(organization, userId)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/organizations/${organizationId}/api-keys/add`)
    })

    it('succeeds adding new api key', () => {
      cy.findByLabelText('Name').type(apiKeyName)
      cy.findByLabelText('Grant all current and future rights').check()
      cy.findByLabelText('Expiry date').type('2056-01-01')
      cy.findByRole('button', { name: 'Create API key' }).click()

      cy.findByTestId('error-notification').should('not.exist')

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Please copy newly created API key', { selector: 'h1' }).should(
            'be.visible',
          )
          cy.findByRole('button', { name: /I have copied the key/ }).click()
        })

      cy.location('pathname').should(
        'eq',
        `${Cypress.config('consoleRootPath')}/organizations/${organizationId}/api-keys`,
      )
    })
  })

  describe('User', () => {
    beforeEach(() => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/user/api-keys/add`)
    })

    it('succeeds adding new api key', () => {
      cy.findByLabelText('Name').type(apiKeyName)
      cy.findByLabelText('Grant all current and future rights').check()
      cy.findByLabelText('Expiry date').type('2056-01-01')
      cy.findByRole('button', { name: 'Create API key' }).click()

      cy.findByTestId('error-notification').should('not.exist')

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Please copy newly created API key', { selector: 'h1' }).should(
            'be.visible',
          )
          cy.findByRole('button', { name: /I have copied the key/ }).click()
        })

      cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/user/api-keys`)
    })
  })
})
