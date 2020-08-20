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

describe('OAuth overview', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays UI elements in place', () => {
    const user = {
      ids: { user_id: 'test-user' },
      primary_email_address: 'test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.loginOAuth({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('oauthRootPath'))

    cy.findByText(`You are logged in as ${user.ids.user_id}.`).should('be.visible')
    cy.findByRole('button', { name: 'Logout' }).should('be.visible')
    cy.findByRole('link', { name: 'Change password' })
      .should('be.visible')
      .should('have.attr', 'href')
      .and('eq', `${Cypress.config('oauthRootPath')}/update-password`)
  })

  it('succeeds when logging out', () => {
    cy.findByRole('button', { name: 'Logout' }).click()

    cy.url().should('include', `${Cypress.config('oauthRootPath')}/login`)
  })
})
