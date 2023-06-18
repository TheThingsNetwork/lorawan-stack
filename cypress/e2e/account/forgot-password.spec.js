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

describe('Account App forgot password', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays UI elements in place', () => {
    cy.visit(`${Cypress.config('accountAppRootPath')}/forgot-password`)
    cy.findByText('Reset password').should('be.visible')
    cy.findByLabelText('User ID').should('be.visible')
    cy.findByRole('button', { name: 'Send' }).should('be.visible')
    cy.findByRole('link', { name: 'Cancel' }).should('be.visible')
  })

  it('validates before submitting the form', () => {
    cy.visit(`${Cypress.config('accountAppRootPath')}/forgot-password`)
    cy.findByRole('button', { name: 'Send' }).click()
    cy.findErrorByLabelText('User ID')
      .should('contain.text', 'User ID is required')
      .and('be.visible')
  })

  it('succeeds when submitting the form', () => {
    const user = {
      ids: { user_id: 'test-user-id1' },
      primary_email_address: 'test-user1@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.visit(`${Cypress.config('accountAppRootPath')}/forgot-password`)
    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByRole('button', { name: 'Send' }).click()
    cy.findByTestId('notification').should('be.visible').should('contain', 'reset instruction')
    cy.location('pathname').should('include', `${Cypress.config('accountAppRootPath')}/login`)
  })
})
