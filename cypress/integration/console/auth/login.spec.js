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
  primary_email_address: 'test-user@example.com',
  password: 'ABCDefg123!',
  password_confirm: 'ABCDefg123!',
}

describe('Account App login', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  it('displays UI elements in place', () => {
    cy.visit(Cypress.config('accountAppRootPath'))

    cy.get('header').within(() => {
      cy.findByRole('link')
        .should('have.attr', 'href', `${Cypress.config('accountAppRootPath')}/`)
        .findByRole('img')
        .should('be.visible')
        .should('have.attr', 'src', `${Cypress.config('accountAppAssetsRootPath')}/account.svg`)
        .should('have.attr', 'alt', `${Cypress.config('accountAppSiteName')} Logo`)
    })

    cy.findByRole('link', { name: 'Create an account' }).should('be.visible')
    cy.findByRole('link', { name: 'Forgot password?' }).should('be.visible')
    cy.findByLabelText('User ID').should('be.visible')
    cy.findByLabelText('Password').should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.visit(Cypress.config('accountAppRootPath'))

    cy.findByRole('button', { name: 'Login' }).click()

    cy.findErrorByLabelText('User ID')
      .should('contain.text', 'User ID is required')
      .and('be.visible')
    cy.findErrorByLabelText('Password')
      .should('contain.text', 'Password is required')
      .and('be.visible')

    cy.location('pathname').should('eq', `${Cypress.config('accountAppRootPath')}/login`)
  })

  it('succeeds logging in with valid credentials', () => {
    const location = `${Cypress.config('consoleRootPath')}/applications/`
    cy.visit(location)

    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(`${user.password}{enter}`)

    cy.location('pathname').should('eq', location)
    cy.findByText(/Applications \(0\)/).should('be.visible')
    cy.findByTestId('full-error-view').should('not.exist')
  })

  it('displays an error when using invalid credentials', () => {
    const usr = { user_id: 'does-not-exist-usr', password: '12345QWERTY!' }
    cy.visit(Cypress.config('consoleRootPath'))

    cy.findByLabelText('User ID').type(usr.user_id)
    cy.findByLabelText('Password').type(`${usr.password}{enter}`)

    cy.location('pathname').should('include', Cypress.config('accountAppRootPath'))
    cy.findByTestId('error-notification')
      .should('be.visible')
      .contains('incorrect password or user ID')
  })

  it('applies the Console logout route when logging out', () => {
    const logout = userName => {
      cy.get('header').within(() => {
        cy.findByTestId('profile-dropdown').should('contain', userName).as('profileDropdown')

        cy.get('@profileDropdown').click()
        cy.get('@profileDropdown').findByText('Logout').click()
      })
    }
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('consoleRootPath'))
    logout(user.ids.user_id)

    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(`${user.password}{enter}`)

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/`)
    cy.findByText('Welcome to the Console!').should('be.visible')
    cy.findByTestId('full-error-view').should('not.exist')
  })

  it('displays an error when the token cannot be retrieved during initialization', () => {
    cy.on('uncaught:exception', err => {
      expect(err.name).to.equal('TokenError')
      return false
    })

    const location = `${Cypress.config('consoleRootPath')}/`
    cy.visit(location)

    cy.findByText('Login')
    cy.intercept('/console/api/auth/token', { statusCode: 500 })
    cy.findByLabelText('User ID').type(user.ids.user_id)
    cy.findByLabelText('Password').type(`${user.password}{enter}`)
    cy.findByText('TokenError').should('be.visible')
    cy.findByTestId('full-error-view').should('be.visible')
  })
})
