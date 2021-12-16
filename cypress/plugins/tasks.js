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

const { execSync, exec, spawn } = require('child_process')
const fs = require('fs')
const readline = require('readline')

const { Client } = require('pg')
const yaml = require('js-yaml')
const codeCoverageTask = require('@cypress/code-coverage/task')

const isCI = process.env.CI === 'true' || process.env.CI === '1'
const devDatabase = 'ttn_lorawan_dev'
const seedDatabase = 'tts_seed'

const pgConfig = {
  user: 'root',
  password: 'root',
  host: 'localhost',
  database: 'ttn_lorawan_dev',
  port: 5432,
}
let client

// Spawn a bash session within the postgres container, so we can
// reset the database repeatedly without the overhead of `docker-compose exec`.
const postgresContainer = spawn('docker-compose', [
  '-p',
  'lorawan-stack-dev',
  'exec',
  '-T',
  'postgres',
  '/bin/bash',
])

// `stackConfigTask` sources stack configuration entires to `Cypress` configuration while preserving
// all entries from `cypress.json`.
const stackConfigTask = (_, config) => {
  try {
    const out = execSync(`${isCI ? './' : 'go run ./cmd/'}ttn-lw-stack config --yml`)
    const yml = yaml.load(out)

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
    config.consoleSiteName = yml.console.ui['site-name']
    config.consoleSubTitle = yml.console.ui['sub-title']
    config.consoleTitle = yml.console.ui.title
    config.consoleAssetsRootPath = yml.console.ui['assets-base-url']
    config.consoleRootPath = new URL(yml.console.ui['canonical-url']).pathname

    // Account App.
    config.accountAppSiteName = yml.is.oauth.ui['site-name']
    config.accountAppSubTitle = yml.is.oauth.ui['sub-title']
    config.accountAppTitle = yml.is.oauth.ui.title
    config.accountAppRootPath = new URL(yml.is.oauth.ui['canonical-url']).pathname
    config.accountAppAssetsRootPath = yml.is.oauth.ui['assets-base-url']
  } catch (err) {
    throw err
  }
}

const sqlTask = on => {
  on('task', {
    execSql: async sql => {
      client = new Client(pgConfig)
      await client.connect()
      const res = await client.query(sql)
      client.end()

      return res
    },
    dropAndSeedDatabase: () => {
      postgresContainer.stdin.write(
        `dropdb --if-exists --force ${devDatabase}; psql template1 -c "CREATE DATABASE ${devDatabase} TEMPLATE ${seedDatabase};"\n`,
      )
      return Promise.all([
        new Promise(resolve => exec('tools/bin/mage dev:redisFlush', resolve)),
        new Promise(resolve => {
          postgresContainer.stdout.on('data', resolve)
          postgresContainer.stderr.on('data', data => {
            throw new Error(data)
          })
          postgresContainer.on('close', resolve)
        }),
      ])
    },
  })
}

const stackLogTask = on => {
  on('task', {
    findEmailInStackLog: async (regExp, capturingGroup = 0) => {
      const re = new RegExp(regExp, 'm')
      const rl = readline.createInterface({
        input: fs.createReadStream('.cache/devStack.log', { encoding: 'utf8' }),
        output: process.stdout,
        terminal: false,
      })
      const results = []
      for await (const line of rl) {
        if (line === '') {
          continue
        }
        const parsed = JSON.parse(line)
        if (parsed.msg !== 'Could not send email without email provider') {
          continue
        }
        const match = parsed.body.match(re)
        if (match) {
          results.push(match)
        }
      }

      // Return the most recent occurrence.
      return results ? results.pop()[capturingGroup] : undefined
    },
  })
}

const fileExistsTask = on => {
  on('task', {
    fileExists: filename => {
      if (fs.existsSync(filename)) {
        return fs.readFileSync(filename)
      }

      return false
    },
  })
}

module.exports = {
  stackConfigTask,
  codeCoverageTask,
  sqlTask,
  stackLogTask,
  fileExistsTask,
}
