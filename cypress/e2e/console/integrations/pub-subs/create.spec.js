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

describe('Application Pub/Sub create', () => {
  const userId = 'create-app-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'create-app-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const appId = 'pub-sub-test-application'
  const application = {
    ids: {
      application_id: appId,
    },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
  })

  describe('MQTT', () => {
    beforeEach(() => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
      )
      cy.findByLabelText('MQTT').check()
    })

    it('displays UI elements in place', () => {
      cy.findByRole('heading', { name: 'Add Pub/Sub' }).should('be.visible')
      cy.findByRole('heading', { name: 'General settings' }).should('be.visible')
      cy.findByRole('heading', { name: 'MQTT configuration' }).should('be.visible')
      cy.findByRole('heading', { name: 'Enabled event types' }).should('be.visible')
      cy.findByLabelText('Pub/Sub ID').should('have.attr', 'placeholder').and('eq', 'my-new-pubsub')
      cy.findByLabelText('Server URL')
        .should('have.attr', 'placeholder')
        .and('eq', 'mqtts://example.com')
      cy.findByLabelText('Client ID').should('have.attr', 'placeholder').and('eq', 'my-client-id')
      cy.findByLabelText('Username').should('have.attr', 'placeholder').and('eq', 'my-username')
      cy.findByLabelText('Password').should('have.attr', 'placeholder').and('eq', 'my-password')
      cy.findByLabelText('Base topic').should('have.attr', 'placeholder').and('eq', 'base-topic')
      cy.findByText('For each enabled message type an optional sub-topic can be defined').should(
        'be.visible',
      )
      cy.findByRole('button', { name: 'Add Pub/Sub' }).should('be.visible')
    })

    describe('has credentials', () => {
      beforeEach(() => {
        cy.findByLabelText('Use credentials').check({ force: true })
      })

      it('validates before submitting an empty form', () => {
        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.findErrorByLabelText('Pub/Sub ID')
          .should('contain.text', 'Pub/Sub ID is required')
          .and('be.visible')
        cy.findErrorByLabelText('Server URL')
          .should('contain.text', 'Server URL is required')
          .and('be.visible')
        cy.findErrorByLabelText('Client ID')
          .should('contain.text', 'Client ID is required')
          .and('be.visible')
        cy.findErrorByLabelText('Username')
          .should('contain.text', 'Username is required')
          .and('be.visible')
        cy.findErrorByLabelText('Subscribe QoS')
          .should('contain.text', 'Subscribe QoS is required')
          .and('be.visible')
        cy.findErrorByLabelText('Publish QoS')
          .should('contain.text', 'Publish QoS is required')
          .and('be.visible')
        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
        )
      })

      it('succeeds adding pub-sub', () => {
        const pubSub = {
          id: 'with-creds-mqtt',
          serverUrl: 'mqtts://example.com',
          clientId: 'with-creds-mqtt-id',
          username: 'test-mqtt-username',
          subscribe_qos: 'AT_MOST_ONCE',
          publish_qos: 'AT_MOST_ONCE',
          format: 'json',
          uplinkSubTopic: 'uplink-test-sub-topic',
        }
        cy.findByLabelText('Pub/Sub ID').type(pubSub.id)
        cy.findByLabelText('Server URL').type(pubSub.serverUrl)
        cy.findByLabelText('Client ID').type(pubSub.clientId)
        cy.findByLabelText('Username').type(pubSub.username)
        cy.findByLabelText('Subscribe QoS').selectOption(pubSub.subscribe_qos)
        cy.findByLabelText('Publish QoS').selectOption(pubSub.publish_qos)
        cy.findByLabelText('Pub/Sub format').selectOption(pubSub.format)
        cy.findByLabelText('Uplink message').check()
        cy.findByLabelText('Uplink message')
          .parents('[data-test-id="form-field"]')
          .within(() => {
            cy.findByPlaceholderText('sub-topic').type(pubSub.uplinkSubTopic)
          })

        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs`,
        )
      })
    })

    describe('has no credentials', () => {
      beforeEach(() => {
        cy.findByLabelText('Use credentials').uncheck({ force: true })
      })

      it('validates before submitting an empty form', () => {
        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.findErrorByLabelText('Pub/Sub ID')
          .should('contain.text', 'Pub/Sub ID is required')
          .and('be.visible')
        cy.findErrorByLabelText('Server URL')
          .should('contain.text', 'Server URL is required')
          .and('be.visible')
        cy.findErrorByLabelText('Client ID')
          .should('contain.text', 'Client ID is required')
          .and('be.visible')
        cy.findErrorByLabelText('Subscribe QoS')
          .should('contain.text', 'Subscribe QoS is required')
          .and('be.visible')
        cy.findErrorByLabelText('Publish QoS')
          .should('contain.text', 'Publish QoS is required')
          .and('be.visible')
        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
        )
      })

      it('succeeds adding pub-sub', () => {
        const pubSub = {
          id: 'no-creds-mqtt',
          serverUrl: 'mqtts://example.com',
          clientId: 'no-creds-mqtt-id',
          subscribe_qos: 'AT_MOST_ONCE',
          publish_qos: 'AT_MOST_ONCE',
          format: 'json',
          uplinkSubTopic: 'uplink-test-sub-topic',
        }
        cy.findByLabelText('Pub/Sub ID').type(pubSub.id)
        cy.findByLabelText('Server URL').type(pubSub.serverUrl)
        cy.findByLabelText('Client ID').type(pubSub.clientId)
        cy.findByLabelText('Use credentials').uncheck({ force: true })
        cy.findByLabelText('Subscribe QoS').selectOption(pubSub.subscribe_qos)
        cy.findByLabelText('Publish QoS').selectOption(pubSub.publish_qos)
        cy.findByLabelText('Pub/Sub format').selectOption(pubSub.format)
        cy.findByLabelText('Uplink message').check()
        cy.findByLabelText('Uplink message')
          .parents('[data-test-id="form-field"]')
          .within(() => {
            cy.findByPlaceholderText('sub-topic').type(pubSub.uplinkSubTopic)
          })
        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs`,
        )
      })
    })
  })

  describe('Nats', () => {
    beforeEach(() => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
      )
      cy.findByLabelText('NATS').check()
    })

    it('displays UI elements in place', () => {
      cy.findByRole('heading', { name: 'Add Pub/Sub' }).should('be.visible')
      cy.findByRole('heading', { name: 'General settings' }).should('be.visible')
      cy.findByRole('heading', { name: 'NATS configuration' }).should('be.visible')
      cy.findByRole('heading', { name: 'Enabled event types' }).should('be.visible')
      cy.findByLabelText('Pub/Sub ID').should('have.attr', 'placeholder').and('eq', 'my-new-pubsub')
      cy.findByLabelText('Username').should('have.attr', 'placeholder').and('eq', 'my-username')
      cy.findByLabelText('Password').should('have.attr', 'placeholder').and('eq', 'my-password')
      cy.findByLabelText('Address').should('have.attr', 'placeholder').and('eq', 'nats.example.com')
      cy.findByLabelText('Port').should('have.attr', 'placeholder').and('eq', '4222')
      cy.findByLabelText('Base topic').should('have.attr', 'placeholder').and('eq', 'base-topic')
      cy.findByText('For each enabled message type an optional sub-topic can be defined').should(
        'be.visible',
      )
      cy.findByRole('button', { name: 'Add Pub/Sub' }).should('be.visible')
    })

    describe('has credentials', () => {
      beforeEach(() => {
        cy.findByLabelText('Use credentials').check({ force: true })
      })

      it('validates before submitting an empty form', () => {
        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.findErrorByLabelText('Pub/Sub ID')
          .should('contain.text', 'Pub/Sub ID is required')
          .and('be.visible')
        cy.findErrorByLabelText('Username')
          .should('contain.text', 'Username is required')
          .and('be.visible')
        cy.findErrorByLabelText('Password')
          .should('contain.text', 'Password is required')
          .and('be.visible')
        cy.findErrorByLabelText('Address')
          .should('contain.text', 'Address is required')
          .and('be.visible')
        cy.findErrorByLabelText('Port').should('contain.text', 'Port is required').and('be.visible')
        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
        )
      })

      it('succeeds adding pub-sub', () => {
        const pubSub = {
          id: 'with-creds-nats',
          username: 'test-nats-username',
          password: 'test-nats-password',
          address: 'nats.example.com',
          port: '4222',
          format: 'json',
          uplinkSubTopic: 'uplink-test-sub-topic',
        }
        cy.findByLabelText('Pub/Sub ID').type(pubSub.id)
        cy.findByLabelText('Username').type(pubSub.username)
        cy.findByLabelText('Password').type(pubSub.password)
        cy.findByLabelText('Address').type(pubSub.address)
        cy.findByLabelText('Port').type(pubSub.port)
        cy.findByLabelText('Pub/Sub format').selectOption(pubSub.format)
        cy.findByLabelText('Uplink message').check()
        cy.findByLabelText('Uplink message')
          .parents('[data-test-id="form-field"]')
          .within(() => {
            cy.findByPlaceholderText('sub-topic').type(pubSub.uplinkSubTopic)
          })

        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs`,
        )
      })
    })

    describe('has no credentials', () => {
      beforeEach(() => {
        cy.findByLabelText('Use credentials').uncheck({ force: true })
      })

      it('validates before submitting an empty form', () => {
        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.findErrorByLabelText('Pub/Sub ID')
          .should('contain.text', 'Pub/Sub ID is required')
          .and('be.visible')
        cy.findErrorByLabelText('Address')
          .should('contain.text', 'Address is required')
          .and('be.visible')
        cy.findErrorByLabelText('Port').should('contain.text', 'Port is required').and('be.visible')
        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
        )
      })

      it('succeeds adding pub-sub', () => {
        const pubSub = {
          id: 'no-creds-nats',
          address: 'nats.example.com',
          port: '4222',
          format: 'json',
          uplinkSubTopic: 'uplink-test-sub-topic',
        }
        cy.findByLabelText('Pub/Sub ID').type(pubSub.id)
        cy.findByLabelText('Address').type(pubSub.address)
        cy.findByLabelText('Port').type(pubSub.port)
        cy.findByLabelText('Pub/Sub format').selectOption(pubSub.format)
        cy.findByLabelText('Uplink message').check()
        cy.findByLabelText('Uplink message')
          .parents('[data-test-id="form-field"]')
          .within(() => {
            cy.findByPlaceholderText('sub-topic').type(pubSub.uplinkSubTopic)
          })

        cy.findByRole('button', { name: 'Add Pub/Sub' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs`,
        )
      })
    })
  })

  describe('Disabled Providers', () => {
    const description = 'Changing the Pub/Sub provider has been disabled by an administrator'

    describe('NATS disabled', () => {
      const response = {
        configuration: {
          pubsub: {
            providers: {
              nats: 'DISABLED',
            },
          },
        },
      }

      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
        cy.visit(
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
        )

        cy.intercept('GET', `/api/v3/as/configuration`, response)
      })
      it('succeeds setting MQTT as default provider', () => {
        cy.findByLabelText('NATS').should('be.disabled')
        cy.findByText(description).should('be.visible')
      })
    })

    describe('MQTT disabled', () => {
      const description = 'Changing the Pub/Sub provider has been disabled by an administrator'
      const response = {
        configuration: {
          pubsub: {
            providers: {
              mqtt: 'DISABLED',
            },
          },
        },
      }

      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
        cy.visit(
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
        )
        cy.intercept('GET', `/api/v3/as/configuration`, response)
      })

      it('succeeds setting NATS as default provider', () => {
        cy.findByLabelText('MQTT').should('be.disabled')
        cy.findByText(description).should('be.visible')
      })
    })

    describe('MQTT and NATS disabled', () => {
      const response = {
        configuration: {
          pubsub: {
            providers: {
              mqtt: 'DISABLED',
              nats: 'DISABLED',
            },
          },
        },
      }

      beforeEach(() => {
        cy.loginConsole({ user_id: userId, password: user.password })
        cy.on('uncaught:exception', () => false)
        cy.visit(
          `${Cypress.config('consoleRootPath')}/applications/${appId}/integrations/pubsubs/add`,
        )
        cy.intercept('GET', `/api/v3/as/configuration`, response)
      })

      it('succeeds showing not found page', () => {
        cy.findByRole('heading', { name: /Not found/ }).should('be.visible')
      })
    })
  })
})
