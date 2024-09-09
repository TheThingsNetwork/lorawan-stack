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

describe('Connection loss detection', () => {
  const userId = 'connection-loss-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'connection-loss-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  it('detects connection losses and attempts reconnects', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('consoleRootPath'))
    // Wait until the page is fully loaded.
    cy.get('#stage').within(() => {
      cy.findByText('No top entities yet').should('be.visible')
    })

    cy.intercept('/api/v3/application*', { forceNetworkError: true }).as('offlineIntercept')
    cy.intercept({ url: '/api/v3/auth_info', times: 3 }, { forceNetworkError: true }).as(
      'reconnectionIntercept',
    )

    cy.findByRole('link', { name: /Applications/ }).click()

    cy.get('header').within(() => {
      cy.findByText(/Offline/).should('be.visible')
    })

    // After the 'offline' toast has disappeared, wait for the reconnection intercept to resolve.
    cy.wait('@reconnectionIntercept')

    cy.get('header').within(() => {
      cy.findByText(/Offline/, { timeout: 20000 }).should('not.exist')
    })
  })
  it('does not see individual network errors as connection loss', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('consoleRootPath'))
    // Wait until the page is fully loaded.
    cy.get('#stage').within(() => {
      cy.findByText('No top entities yet').should('be.visible')
    })
    cy.intercept('/api/v3/application*', { forceNetworkError: true })

    cy.findByRole('link', { name: /Applications/ }).click()
    cy.get('header').within(() => {
      // Connection issue note will appear in the footer and
      // dissappear shortly thereafter.
      cy.findByText(/Connection issues/, { timeout: 20000 }).should('be.visible')
      cy.findByText(/Connection issues/).should('not.exist')

      cy.findByText(/Offline/).should('not.exist')
    })

    // The error will be displayed by the consuming view.
    cy.findByTestId('error-notification').should('be.visible')
  })
})
