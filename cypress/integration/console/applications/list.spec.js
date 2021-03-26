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

describe('Applications list', () => {
  const userId = 'list-apps-test-user'
  const user = {
    ids: { user_id: 'list-apps-test-user' },
    primary_email_address: 'list-apps-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const appIds = ['xyz-test-app', 'some-test-app', 'other-test-app']

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication({ ids: { application_id: appIds[0] } }, userId)
    cy.createApplication({ ids: { application_id: appIds[1] } }, userId)
    cy.createApplication({ ids: { application_id: appIds[2] } }, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications`)
  })

  it('succeeds searching by application id', () => {
    cy.get('tbody').within(() => {
      cy.findAllByRole('row').should('have.length', 3)
    })
    cy.findByRole('cell', { name: appIds[0] }).should('be.visible')
    cy.findByRole('cell', { name: appIds[1] }).should('be.visible')
    cy.findByRole('cell', { name: appIds[2] }).should('be.visible')

    cy.findByTestId('search-input').as('searchInput')
    cy.get('@searchInput').type('xyz')

    cy.get('tbody').within(() => {
      cy.findAllByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: appIds[0] }).should('be.visible')
    cy.findByRole('cell', { name: appIds[1] }).should('not.exist')
    cy.findByRole('cell', { name: appIds[2] }).should('not.exist')

    cy.get('@searchInput').clear()
    cy.get('@searchInput').type('some')

    cy.get('tbody').within(() => {
      cy.findByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: appIds[0] }).should('not.exist')
    cy.findByRole('cell', { name: appIds[1] }).should('be.visible')
    cy.findByRole('cell', { name: appIds[2] }).should('not.exist')

    cy.get('@searchInput').clear()
    cy.get('@searchInput').type('test-app')

    cy.get('tbody').within(() => {
      cy.findAllByRole('row').should('have.length', 3)
    })
    cy.findByRole('cell', { name: appIds[0] }).should('be.visible')
    cy.findByRole('cell', { name: appIds[1] }).should('be.visible')
    cy.findByRole('cell', { name: appIds[2] }).should('be.visible')
  })
})
