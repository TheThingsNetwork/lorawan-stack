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


/* eslint-env node */
/* eslint-disable import/no-commonjs */
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
require('@babel/register')

const config = require('../webpack.config.babel.js').default
const development = process.env.NODE_ENV !== 'production'

// list of allowed plugins
const allow = [
  MiniCssExtractPlugin,
]

if (development) {
  // TODO: reenable this when the bug has been fixed in https://github.com/webpack/webpack/issues/5478
  // const webpack = require("webpack")
  // allow.push(webpack.DllReferencePlugin)
}

// filter plugins on allowed type
config.plugins = config.plugins.filter(function (plugin) {
  return allow.reduce((ok, klass) => ok || plugin instanceof klass, false)
})

config.module.rules = config.module.rules.filter(function (rule) {
  return rule.loader !== 'babel-loader'
})

module.exports = config
