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

  it('succeeds showing modal when session context is lost', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}`)
    cy.get('header').should('be.visible')
    // When losing console auth cookie only.
    cy.clearCookie('_console_auth')
    cy.clearLocalStorage()
    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Applications/ }).click()
    })

    cy.findByText('Please sign in again').should('be.visible')
    cy.findByText('Reload').click()
    cy.findByText(/Applications \(0\)/).should('be.visible')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/applications`)

    // When losing all cookies.
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Gateways/ }).click()
    })

    cy.findByText('Please sign in again').should('be.visible')
    cy.findByText('Reload').click()
    cy.findByTestId('full-error-view').should('not.exist')
    cy.location('pathname').should('eq', `${Cypress.config('accountAppRootPath')}/login`)
  })

  it('handles token retrieval correctly when access token was lost', () => {
    cy.visit(Cypress.config('consoleRootPath'))
    cy.findByText('Welcome to the Console!').should('be.visible')
    cy.intercept('/console/api/auth/token', { statusCode: 403 })
    cy.clearLocalStorage()
    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Applications/ }).click()
    })

    cy.findByText('Please sign in again').should('be.visible')
    cy.findByTestId('full-error-view').should('not.exist')

    cy.clearCookie('_console_auth')
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('consoleRootPath'))
    cy.findByText('Welcome to the Console!').should('be.visible')
    cy.intercept('/console/api/auth/token', { statusCode: 401 })
    cy.clearLocalStorage()
    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Applications/ }).click()
    })

    cy.findByText('Please sign in again').should('be.visible')
    cy.findByTestId('full-error-view').should('not.exist')
  })

  it('retrieves new access tokens successfully', () => {
    cy.visit(Cypress.config('consoleRootPath'))
    cy.findByText('Welcome to the Console!').should('be.visible')
    cy.clearLocalStorage()
    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Applications/ }).click()
    })

    cy.findByText('Please sign in again').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
  })

  it('forwards unexpected errors during token retrieval', () => {
    cy.visit(Cypress.config('consoleRootPath'))
    cy.findByText('Welcome to the Console!').should('be.visible')
    cy.intercept('/console/api/auth/token', { statusCode: 500 })
    cy.clearLocalStorage()
    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Applications/ }).click()
    })

    cy.findByTestId('error-notification').should('be.visible')
  })
})
