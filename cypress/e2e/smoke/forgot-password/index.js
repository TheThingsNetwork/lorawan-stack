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

// eslint-disable-next-line no-useless-escape
const updatePasswordLinkRegExp = `https?:\\/\\/[a-zA-Z0-9-_.:]+/[a-zA-Z0-9-_]+\\/update-password\\?.+&current=[A-Z0-9]+`

const forgotPasswordFlow = defineSmokeTest('forgot password flow succeeds', () => {
  const user = {
    ids: { user_id: 'test-user-id1' },
    primary_email_address: 'test-user1@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const newPassword = 'ABCDefg321!'

  cy.createUser(user)

  // Start the flow at the login screen and navigate to the forgot password view.
  cy.visit(`${Cypress.config('accountAppRootPath')}`)
  cy.findByRole('link', { name: 'Forgot password?' }).click()
  cy.location('pathname').should(
    'include',
    `${Cypress.config('accountAppRootPath')}/forgot-password`,
  )

  // Type in own username and hit `Send`.
  cy.findByLabelText('User ID').type(user.ids.user_id)
  cy.findByRole('button', { name: 'Send' }).click()
  cy.findByTestId('notification').should('be.visible').should('contain', 'reset instruction')
  cy.location('pathname').should('include', `${Cypress.config('accountAppRootPath')}/login`)

  // Retrieve password token link from email (via stack logs).
  cy.task('findInLatestEmail', updatePasswordLinkRegExp).then(temporaryPasswordLink => {
    cy.log(temporaryPasswordLink)

    // Navigate to the password token link and submit new credentials.
    cy.visit(temporaryPasswordLink)
    cy.findByLabelText('New password').type(newPassword)
    cy.findByLabelText('Confirm new password').type(newPassword)
    cy.findByRole('button', { name: 'Change password' }).click()
    cy.findByTestId('notification').should('be.visible').should('contain', 'Password changed')
    cy.location('pathname').should('include', `${Cypress.config('accountAppRootPath')}/login`)

    // Verify new credentials by logging in.
    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(newPassword)
    cy.findByRole('button', { name: 'Login' }).click()
    cy.location('pathname').should('eq', `${Cypress.config('accountAppRootPath')}/`)
    cy.findByTestId('full-error-view').should('not.exist')
  })
})

export default [forgotPasswordFlow]
