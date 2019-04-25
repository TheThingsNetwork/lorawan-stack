// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

/* eslint-disable no-console */
/* eslint-disable import/no-commonjs */

const fs = require('fs')
const api = require('../generated/api.json')

function map (files) {
  const result = {}
  const paramRegex = /{([a-z._]+)}/gm

  for (const file of files) {

    for (const service of file.services) {

      result[service.name] = {}

      for (const method of service.methods) {
        result[service.name][method.name] = { file: file.name, http: []}

        if (method.options && method.options['google.api.http']) {
          for (const rule of method.options['google.api.http'].rules) {
            rule.parameters = []
            let match
            while ((match = paramRegex.exec(rule.pattern)) !== null) {
              rule.parameters.push(match[1])
            }

            rule.method = rule.method.toLowerCase()
            result[service.name][method.name].http.push(rule)
          }
        }
      }
    }
  }

  return result
}


fs.writeFile(`${__dirname}/../generated/api-definition.json`, JSON.stringify(map(api.files), null, 2), function (err) {
  if (err) {
    return console.error(err)
  }
  console.log('File saved.')
})
