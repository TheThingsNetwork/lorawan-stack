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

describe('Header', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })
  describe('Console logout', () => {
    const logout = userName => {
      cy.get('header').within(() => {
        cy.findByTestId('profile-dropdown').should('contain', userName).as('profileDropdown')

        cy.get('@profileDropdown').click()
        cy.get('@profileDropdown').findByText('Logout').click()
      })
    }

    it('succeeds when logged in properly', () => {
      const user = {
        ids: { user_id: 'test-logout-user' },
        primary_email_address: 'test-logout-user@example.com',
        password: 'ABCDefg123!',
        password_confirm: 'ABCDefg123!',
      }
      cy.createUser(user)
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(Cypress.config('consoleRootPath'))

      logout(user.ids.user_id)

      cy.location('pathname').should('eq', `${Cypress.config('accountAppRootPath')}/login`)
    })

    it('obtains a new CSRF token and succeeds when CSRF token not present', () => {
      const user = {
        ids: { user_id: 'test-logout-user2' },
        primary_email_address: 'test-logout-user2@example.com',
        password: 'ABCDefg123!',
        password_confirm: 'ABCDefg123!',
      }
      const baseUrl = Cypress.config('baseUrl')
      const consoleRootPath = Cypress.config('consoleRootPath')
      const accountAppRootPath = Cypress.config('accountAppRootPath')

      cy.intercept('POST', `${baseUrl}${consoleRootPath}/api/auth/logout`, req => {
        // Asserting on the request headers
        expect(req.headers).to.have.property('x-csrf-token')
      }).as('logout')

      cy.createUser(user)
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(Cypress.config('consoleRootPath'))
      cy.clearCookie('_console_csrf')

      logout(user.ids.user_id)

      cy.location('pathname').should('eq', `${accountAppRootPath}/login`)
    })

    it('displays UI elements in place', () => {
      const user = {
        ids: { user_id: 'test-header-user' },
        primary_email_address: 'test-header-user@example.com',
        password: 'ABCDefg123!',
        password_confirm: 'ABCDefg123!',
        name: 'Test Header User',
      }

      cy.createUser(user)
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(Cypress.config('consoleRootPath'))

      cy.get('header').within(() => {
        cy.findByTestId('profile-dropdown').should('contain', user.name)

        cy.findByAltText('Profile picture')
          .should('be.visible')
          .and('have.attr', 'src')
          .and('match', /missing-profile-picture/)
      })
    })
  })
})
