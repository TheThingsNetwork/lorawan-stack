// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

describe('OAuth Client general settings', () => {
  const clientId = 'test-client'
  const client = {
    ids: { client_id: clientId },
    grants: ['GRANT_AUTHORIZATION_CODE'],
  }
  const userId = '1-oauth-client-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: '1-oauth-client-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createClient(client, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('succeeds editing client', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/oauth-clients/${clientId}`)

    cy.findByLabelText('Name').type('test-name')
    cy.findByLabelText('Description').type('test-description')
    cy.findByRole('button', { name: /Add redirect URL/ }).click()
    cy.get(`[name="redirect_uris[0].value"]`).type('client-test-url')
    cy.findByRole('button', { name: /Add logout redirect URL/ }).click()
    cy.get(`[name="logout_redirect_uris[0].value"]`).type('client-test-url')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .findByText(`OAuth client updated`)
      .should('be.visible')

    cy.reload()

    cy.findByLabelText('Name').should('be.visible').and('have.attr', 'value').and('eq', 'test-name')
    cy.findByLabelText('Description').should('be.visible').should('have.text', 'test-description')
    cy.get(`[name="redirect_uris[0].value"]`)
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', 'client-test-url')
    cy.get(`[name="logout_redirect_uris[0].value"]`)
      .should('be.visible')
      .and('have.attr', 'value')
      .and('eq', 'client-test-url')
  })

  it('succeeds deleting client', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/oauth-clients/${clientId}`)
    cy.findByRole('button', { name: /Delete OAuth client/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Delete OAuth client/ }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/user-settings/oauth-clients`,
    )

    cy.findByRole('cell', { name: clientId }).should('not.exist')
  })
})
