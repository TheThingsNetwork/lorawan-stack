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

const { default: bundleConfig, styleConfig } = require('../webpack.config.babel.js')

// list of allowed plugins
const allow = [MiniCssExtractPlugin]

module.exports = async function({ config, mode }) {
  if (mode === 'PRODUCTION') {
    const webpack = require('webpack')
    allow.push(webpack.DllReferencePlugin)
  }

  // Filter plugins on allowed type
  const filteredPlugins = bundleConfig.plugins.filter(function(plugin) {
    return allow.reduce((ok, klass) => ok || plugin instanceof klass, false)
  })

  // Compose storybook config, making use of stack webpack config
  const cfg = {
    ...config,
    output: {
      ...config.output,
      publicPath: '',
    },
    module: {
      rules: [...config.module.rules, styleConfig],
    },
    plugins: [...config.plugins, ...filteredPlugins],
  }

  return cfg
}
