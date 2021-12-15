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

describe('Gateway location', () => {
  const gatewayId = 'test-gateway-location'
  const gateway = { ids: { gateway_id: gatewayId } }
  const coordinates = {
    latitude: 2,
    longitude: 2,
    altitude: 2,
  }
  const userId = 'edit-gateway-location-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-gateway-location-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const updatedGatewayLocation = {
    gateway: {
      antennas: [
        {
          location: {
            altitude: 1,
            latitude: 1,
            longitude: 1,
          },
        },
      ],
      location_public: true,
      update_location_from_status: false,
    },
    field_mask: {
      paths: ['antennas', 'location_public', 'update_location_from_status'],
    },
  }

  const updatedGatewayNullLocation = {
    gateway: {
      antennas: [{ location: null }],
      location_public: true,
      update_location_from_status: false,
    },
    field_mask: {
      paths: ['antennas', 'location_public', 'update_location_from_status'],
    },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createGateway(gateway, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('succeeds editing latitude, longitude, altitude', () => {
    cy.updateGateway(gatewayId, updatedGatewayLocation)
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)
    cy.findByLabelText('Latitude').type(coordinates.latitude)
    cy.findByLabelText('Longitude').type(coordinates.longitude)
    cy.findByLabelText('Altitude').type(coordinates.altitude)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Location updated`)
      .should('be.visible')
  })

  it('succeeds editing the location when location was null', () => {
    cy.updateGateway(gatewayId, updatedGatewayNullLocation)
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)
    cy.findByLabelText('Latitude').type(coordinates.latitude)
    cy.findByLabelText('Longitude').type(coordinates.longitude)
    cy.findByLabelText('Altitude').type(coordinates.altitude)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Location updated`)
      .should('be.visible')
  })

  it('succeeds editing latitude and longitude based map widget location change', () => {
    cy.updateGateway(gatewayId, updatedGatewayLocation)
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)
    cy.findByTestId('location-map').should('be.visible')
    cy.findByTestId('location-map').click(30, 30)

    cy.findByLabelText('Latitude').should('not.eq', coordinates.latitude)
    cy.findByLabelText('Longitude').should('not.eq', coordinates.longitude)
    cy.findByLabelText('Altitude').should('not.eq', coordinates.altitude)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Location updated`)
      .should('be.visible')
  })

  it('succeeds deleting location entry', () => {
    cy.updateGateway(gatewayId, updatedGatewayLocation)
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/location`)
    cy.findByRole('button', { name: /Remove location/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Remove location data', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Remove location/ }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Location deleted`)
      .should('be.visible')
    cy.findByLabelText('Latitude').should('have.attr', 'value').and('eq', '')
    cy.findByLabelText('Longitude').should('have.attr', 'value').and('eq', '')
    cy.findByLabelText('Altitude').should('have.attr', 'value').and('eq', '')
  })
})
