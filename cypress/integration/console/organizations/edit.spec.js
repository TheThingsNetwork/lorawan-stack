// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

describe('Organization general settings', () => {
  const organizationId = 'test-organization'
  const organization = { ids: { organization_id: organizationId } }
  const userId = 'main-organization-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-organization-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.loginConsole({ user_id: userId, password: user.password })
    cy.createOrganization(organization, userId)
    cy.clearLocalStorage()
    cy.clearCookies()
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('successfully edit organization name and description', () => {
    cy.visit(
      `${Cypress.config('consoleRootPath')}/organizations/${organizationId}/general-settings`,
    )

    cy.findByLabelText('Name').type('test-name')
    cy.findByLabelText('Description').type('test-description')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Organization updated`)
      .should('be.visible')
  })

  it('successfully delete organization', () => {
    cy.visit(
      `${Cypress.config('consoleRootPath')}/organizations/${organizationId}/general-settings`,
    )
    cy.findByRole('button', { name: /Delete organization/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Delete organization', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Delete organization/ }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/organizations`)

    cy.findByRole('cell', { name: organizationId }).should('not.exist')
  })
})
