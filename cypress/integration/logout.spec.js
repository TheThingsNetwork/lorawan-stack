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

describe('Console logout', () => {
  it('succeeds when logged in properly', () => {
    cy.login()
    cy.visit('/console')
    cy.get('header').within(() => {
      cy.findByText('admin').click()
      cy.findByText('Logout').click()
    })
    cy.findByTestId('user_id').should('exist')
    cy.url().should('include', '/login')
  })

  it('obtains a new CSRF token and succeeds when CSRF token not present', () => {
    cy.server()
    cy.route({
      method: 'POST',
      url: 'http://localhost:8080/console/api/auth/logout',
      onRequest: req => {
        expect(req.request.headers).to.have.property('X-CSRF-Token')
      },
    })

    cy.login()
    cy.visit('/console')
    cy.clearCookie('_console_csrf')
    cy.get('header').within(() => {
      cy.findByText('admin').click()
      cy.findByText('Logout').click()
    })
    cy.findByTestId('user_id').should('exist')
    cy.location('pathname').should('include', '/login')
  })
})
