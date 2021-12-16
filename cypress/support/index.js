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

/* eslint-disable no-invalid-this */

import '@cypress/code-coverage/support'
import '@testing-library/cypress/add-commands'
import { configure } from '@testing-library/cypress'
import './commands'

const failedSpecsFilename = `./.cache/.failed-specs-${Cypress.env('MACHINE_NUMBER') || '0'}.txt`

configure({ testIdAttribute: 'data-test-id' })
Cypress.SelectorPlayground.defaults({
  selectorPriority: ['data-test-id', 'id', 'class', 'tag', 'attributes', 'nth-child'],
})

// Enable fail early if set.
afterEach(function () {
  if (this.currentTest.state === 'failed' && Cypress.env('FAIL_FAST')) {
    cy.log('Skipping rest of run due to test failure (fail fast)')
    cy.writeFile(failedSpecsFilename, this.currentTest.invocationDetails.relativeFile)
  }
})

// Skip remaining runs if fail early is set.
const skipIfNecessary = function () {
  cy.task('fileExists', failedSpecsFilename).then(content => {
    if (content !== '' && content !== false && Cypress.env('FAIL_FAST')) {
      this.currentTest.pending = true
      Cypress.runner.stop()
    }
  })
}

before(skipIfNecessary)
beforeEach(skipIfNecessary)
