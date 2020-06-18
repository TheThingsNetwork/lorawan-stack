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

describe('OAuth user registration', () => {
  beforeEach(cy.seedDatabase)
  it('succeeds when using valid user data', () => {
    cy.visit('/oauth/register')

    cy.findByLabelText('User ID')
      .type('test-user')
      .should('have.value', 'test-user')
    cy.findByLabelText('Name')
      .type('Test User')
      .should('have.value', 'Test User')
    cy.findByLabelText('Email')
      .type('mail@example.com')
      .should('have.value', 'mail@example.com')
    cy.findByLabelText('Password')
      .type('ABCdefg123456!')
      .should('have.value', 'ABCdefg123456!')
    cy.findByLabelText('Confirm password')
      .type('ABCdefg123456!')
      .should('have.value', 'ABCdefg123456!')
      .type('{enter}')

    cy.findByTestId('notification')
      .should('exist')
      .should('contain', 'You have successfully registered and can login now')
    cy.location('pathname').should('include', '/oauth/login')
  })
})
