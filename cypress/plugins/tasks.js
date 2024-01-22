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

const childProcess = require('child_process')
const fs = require('fs')
const path = require('path')

const util = require('util')
const { Client } = require('pg')
const yaml = require('js-yaml')

const isCI = process.env.CI === 'true' || process.env.CI === '1'

const pgConfig = {
  user: 'root',
  password: 'root',
  host: 'localhost',
  database: 'ttn_lorawan_dev',
  port: 5432,
}

// Sources stack configuration entries to Cypress configuration while preserving all entries from cypress.json.
const stackConfigTask = (_, config) => {
  const out = childProcess.execSync(`${isCI ? '../' : 'go run ../cmd/'}ttn-lw-stack config --yml`)
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
}

const sqlTask = on => {
  on('task', {
    execSql: async sql => {
      const client = new Client(pgConfig)
      await client.connect()
      const res = await client.query(sql)
      client.end()
      return res
    },
    dropAndSeedDatabase: async () => {
      const exec = util.promisify(childProcess.exec)
      await Promise.all([
        exec('tools/bin/mage dev:sqlRestore', { cwd: '..' }),
        exec('tools/bin/mage dev:redisFlush', { cwd: '..' }),
      ])
      return null
    },
  })
}

const emailTask = on => {
  on('task', {
    findInLatestEmail: async (regExp, capturingGroup = 0) => {
      const emailDir = '../.dev/email'
      const re = new RegExp(regExp, 'm')
      const files = fs.readdirSync(emailDir)
      const latestMails = files
        .filter(file => fs.lstatSync(path.join(emailDir, file)).isFile())
        .map(file => ({
          file: path.join(emailDir, file),
          mtime: fs.lstatSync(path.join(emailDir, file)).mtime,
        }))
        .sort((a, b) => b.mtime.getTime() - a.mtime.getTime())

      if (latestMails.length === 0) {
        throw new Error('No emails found')
      }

      const latestMailContent = fs.readFileSync(latestMails[0].file, {
        encoding: 'utf8',
        flag: 'r',
      })
      const res = latestMailContent.match(re)

      if (!res) {
        throw new Error('Could not match regex in last email')
      }

      return res[capturingGroup]
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
  sqlTask,
  fileExistsTask,
  emailTask,
}
