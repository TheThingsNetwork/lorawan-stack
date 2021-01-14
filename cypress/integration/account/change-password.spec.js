// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

describe('OAuth change password', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays UI elements in place', () => {
    const user = { user_id: 'admin', password: 'admin' }
    cy.loginOAuth(user)
    cy.visit(`${Cypress.config('oauthRootPath')}/update-password`)

    cy.findByText('Change password', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Old password').should('be.visible')
    cy.findByLabelText('New password').should('be.visible')
    cy.findByLabelText('Confirm password').should('be.visible')
    cy.findByLabelText('Revoke access')
      .should('be.visible')
      .should('have.attr', 'value', 'true')
    cy.findWarningByLabelText('Revoke access')
      .should('contain', 'This will revoke access from all logged in devices')
      .and('be.visible')
    cy.findByRole('button', { name: 'Change password' }).should('be.visible')
    cy.findByRole('button', { name: 'Cancel' }).should('be.visible')
  }).skip()

  it('validates before submitting an empty form', () => {
    const user = {
      ids: { user_id: 'test-user-id1' },
      primary_email_address: 'test-user1@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.loginOAuth({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('oauthRootPath')}/update-password`)
    cy.findByRole('button', { name: 'Change password' }).click()

    cy.findErrorByLabelText('New password')
      .should('contain.text', 'New password is required')
      .and('be.visible')
    cy.findErrorByLabelText('Confirm password')
      .should('contain.text', 'Confirm password is required')
      .and('be.visible')

    cy.location('pathname').should('eq', `${Cypress.config('oauthRootPath')}/update-password`)
  }).skip()

  it('succeeds changing password when revoking access', () => {
    const newPassword = 'ABCDefg321!'
    const user = {
      ids: { user_id: 'test-user-id2' },
      primary_email_address: 'test-user2@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.loginOAuth({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('oauthRootPath')}/update-password`)

    cy.findByLabelText('Old password').type(user.password)
    cy.findByLabelText('New password').type(newPassword)
    cy.findByLabelText('Confirm password').type(`${newPassword}{enter}`)

    cy.findByTestId('notification')
      .should('be.visible')
      .should('contain', 'Password changed')
    cy.location('pathname').should('include', `${Cypress.config('oauthRootPath')}/login`)
  }).skip()

  it('succeeds changing password without revoking access', () => {
    const newPassword = 'ABCDefg321!'
    const user = {
      ids: { user_id: 'test-user-id3' },
      primary_email_address: 'test-user3@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.loginOAuth({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('oauthRootPath')}/update-password`)

    cy.findByLabelText('Old password').type(user.password)
    cy.findByLabelText('New password').type(newPassword)
    cy.findByLabelText('Confirm password').type(newPassword)
    cy.findByLabelText('Revoke access').uncheck()
    cy.findByRole('button', { name: 'Change password' }).click()

    cy.findByTestId('notification')
      .should('be.visible')
      .should('contain', 'Password changed')
    cy.location('pathname').should('include', `${Cypress.config('oauthRootPath')}/login`)
  }).skip()
})

const updatePasswordLinkRegExp = `http:\\/\\/localhost:\\d{4}\\/[a-zA-Z0-9-_]+\\/update-password\\?.+&current=[A-Z0-9]+`
const user = {
  ids: { user_id: 'test-user-id1' },
  primary_email_address: 'test-user1@example.com',
  password: 'ABCDefg123!',
  password_confirm: 'ABCDefg123!',
}

describe('Account App change password (via forgot password)', () => {
  let temporaryPasswordLink
  before(() => {
    cy.dropAndSeedDatabase()

    cy.createUser(user)

    cy.request({
      method: 'POST',
      url: `${Cypress.config('baseUrl')}/api/v3/users/${user.ids.user_id}/temporary_password`,
    })
    cy.task('findInStackLog', updatePasswordLinkRegExp).then(res => {
      temporaryPasswordLink = res
    })
  })

  it('displays UI elements in place', () => {
    cy.visit(temporaryPasswordLink)
    cy.findByLabelText('New password').should('be.visible')
    cy.findByLabelText('Confirm new password').should('be.visible')
    cy.findByLabelText('Revoke all access')
      .should('be.visible')
      .should('have.attr', 'value', 'false')
    cy.findByRole('button', { name: 'Update password' }).should('be.visible')
    cy.findByRole('link', { name: 'Cancel' }).should('be.visible')
  })

  it('validates before submitting the form', () => {
    cy.visit(temporaryPasswordLink)
    cy.findByRole('button', { name: 'Update password' }).click()

    cy.findErrorByLabelText('New password')
      .should('contain.text', 'New password is required')
      .and('be.visible')
    cy.findErrorByLabelText('Confirm new password')
      .should('contain.text', 'Confirm new password is required')
      .and('be.visible')

    cy.location('pathname').should('eq', `${Cypress.config('oauthRootPath')}/update-password`)
  })

  it('succeeds changing password when revoking access', () => {
    const newPassword = 'ABCDefg321!'
    cy.visit(temporaryPasswordLink)
    cy.findByLabelText('New password').type(newPassword)
    cy.findByLabelText('Confirm new password').type(`${newPassword}`)
    cy.findByLabelText('Revoke all access').check()
    cy.findByRole('button', { name: 'Update password' }).click()

    cy.findByTestId('notification')
      .should('be.visible')
      .should('contain', 'password was changed')
    cy.location('pathname').should('include', `${Cypress.config('oauthRootPath')}/login`)
  })
})
