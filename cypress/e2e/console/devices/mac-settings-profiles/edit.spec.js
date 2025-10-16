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

describe('MAC settings profile edit', () => {
  const userId = 'mac-settings-profile-edit-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'mac-settings-profile-edit-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const applicationId = 'app-mac-settings-profile-edit-test'
  const application = {
    ids: { application_id: applicationId },
    name: 'Application for MAC settings profile edit test',
  }

  const profileId = 'edit-test-profile'
  const macSettingsProfile = {
    ids: {
      application_ids: { application_id: applicationId },
      profile_id: profileId,
    },
    mac_settings: {
      supports_32_bit_f_cnt: true,
      rx1_delay: 1,
      rx2_data_rate_index: 0,
      adr: {
        dynamic: {
          margin: 10,
          min_nb_trans: 1,
          max_nb_trans: 3,
        },
      },
    },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createMacSettingsProfile(applicationId, macSettingsProfile)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })

    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/mac-settings-profiles/${profileId}`,
    )
  })

  it('displays MAC settings profile edit form with existing values', () => {
    cy.findByText('Edit a MAC settings profile', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Profile ID').should('be.disabled').and('have.value', profileId)
    cy.findByLabelText('Rx1 delay').should('have.value', '1')
    cy.findByLabelText('Rx2 data rate index').should('have.value', '0')
    cy.findByLabelText('Dynamic mode').should('be.checked')
    cy.findByLabelText('ADR margin').should('have.value', '10')
  })

  it('succeeds to update MAC settings profile', () => {
    // eslint-disable-next-line cypress/unsafe-to-chain-command
    cy.findByLabelText('Rx1 delay').clear().type('3')
    // eslint-disable-next-line cypress/unsafe-to-chain-command
    cy.findByLabelText('Rx2 data rate index').clear().type('5')
    // eslint-disable-next-line cypress/unsafe-to-chain-command
    cy.findByLabelText('ADR margin').clear().type('15')

    cy.findByRole('button', { name: /Save changes/ }).click()

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'MAC settings profile updated')
  })

  it('succeeds to change ADR mode from dynamic to static', () => {
    cy.findByLabelText('Static mode').check()
    cy.findByLabelText('ADR data rate index').type('3')
    cy.findByLabelText('ADR transmission power index').type('1')
    cy.findByLabelText('ADR number of transmissions').type('2')

    cy.findByRole('button', { name: /Save changes/ }).click()

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'MAC settings profile updated')
  })

  it('succeeds to disable ADR', () => {
    cy.findByLabelText('Disabled').check()

    cy.findByRole('button', { name: /Save changes/ }).click()

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'MAC settings profile updated')
  })
})
