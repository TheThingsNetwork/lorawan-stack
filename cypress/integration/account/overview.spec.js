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

const user = {
  ids: { user_id: 'test-user' },
  name: 'Test User',
  primary_email_address: 'test-user@example.com',
  password: 'ABCDefg123!',
  password_confirm: 'ABCDefg123!',
}

describe('Account App overview', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  it('displays UI elements in place', () => {
    cy.createUser(user)
    cy.loginAccountApp({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('accountAppRootPath'))

    // Profile card section.
    cy.findByText(user.name, { selector: 'h3' }).should('be.visible')
    cy.findByText(user.ids.user_id).should('be.visible')
    cy.findByAltText('Profile picture')
      .should('be.visible')
      .and('have.attr', 'src')
      .and('match', /missing-profile-picture/)
    cy.findByRole('link', { name: /Edit profile settings/ }).should('be.visible')

    // Info text section.
    cy.findByText(`Welcome, ${user.name}! 👋`).should('be.visible')
    cy.findByText('You have successfully logged into the Account App', { exact: false }).should(
      'be.visible',
    )
    cy.findByText('you can use the button below to head over to the Console', {
      exact: false,
    }).should('be.visible')
    cy.findByRole('link', { name: 'Go to the Console' }).should('be.visible')
  })

  it('succeeds when logging out', () => {
    cy.loginAccountApp({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('accountAppRootPath'))
    cy.get('header').within(() => {
      cy.findByTestId('profile-dropdown')
        .should('contain', user.name)
        .as('profileDropdown')

      cy.get('@profileDropdown').click()
      cy.get('@profileDropdown')
        .findByText('Logout')
        .click()
    })

    cy.url().should('include', `${Cypress.config('accountAppRootPath')}/login`)
  })

  it('succeeds linking to the Console', () => {
    cy.loginAccountApp({ user_id: user.ids.user_id, password: user.password })
    cy.visit(Cypress.config('accountAppRootPath'))
    cy.findByRole('link', { name: 'Go to the Console' }).click()

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/`)
  })
})
