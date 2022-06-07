// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

describe('Send invite', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: 'admin', password: 'admin' })
  })

  it('displays UI elements in place', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin/user-management/invitations/add`)

    cy.findByText('Invite', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Email address')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'mail@example.com')
    cy.findByRole('button', { name: 'Invite' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin/user-management/invitations/add`)

    cy.findByRole('button', { name: 'Invite' }).should('be.visible').click()

    cy.findErrorByLabelText('Email address')
      .should('contain.text', 'Email address is required')
      .and('be.visible')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/admin/user-management/invitations/add`,
    )
  })

  it('succeeds inviting a user', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin/user-management/invitations/add`)
    cy.findByLabelText('Email address').type('mail@example.com')

    cy.findByRole('button', { name: 'Invite' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/admin/user-management`,
    )
    cy.findByText('Invitations').click()
    cy.findByRole('cell', { email: 'mail@example.com' }).should('not.exist')
  })
})
