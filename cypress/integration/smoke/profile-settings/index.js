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

  cy.get('header').within(() => {
    cy.findByTestId('profile-dropdown').should('contain', user.name).as('profileDropdown')

    cy.get('@profileDropdown').click()

    cy.get('@profileDropdown')
      .findByText(/Profile settings/)
      .should('have.attr', 'href', '/oauth/profile-settings')
      .should('have.attr', 'target', '_blank')

    cy.get('@profileDropdown')
      .findByText('Profile settings')
      .then(link => {
        cy.visit(link.prop('href'))
        cy.location('pathname').should('eq', '/oauth/profile-settings')
      })
  })

  cy.findByText('General settings', { selector: 'h3' })
    .closest('[data-test-id="collapsible-section"]')
    .within(() => {
      cy.findByRole('button', { name: 'Expand' }).click()
    })

  cy.findByRole('button', { name: 'Save changes' }).click()
  cy.findByTestId('error-notification').should('not.exist')
  cy.findByTestId('toast-notification')
    .should('be.visible')
    .findByText('Profile updated')
    .should('be.visible')
})

export default [profileSettingsNavigation]
