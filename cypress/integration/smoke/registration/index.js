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

import { defineSmokeTest } from '../utils'

const loginConsole = defineSmokeTest('succeeds registering and logging into the Console', () => {
  const user = {
    user_id: 'console-login-test-user',
    name: 'Console Login Test User',
    password: '123456QWERTY!',
    email: 'console-login-test-user@example.com',
  }
  cy.visit(Cypress.config('consoleRootPath'))

  cy.findByRole('button', { name: 'Create an account' }).click()
  cy.findByLabelText('User ID').type(user.user_id)
  cy.findByLabelText('Name').type(user.name)
  cy.findByLabelText('Email').type(user.email)
  cy.findByLabelText('Password').type(user.password)
  cy.findByLabelText('Confirm password').type(user.password)
  cy.findByRole('button', { name: 'Register' }).click()

  cy.findByTestId('notification')
    .should('be.visible')
    .should('contain', 'You have successfully registered and can login now')

  // Login.
  // TODO: https://github.com/TheThingsNetwork/lorawan-stack/issues/2923
  cy.visit(Cypress.config('consoleRootPath'))
  cy.findByLabelText('User ID').type(user.user_id)
  cy.findByLabelText('Password').type(`${user.password}`)
  cy.findByRole('button', { name: 'Login' }).click()

  cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/`)
  cy.findByTestId('full-error-view').should('not.exist')
})

const loginOAuth = defineSmokeTest('succeeds registering and logging into the Oauth app', () => {
  const user = {
    user_id: 'oauth-login-test-user',
    name: 'OAuth Login Test User',
    password: '123456QWERTY!',
    email: 'oauth-login-test-user@example.com',
  }
  cy.visit(Cypress.config('oauthRootPath'))

  cy.findByRole('button', { name: 'Create an account' }).click()
  cy.findByLabelText('User ID').type(user.user_id)
  cy.findByLabelText('Name').type(user.name)
  cy.findByLabelText('Email').type(user.email)
  cy.findByLabelText('Password').type(user.password)
  cy.findByLabelText('Confirm password').type(user.password)
  cy.findByRole('button', { name: 'Register' }).click()

  cy.findByTestId('notification')
    .should('be.visible')
    .should('contain', 'You have successfully registered and can login now')

  cy.findByLabelText('User ID').type(user.user_id)
  cy.findByLabelText('Password').type(`${user.password}`)
  cy.findByRole('button', { name: 'Login' }).click()

  cy.location('pathname').should('eq', `${Cypress.config('oauthRootPath')}/`)
  cy.findByTestId('full-error-view').should('not.exist')
})

export default [loginConsole, loginOAuth]
