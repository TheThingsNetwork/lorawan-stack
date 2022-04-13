// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
  ids: { user_id: 'test-user' },
  name: 'Test User',
  primary_email_address: 'test-user@example.com',
  password: 'ABCDefg123!',
  password_confirm: 'ABCDefg123!',
}

describe('Account App session management', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.loginAccountApp({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('accountAppRootPath'))
    cy.loginAccountApp({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('accountAppRootPath'))
  })

  it('succeeds showing a list of sessions', () => {
    cy.visit(`${Cypress.config('accountAppRootPath')}/session-management`)

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 2)
    })
  })

  it('succeeds deleting a session', () => {
    cy.loginAccountApp({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('accountAppRootPath')}/session-management`)

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 3)
    })

    cy.findByRole('rowgroup').within(() => {
      cy.get('button', { name: 'deleteRemove this session' }).first().click()
    })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .findByText('Session removed successfully')
      .should('be.visible')

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 2)
    })
  })
})
