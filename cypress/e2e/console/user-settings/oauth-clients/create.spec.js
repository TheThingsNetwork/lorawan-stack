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

describe('OAuth Client create', () => {
  const clientId = 'test-client'
  const user = {
    ids: { user_id: 'create-client-test-user' },
    primary_email_address: 'create-client-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/oauth-clients/add`)
  })

  it('displays UI elements in place', () => {
    cy.findByText('Add OAuth client', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('OAuth client ID')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'my-new-oauth-client')
    cy.findByLabelText('Name')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'My new OAuth client')
    cy.findByLabelText('Description')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'Description for my new OAuth client')
    cy.findDescriptionByLabelText('Description')
      .should(
        'contain',
        'The description is displayed to the user when authorizing the client. Use it to explain the purpose of your client.',
      )
      .and('be.visible')
    cy.findByRole('button', { name: 'Create OAuth client' }).should('be.exist')
  })

  it('validates before submitting an empty form', () => {
    cy.findByRole('button', { name: 'Create OAuth client' }).click()

    cy.findErrorByLabelText('OAuth client ID')
      .should('contain.text', 'OAuth client ID is required')
      .and('be.visible')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/user-settings/oauth-clients/add`,
    )
  })

  it('succeeds adding client', () => {
    cy.findByLabelText('OAuth client ID').type(clientId)

    cy.findByRole('button', { name: 'Create OAuth client' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/user-settings/oauth-clients`,
    )
    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: clientId }).should('be.visible')
  })
})
