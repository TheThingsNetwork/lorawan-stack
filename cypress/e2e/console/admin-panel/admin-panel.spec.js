// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

describe('Admin Panel', () => {
  before(() => {
    cy.dropAndSeedDatabase()
    cy.intercept('/api/v3/pba/info', { fixture: 'console/packet-broker/info-registered.json' })
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: 'admin', password: 'admin' })
  })

  it('succeeds displaying different views in the admin panel', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin-panel`)

    cy.findByText('User management').should('be.visible').click()
    cy.findByText('User management', { selector: 'h1' }).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')

    cy.findByText('Peering settings').should('be.visible').click()
    cy.findByText('Packet Broker', { selector: 'h1' }).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')
  })
})
