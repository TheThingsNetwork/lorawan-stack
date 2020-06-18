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

describe('Console login', () => {
  it('succeeds when provided with valid credentials', () => {
    cy.visit('/console')

    cy.findByLabelText('User ID')
      .type('admin')
      .should('have.value', 'admin')
    cy.findByLabelText('Password')
      .type('admin')
      .should('have.value', 'admin')
      .type('{enter}')

    cy.findAllByText('Welcome to the Console!', { timeout: 10000 }).should('exist')
  })

  it('fails and displays an error when using invalid login credentials', () => {
    cy.visit('/console')

    cy.findByLabelText('User ID', { exact: false })
      .type('adminwrong')
      .should('have.value', 'adminwrong')
    cy.findByLabelText('Password', { exact: false })
      .type('adminwrong')
      .should('have.value', 'adminwrong')
      .type('{enter}')

    cy.location('pathname').should('include', '/oauth')
    cy.findByTestId('error-notification')
      .should('exist')
      .should('contain', 'incorrect password or user ID')
  })
})
