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

/* eslint-disable jest/valid-expect */

import { defineSmokeTest } from '../utils'

const profileSettingsNavigation = defineSmokeTest('succeeds navigating to Account App', () => {
  const user = {
    ids: { user_id: 'test-account-app-user' },
    primary_email_address: 'test-account-app-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    name: 'Test Account App User',
  }

  cy.createUser(user)
  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))

  cy.findByRole('button', { name: /User settings/ })
    .should('be.visible')
    .click()
  cy.findByRole('link', { name: /Profile/ })
    .should('be.visible')
    .click()
  cy.findByRole('button', { name: 'Save changes' }).click()
  cy.findByTestId('error-notification').should('not.exist')
  cy.findByTestId('toast-notification-success')
    .should('be.visible')
    .findByText('Profile updated')
    .should('be.visible')

  cy.findByRole('link', { name: /Password/ })
    .should('be.visible')
    .click()
  cy.findByLabelText('Current password').type(user.password)
  cy.findByLabelText('New password').type('ABCDefg321!')
  cy.findByLabelText('Confirm new password').type('ABCDefg321!')
  cy.findByRole('button', { name: 'Change password' }).click()
  cy.findByTestId('error-notification').should('not.exist')
  cy.findByTestId('toast-notification-success')
    .should('be.visible')
    .and('contain', 'Password changed')

  cy.findByRole('link', { name: /API keys/ })
    .should('be.visible')
    .click()
  cy.findByRole('link', { name: /Add API key/ }).click()
  cy.findByLabelText('Name').type('Test API key')
  cy.findByRole('button', { name: 'Create API key' }).click()
  cy.findByRole('button', { name: /I have copied the key/ }).click()
  cy.findByTestId('error-notification').should('not.exist')
  cy.findByRole('cell', { name: 'Test API key' }).should('be.visible')

  cy.findByRole('link', { name: /Session management/ })
    .should('be.visible')
    .click()
  cy.findByText('Sessions (1)').should('be.visible')
  cy.findByRole('rowgroup').within(() => {
    cy.findByRole('row').should('have.length', 1)
  })
  cy.findByRole('link', { name: /Authorizations/ })
    .should('be.visible')
    .click()
  cy.findByText('OAuth client authorizations (1)').should('be.visible')
  cy.findByRole('rowgroup').within(() => {
    cy.findByRole('row').should('have.length', 1)
  })
  cy.findByRole('link', { name: /OAuth clients/ })
    .should('be.visible')
    .click()
  cy.findByRole('link', { name: /Add OAuth client/ }).click()
  cy.findByLabelText('OAuth client ID').type('test-client')
  cy.findByLabelText('Name').type('Test OAuth client')
  cy.findByRole('button', { name: 'Create OAuth client' }).click()
  cy.findByTestId('error-notification').should('not.exist')
  cy.findByRole('cell', { name: 'Test OAuth client' }).should('be.visible')
})

export default [profileSettingsNavigation]
