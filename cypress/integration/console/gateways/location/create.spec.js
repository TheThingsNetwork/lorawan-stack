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

describe('Gateway location create', () => {
  const gatewayId = 'test-gateway-location'
  const gateway = { ids: { gateway_id: gatewayId } }
  const coordinates = {
    latitude: 56.95,
    longitude: 24.11,
    altitude: 0,
  }
  const userId = 'create-gateway-location-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'create-gateway-location-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.createGateway(gateway, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('displays UI elements in place', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)

    cy.findByText('Location', { selector: 'h1' }).should('be.visible')
    cy.findByText('Gateway antenna location settings', { selector: 'h3' }).should('be.visible')
    cy.findByLabelText('Privacy').should('exist')
    cy.findDescriptionByLabelText('Privacy')
      .should('contain', 'The location of this gateway may be publicly displayed')
      .and('be.visible')
    cy.findByLabelText('Location source').should('exist')
    cy.findDescriptionByLabelText('Location source')
      .should('contain', 'Update the location of this gateway based on incoming status messages')
      .and('be.visible')
    cy.findByTestId('location-map').should('be.visible')
    cy.findByLabelText('Latitude').should('be.visible')
    cy.findDescriptionByLabelText('Latitude')
      .should('contain', 'The north-south position in degrees, where 0 is the equator')
      .and('be.visible')
    cy.findByLabelText('Longitude').should('be.visible')
    cy.findDescriptionByLabelText('Longitude')
      .should(
        'contain',
        'The east-west position in degrees, where 0 is the prime meridian (Greenwich)',
      )
      .and('be.visible')
    cy.findByLabelText('Altitude').should('be.visible')
    cy.findDescriptionByLabelText('Altitude')
      .should('contain', 'The altitude in meters, where 0 means sea level')
      .and('be.visible')

    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findErrorByLabelText('Latitude')
      .should('contain.text', 'Latitude is required')
      .and('be.visible')
    cy.findErrorByLabelText('Longitude')
      .should('contain.text', 'Longitude is required')
      .and('be.visible')
    cy.findErrorByLabelText('Altitude')
      .should('contain.text', 'Altitude is required')
      .and('be.visible')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`,
    )
  })

  it('disables inputs when location source is checked', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)
    cy.findByLabelText('Location source').check()

    cy.findByLabelText('Latitude').should('be.disabled')
    cy.findByLabelText('Longitude').should('be.disabled')
    cy.findByLabelText('Altitude').should('be.disabled')
    cy.findByRole('button', { name: /Remove location entry/ }).should('be.disabled')
  })

  it('successfully saves location', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)
    cy.findByLabelText('Latitude').type(coordinates.latitude)
    cy.findByLabelText('Longitude').type(coordinates.longitude)
    cy.findByLabelText('Altitude').type(coordinates.altitude)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Location updated`)
      .should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
  })
})
