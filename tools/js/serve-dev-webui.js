// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import { spawn } from 'child_process'

import { select, confirm } from '@inquirer/prompts'

const STAGING_URL = process.env.STAGING_URL || ''

// Base configuration environment variables
const baseConfig = `
export GOFLAGS="--tags=tti"
export NODE_ENV="development"
export TTN_LW_IS_ADMIN_RIGHTS_ALL="true"
export TTN_LW_IS_EMAIL_DIR=".dev/email"
export TTN_LW_IS_EMAIL_PROVIDER="dir"
export TTN_LW_LOG_LEVEL="debug"
export TTN_LW_NOC_ACCESS_EXTENDED="true"
export TTN_LW_PLUGINS_SOURCE="directory"
export TTN_LW_CONSOLE_UI_CANONICAL_URL="http://localhost:8080/console"
`

// Local configuration environment variables
const localConfig = `
export TTN_LW_CONSOLE_OAUTH_AUTHORIZE_URL="http://localhost:8080/oauth/authorize"
export TTN_LW_CONSOLE_OAUTH_LOGOUT_URL="http://localhost:8080/oauth/logout"
export TTN_LW_CONSOLE_OAUTH_TOKEN_URL="http://localhost:8080/oauth/token"
export TTN_LW_CONSOLE_UI_ASSETS_BASE_URL="http://localhost:8080/assets"
export TTN_LW_CONSOLE_UI_AS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_EDTC_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_GCS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_GS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_IS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_JS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_JS_FILE="libs.bundle.js console.js"
export TTN_LW_CONSOLE_UI_NS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_QRG_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_IS_OAUTH_UI_IS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_IS_OAUTH_UI_JS_FILE="libs.bundle.js account.js"
export WEBPACK_DEV_BACKEND_API_PROXY_URL="http://localhost:1885"
export TTN_LW_CONSOLE_UI_ACCOUNT_URL="http://localhost:8080/oauth"
export TTN_LW_IS_EMAIL_NETWORK_IDENTITY_SERVER_URL="http://localhost:8080/oauth"
export TTN_LW_IS_OAUTH_UI_CANONICAL_URL="http://localhost:8080/oauth"
`

// Branding configuration environment variables
const brandingConfig = `
export TTN_LW_CONSOLE_UI_BRANDING_CLUSTER_ID="eu1"
export TTN_LW_CONSOLE_UI_BRANDING_TEXT="Local"
export TTN_LW_CONSOLE_UI_FAIR_USE_POLICY_INFORMATION_URL="https://example.com"
export TTN_LW_CONSOLE_UI_SLA_APPLIES=">99%"
export TTN_LW_CONSOLE_UI_SLA_INFORMATION_URL="http://example.com"
export TTN_LW_CONSOLE_UI_SUPPORT_LINK="http://example.com"
export TTN_LW_CONSOLE_UI_SUPPORT_PLAN_APPLIES="Premium"
export TTN_LW_CONSOLE_UI_SUPPORT_PLAN_INFORMATION_URL="http://example.com"
`

// Staging configuration environment variables
const stagingConfig = `
export TTN_LW_CONSOLE_OAUTH_AUTHORIZE_URL="${STAGING_URL}/oauth/authorize"
export TTN_LW_CONSOLE_OAUTH_CLIENT_ID="localhost-console"
export TTN_LW_CONSOLE_OAUTH_CLIENT_SECRET="console"
export TTN_LW_CONSOLE_OAUTH_LOGOUT_URL="${STAGING_URL}/oauth/logout"
export TTN_LW_CONSOLE_OAUTH_TOKEN_URL="${STAGING_URL}/oauth/token"
export TTN_LW_CONSOLE_UI_AS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_EDTC_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_GCS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_GS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_IS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_JS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_NS_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_QRG_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_NOC_BASE_URL="http://localhost:8080/api/v3"
export TTN_LW_CONSOLE_UI_NOC_URL="${STAGING_URL}/noc"
export WEBPACK_DEV_BACKEND_API_PROXY_URL="${STAGING_URL}"
export TTN_LW_CONSOLE_UI_ACCOUNT_URL="${STAGING_URL}/oauth"
export TTN_LW_IS_OAUTH_UI_CANONICAL_URL="${STAGING_URL}/oauth"
`

let envConfig = ''

// Show all environment variables that begin with TTN_LW and WEBPACK_DEV.
const relevantSetEnvs = Object.entries(process.env).filter(
  ([key]) => key.startsWith('TTN_LW') || key.startsWith('WEBPACK_DEV'),
)

if (relevantSetEnvs.length > 0) {
  console.log('ℹ️  Current relevant environment variables:')
  console.log(relevantSetEnvs.map(([key, value]) => `${key}=${value}`).join('\n'))
  console.log(
    '\n⚠️ Before running this script, make sure that you have no environment or stack configuration is set that could interfere with the setup of this script.\n',
  )
}

;(async () => {
  const environment = await select({
    message: 'Select environment:',
    choices: [
      { name: 'Staging', value: 'staging' },
      { name: 'Local', value: 'local' },
      { name: 'Cypress', value: 'cypress' },
    ],
  })

  if (environment !== 'cypress') {
    const branding = await confirm({
      message: 'Enable branding?',
      default: true,
    })

    envConfig = baseConfig

    if (branding) {
      envConfig += brandingConfig
    }

    if (environment === 'local') {
      envConfig += localConfig
    } else if (environment === 'staging') {
      if (!STAGING_URL) {
        console.error(
          '⛔️ STAGING_URL environment variable is not set. Please set it to the base URL of the staging environment and try again.',
        )
        process.exit(1)
      }
      envConfig += stagingConfig
    }
  }

  const envVars = envConfig
    .split('\n')
    .filter(line => line.trim().startsWith('export'))
    .map(line => {
      const [_, key, value] = line.match(/export (\w+)="(.+)"/)
      return [key, value]
    })

  const env = Object.fromEntries(envVars)

  console.log('Environment variables:')
  console.log(envConfig)
  console.log('Starting dev stack and webpack dev server…')

  // Helper function to handle streaming output
  const streamOutput = (prefix, stream) => {
    stream.on('data', data => {
      process.stdout.write(`[${prefix}] ${data}`)
    })

    stream.on('error', data => {
      process.stderr.write(`[${prefix}] ${data}`)
    })
  }

  // Spawn 'go run ./cmd/ttn-lw-stack start'
  const goProcess = spawn('go', ['run', './cmd/ttn-lw-stack', 'start'], {
    env: { ...process.env, ...env },
    shell: true,
  })

  streamOutput('stack', goProcess.stdout)
  streamOutput('stack', goProcess.stderr)

  goProcess.on('close', code => {
    console.log(`[stack] process exited with code ${code}`)
  })

  // Spawn 'tools/bin/mage js:serve'
  const mageProcess = spawn('tools/bin/mage', ['js:serve'], {
    env: { ...process.env, ...env },
    shell: true,
  })

  streamOutput('js:serve', mageProcess.stdout)
  streamOutput('js:serve', mageProcess.stderr)

  mageProcess.on('close', code => {
    console.log(`[js:serve] process exited with code ${code}`)
  })
})()
