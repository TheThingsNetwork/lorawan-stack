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

const user = {
  ids: { user_id: 'test-user-id1' },
  primary_email_address: 'test-user1@example.com',
  password: 'ABCDefg123!',
  password_confirm: 'ABCDefg123!',
}

const newPassword = 'ABCDefg321!'

describe('User settings / change password', () => {
  beforeEach(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/user-settings/password`)
  })

  it('displays UI elements in place', () => {
    cy.findByLabelText('Current password').should('be.visible')
    cy.findByLabelText('New password').should('be.visible')
    cy.findByLabelText('Confirm new password').should('be.visible')
    cy.findByLabelText('Revoke all access').should('exist').and('not.be.checked')
    cy.findByRole('button', { name: 'Change password' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.findByRole('button', { name: 'Change password' }).click()

    cy.findErrorByLabelText('New password')
      .should('contain.text', 'New password is required')
      .and('be.visible')
    cy.findErrorByLabelText('Confirm new password')
      .should('contain.text', 'Confirm new password is required')
      .and('be.visible')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/user-settings/password`,
    )
  })

  it('succeeds changing password when revoking access', () => {
    cy.findByLabelText('Current password').type(user.password)
    cy.findByLabelText('New password').type(newPassword)
    cy.findByLabelText('Revoke all access').check()
    cy.findByLabelText('Confirm new password').type(`${newPassword}{enter}`)

    cy.findByTestId('toast-notification').should('be.visible').and('contain', 'Password changed')
  })

  it('succeeds changing password without revoking access', () => {
    cy.findByLabelText('Current password').type(user.password)
    cy.findByLabelText('New password').type(newPassword)
    cy.findByLabelText('Confirm new password').type(newPassword)
    cy.findByRole('button', { name: 'Change password' }).click()

    cy.findByTestId('toast-notification').should('be.visible').and('contain', 'Password changed')
  })
})
