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

describe('Console login', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays an error when using invalid credentials', () => {
    const usr = { user_id: 'does-not-exist-usr', password: '12345QWERTY!' }
    cy.visit(Cypress.config('consoleRootPath'))

    cy.findByLabelText('User ID').type(usr.user_id)
    cy.findByLabelText('Password').type(`${usr.password}{enter}`)

    cy.location('pathname').should('include', Cypress.config('oauthRootPath'))
    cy.findByTestId('error-notification')
      .should('be.visible')
      .findByText('incorrect password or user ID')
      .should('be.visible')
  })

  it('succeeds logging in with valid credentials', () => {
    const user = {
      ids: { user_id: 'test-user' },
      primary_email_address: 'test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.visit(Cypress.config('consoleRootPath'))

    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(`${user.password}{enter}`)

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/`)
    cy.findByTestId('full-error-view').should('not.exist')
  })
})
