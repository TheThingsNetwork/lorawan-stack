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

describe('OAuth login', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays UI elements in place', () => {
    cy.visit(Cypress.config('oauthRootPath'))

    cy.findByRole('link')
      .should('have.attr', 'href', `${Cypress.config('oauthRootPath')}/`)
      .findByRole('img')
      .should('be.visible')
      .should('have.attr', 'src', `${Cypress.config('assetsRootPath')}/logo.svg`)
      .should('have.attr', 'alt', 'The Things Stack for LoRaWAN Logo')
    cy.findByText('The Things Stack for LoRaWAN', { selector: 'h1' }).should('be.visible')
    cy.findByText('Please login to continue').should('be.visible')
    cy.findByRole('button', { name: 'Create an account' }).should('be.visible')
    cy.findByRole('button', { name: 'Forgot password?' }).should('be.visible')
    cy.findByLabelText('User ID').should('be.visible')
    cy.findByLabelText('Password').should('be.visible')
    cy.title().should('eq', `Login - ${Cypress.config('siteName')}`)
  })

  it('validates before submitting an empty form', () => {
    cy.visit(Cypress.config('oauthRootPath'))

    cy.findByRole('button', { name: 'Login' }).click()

    cy.findErrorByLabelText('User ID')
      .should('contain.text', 'User ID is required')
      .and('be.visible')
    cy.findErrorByLabelText('Password')
      .should('contain.text', 'Password is required')
      .and('be.visible')

    cy.location('pathname').should('eq', `${Cypress.config('oauthRootPath')}/login`)
  })

  it('succeeds logging in when using valid credentials', () => {
    const user = {
      ids: { user_id: 'test-user' },
      primary_email_address: 'test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
    cy.createUser(user)
    cy.visit(Cypress.config('oauthRootPath'))

    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(`${user.password}{enter}`)

    cy.location('pathname').should('eq', `${Cypress.config('oauthRootPath')}/`)
    cy.findByTestId('full-error-view').should('not.exist')
  })

  it('displays an error when using invalid credentials', () => {
    const user = { user_id: 'userwrong', password: 'userWr0ng!' }
    cy.visit(Cypress.config('oauthRootPath'))

    cy.findByLabelText('User ID').type(user.user_id)
    cy.findByLabelText('Password').type(`${user.password}{enter}`)

    cy.findByTestId('error-notification')
      .should('be.visible')
      .findByText('incorrect password or user ID')
      .should('be.visible')
    cy.location('pathname').should('eq', `${Cypress.config('oauthRootPath')}/login`)
  })
})
