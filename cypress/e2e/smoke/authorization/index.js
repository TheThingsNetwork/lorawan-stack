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

import { defineSmokeTest } from '../utils'

const authorizeConsole = defineSmokeTest(
  'succeeds manually authorizing Account App clients',
  () => {
    const user = {
      ids: { user_id: 'authorize-test-user' },
      primary_email_address: 'create-app-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }

    cy.createUser(user)

    // Set Console client's `skip_authorization` flag to false.
    cy.task('execSql', `UPDATE clients SET skip_authorization=false WHERE client_id='console';`)
    cy.visit(Cypress.config('consoleRootPath'))

    // Login.
    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(user.password)
    cy.findByRole('button', { name: 'Login' }).click()

    // Authorize.
    cy.location('pathname').should('contain', `${Cypress.config('accountAppRootPath')}/authorize`)
    cy.findByRole('button', { name: /Authorize Console/ }).click()

    // Check Console landing view.
    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/`)
    cy.findByTestId('full-error-view').should('not.exist')
  },
)

const abortAuthorization = defineSmokeTest(
  'succeeds manually aborting Account App client authorization',
  () => {
    const user = {
      ids: { user_id: 'authorize-test-user' },
      primary_email_address: 'create-app-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }

    cy.createUser(user)

    // Set Console client's `skip_authorization` flag to false.
    cy.task('execSql', `UPDATE clients SET skip_authorization=false WHERE client_id='console';`)
    cy.visit(Cypress.config('consoleRootPath'))

    // Login.
    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(user.password)
    cy.findByRole('button', { name: 'Login' }).click()

    // Deny authorization.
    cy.location('pathname').should('contain', `${Cypress.config('accountAppRootPath')}/authorize`)
    cy.findByRole('button', { name: /Cancel/ }).click()

    // Check Console error.
    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/oauth/callback`)
    cy.findByTestId('full-error-view').should('exist')
    cy.findByText(/Login failed/).should('be.visible')
  },
)

export default [authorizeConsole, abortAuthorization]
