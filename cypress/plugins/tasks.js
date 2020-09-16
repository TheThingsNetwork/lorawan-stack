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

const { execSync } = require('child_process')

const yaml = require('js-yaml')
const codeCoverageTask = require('@cypress/code-coverage/task')

// `stackConfigTask` sources stack configuration entires to `Cypress` configuration while preserving
// all entries from `cypress.json`.
const stackConfigTask = (_, config) => {
  try {
    const out = execSync('go run ./cmd/ttn-lw-stack config --yml')
    const yml = yaml.safeLoad(out)

    // General.
    config.siteName = yml.console.ui['site-name']
    config.title = yml.console.ui.title
    config.subTitle = yml.console.ui['sub-title']

    // Cluster.
    config.asBaseUrl = yml.console.ui.as['base-url']
    config.asEnabled = yml.console.ui.as.enabled
    config.nsBaseUrl = yml.console.ui.ns['base-url']
    config.nsEnabled = yml.console.ui.ns.enabled
    config.jsBaseUrl = yml.console.ui.js['base-url']
    config.jsEnabled = yml.console.ui.js.enabled
    config.isBaseUrl = yml.console.ui.is['base-url']
    config.isEnabled = yml.console.ui.is.enabled
    config.gsBaseUrl = yml.console.ui.gs['base-url']
    config.gsEnabled = yml.console.ui.gs.enabled
    config.edtcBaseUrl = yml.console.ui.edtc['base-url']
    config.edtcEnabled = yml.console.ui.edtc.enabled
    config.qrgBaseUrl = yml.console.ui.qrg['base-url']
    config.qrgEnabled = yml.console.ui.qrg.enabled

    // Console.
    config.consoleAssetsRootPath = yml.console.ui['assets-base-url']
    config.consoleRootPath = new URL(yml.console.ui['canonical-url']).pathname

    // OAuth.
    config.oauthRootPath = new URL(yml.is.oauth.ui['canonical-url']).pathname
    config.oauthAssetsRootPath = yml.is.oauth.ui['assets-base-url']
  } catch (err) {
    throw err
  }
}

module.exports = {
  stackConfigTask,
  codeCoverageTask,
}
