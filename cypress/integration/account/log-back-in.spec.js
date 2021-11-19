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

describe('Log back in after the session has expired', () => {
  const user = {
    ids: { user_id: 'test-user-id1' },
    primary_email_address: 'test-user1@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('succeeds showing modal on session expired', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}`)
    cy.clearCookie('_console_auth')
    cy.clearLocalStorage()

    cy.findByLabelText('Please sign in again').should('be.visible')
    cy.findByLabelText('Reload').click()
    cy.findByRole('link', { name: 'Create an account' }).should('be.visible')
    cy.findByRole('link', { name: 'Forgot password?' }).should('be.visible')
    cy.findByLabelText('User ID').should('be.visible')
    cy.findByLabelText('Password').should('be.visible')
  })
})
