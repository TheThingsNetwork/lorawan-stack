// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

describe('MAC settings profile creation', () => {
  const userId = 'mac-settings-prof-create-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'mac-settings-profile-create-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const applicationId = 'app-mac-settings-prof-create-test'
  const application = {
    ids: { application_id: applicationId },
    name: 'Application for MAC settings profile creation test',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })

    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/mac-settings-profiles/add`,
    )
  })

  it('displays MAC settings profile creation form', () => {
    cy.findByText('Create a new MAC settings profile', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Profile ID').should('be.visible')
    cy.findByText('Frame counter width').should('be.visible')
    cy.findByLabelText('32 bit').should('be.checked')
    cy.findByLabelText('Rx1 delay').should('be.visible')
    cy.findByLabelText('Desired Rx1 delay').should('be.visible')
    cy.findByLabelText('Rx1 data rate offset').should('be.visible')
    cy.findByLabelText('Desired Rx1 data rate offset').should('be.visible')
    cy.findByLabelText('Rx2 data rate index').should('be.visible')
    cy.findByLabelText('Desired Rx2 data rate index').should('be.visible')
    cy.findByLabelText('Rx2 frequency').should('be.visible')
    cy.findByLabelText('Desired Rx2 frequency').should('be.visible')
    cy.findByLabelText('Maximum duty cycle').should('be.visible')
    cy.findByLabelText('Desired maximum duty cycle').should('be.visible')
    cy.findByText('Factory preset frequencies').should('be.visible')
    cy.findByText('Class C timeout').should('be.visible')
    cy.get('input[name="mac_settings.beacon_frequency"]').scrollIntoView()
    cy.findByLabelText('Ping slot periodicity').should('be.visible')
    cy.findByLabelText('Beacon frequency').should('be.visible')
    cy.findByLabelText('Desired beacon frequency').should('be.visible')
    cy.findByLabelText('Ping slot frequency').should('be.visible')
    cy.findByLabelText('Desired ping slot frequency').should('be.visible')
    cy.findByLabelText('Ping slot data rate index').should('be.visible')
    cy.findByLabelText('Desired ping slot data rate').should('be.visible')
    cy.findByLabelText('Status count periodicity').should('be.visible')
    cy.findByLabelText('Status time periodicity').should('be.visible')
    cy.findByText('Adaptive data rate (ADR)').should('be.visible')
  })

  it('validates required fields', () => {
    cy.findByRole('button', { name: /Create a new MAC settings profile/ }).click()
    cy.findErrorByLabelText('Profile ID')
      .should('contain.text', 'Profile ID is required')
      .and('be.visible')
  })

  it('succeeds to create basic MAC settings profile', () => {
    const profileData = {
      profile_id: 'basic-test-profile',
      rx1_delay: '2',
      rx2_data_rate_index: '3',
    }

    cy.findByLabelText('Profile ID').type(profileData.profile_id)
    // eslint-disable-next-line cypress/unsafe-to-chain-command
    cy.findByLabelText('Rx1 delay').clear().type(profileData.rx1_delay)
    // eslint-disable-next-line cypress/unsafe-to-chain-command
    cy.findByLabelText('Rx2 data rate index').clear().type(profileData.rx2_data_rate_index)

    cy.findByRole('button', { name: /Create a new MAC settings profile/ }).click()

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'MAC settings profile created')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/mac-settings-profiles`,
    )
  })

  it('succeeds to create MAC settings profile with ADR dynamic settings', () => {
    const profileData = {
      profile_id: 'adr-dynamic-profile',
      adr_margin: '10',
      min_nb_trans: '1',
      max_nb_trans: '3',
    }

    cy.findByLabelText('Profile ID').type(profileData.profile_id)

    cy.findByLabelText('ADR margin').type(profileData.adr_margin)

    cy.findByLabelText('Use default settings for number of retransmissions').uncheck()
    cy.findByLabelText('Override server defaults for NbTrans (all data rates)').check()

    cy.findByLabelText('Min. NbTrans').type(profileData.min_nb_trans)
    cy.findByLabelText('Max. NbTrans').type(profileData.max_nb_trans)

    cy.findByRole('button', { name: /Create a new MAC settings profile/ }).click()

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'MAC settings profile created')
  })

  it('succeeds to create MAC settings profile with static ADR settings', () => {
    const profileData = {
      profile_id: 'adr-static-profile',
      data_rate_index: '5',
      tx_power_index: '2',
      nb_trans: '2',
    }

    cy.findByLabelText('Profile ID').type(profileData.profile_id)

    // Select static ADR
    cy.findByLabelText('Static mode').check()
    cy.findByLabelText('ADR data rate index').type(profileData.data_rate_index)
    cy.findByLabelText('ADR transmission power index').type(profileData.tx_power_index)
    cy.findByLabelText('ADR number of transmissions').type(profileData.nb_trans)

    cy.findByRole('button', { name: /Create a new MAC settings profile/ }).click()

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'MAC settings profile created')
  })
})
