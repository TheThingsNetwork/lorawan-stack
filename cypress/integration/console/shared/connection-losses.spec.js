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
    cy.findByText('Welcome to the Console!').should('be.visible')
    cy.intercept('/api/v3/application*', { forceNetworkError: true })
    cy.intercept('/api/v3/auth_info', { times: 2 }, { forceNetworkError: true })

    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Applications/ }).click()
    })

    cy.get('footer').within(() => {
      cy.findByText(/Connection issues/).should('be.visible')
      cy.findByText(/Offline/).should('be.visible')
    })
    cy.findByTestId('toast-notification')
      .as('offlineToast')
      .findByText(/offline/)
      .should('be.visible')

    cy.get('@offlineToast', { timeout: 20000 }).should('not.be.visible')

    cy.findByTestId('toast-notification')
      .findByText(/online/)
      .should('be.visible')

    cy.get('footer').within(() => {
      cy.findByText(/Connection issues/).should('not.exist')
      cy.findByText(/Offline/).should('not.exist')
    })
  })

  it('does not see individual network errors as connection loss', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('consoleRootPath'))
    cy.findByText('Welcome to the Console!').should('be.visible')
    cy.intercept('/api/v3/application*', { forceNetworkError: true })

    cy.get('header').within(() => {
      cy.findByRole('link', { name: /Applications/ }).click()
    })

    cy.get('footer').within(() => {
      // Connection issue note will appear in the footer and
      // dissappear shortly thereafter.
      cy.findByText(/Connection issues/).should('be.visible')
      cy.findByText(/Connection issues/).should('not.exist')

      cy.findByText(/Offline/).should('not.exist')
    })

    cy.findByTestId('toast-notification').should('not.exist')

    // The error will be displayed by the consuming view.
    cy.findByTestId('error-notification').should('be.visible')
  })
})
