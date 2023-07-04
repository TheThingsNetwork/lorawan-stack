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

const user = {
  ids: { user_id: 'authorization-code-test-user' },
  name: 'Test User',
  primary_email_address: 'test-user@example.com',
  password: 'ABCDefg123!',
  password_confirm: 'ABCDefg123!',
}

describe('Account App code view', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  beforeEach(() => {
    cy.loginAccountApp({ user_id: user.ids.user_id, password: user.password })
  })

  it('displays UI elements in place', () => {
    const code = '12345code'

    cy.visit(`${Cypress.config('accountAppRootPath')}/code?code=${code}`)
    cy.findByText('Authorization code').should('be.visible')
    cy.findByText('Your authorization code is:').should('be.visible')
    cy.findByText(code)
    cy.findByRole('link', { name: /Back to Account/ }).should('be.visible')
  })

  it('redirects back if no code is supplied', () => {
    cy.visit(`${Cypress.config('accountAppRootPath')}/code`)
    cy.location('pathname').should('eq', Cypress.config('accountAppRootPath'))
  })
})
