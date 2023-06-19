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

describe('Network information section', () => {
  before(() => {
    cy.dropAndSeedDatabase()
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: 'admin', password: 'admin' })
  })

  it('succeeds showing registry totals', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/admin-panel/network-information`)
    cy.findByText('Network information', { selector: 'h1' }).should('be.visible')
    // Shows registry totals.
    cy.findAllByText('Total applications').should('be.visible')
    cy.findAllByText('Total gateways').should('be.visible')
    cy.findAllByText('Registered users').should('be.visible')
    cy.findAllByText('Organizations').should('be.visible')
    cy.findByText('Deployment').should('be.visible')
    cy.findByText('Available components').should('be.visible')
  })
})
