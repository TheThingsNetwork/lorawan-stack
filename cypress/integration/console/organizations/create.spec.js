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

describe('Organization create', () => {
  const organizationId = 'test-organization'
  const userId = 'create-organization-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-organization-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/organizations/add`)
  })

  it('displays UI elements in place', () => {
    cy.findByText('Create organization', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Organization ID').should('be.visible')
    cy.findByLabelText('Name').should('be.visible')
    cy.findByLabelText('Description').should('be.visible')
    cy.findDescriptionByLabelText('Description')
      .should(
        'contain',
        'Optional organization description; can also be used to save notes about the organization',
      )
      .and('be.visible')
    cy.findByRole('button', { name: 'Create organization' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.findByRole('button', { name: 'Create organization' }).should('be.visible').click()

    cy.findErrorByLabelText('Organization ID')
      .should('contain.text', 'Organization ID is required')
      .and('be.visible')
    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/organizations/add`)
  })

  it('succeeds adding organization', () => {
    cy.findByLabelText('Organization ID').type(organizationId)

    cy.findByRole('button', { name: 'Create organization' }).should('be.visible').click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/organizations/${organizationId}`,
    )
  })
})
