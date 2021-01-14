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

describe('Account App authorization view', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays UI elements in place', () => {
    const user = {
      ids: { user_id: 'create-app-test-user' },
      primary_email_address: 'create-app-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }

    cy.createUser(user)
    cy.visitConsoleAuthorizationScreen({ user_id: user.ids.user_id, password: user.password })

    // Check authorization screen.
    cy.location('pathname').should('contain', `${Cypress.config('accountAppRootPath')}/authorize`)
    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByText('Request for permission').should('be.visible')
    cy.findByText('Console').should('be.visible')
    cy.findByText('All possible rights', { exact: false }).should('be.visible')
    cy.findByText('all possible current and future rights').should('be.visible')
    cy.findByText(`You are logged in as ${user.ids.user_id}.`).should('be.visible')
    cy.findByText(
      `You will be redirected to ${Cypress.config('consoleRootPath')}/oauth/callback`,
    ).should('be.visible')
    cy.findByRole('button', { name: /Cancel/ }).should('be.visible')
    cy.findByRole('button', { name: /Authorize Console/ }).should('be.visible')
  })
})
