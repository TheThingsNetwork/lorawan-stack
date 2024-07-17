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
import fs from 'fs'
import path from 'path'

const MiniCssExtractPlugin = require('mini-css-extract-plugin')
require('@babel/register')

const { default: bundleConfig, styleConfig } = require('../webpack.config.babel')

// List of allowed plugins.
const allow = [MiniCssExtractPlugin]

const { CONTEXT = '.' } = process.env
const context = path.resolve(CONTEXT)
const supportedLocales = fs
  .readdirSync(path.resolve(context, 'pkg/webui/locales'))
  .filter(fn => fn.endsWith('.json'))
  .map(fn => fn.split('.')[0])

module.exports = async ({ config, mode }) => {
  const webpack = require('webpack')
  if (mode === 'PRODUCTION') {
    allow.push(webpack.DllReferencePlugin)
  }

  // Filter plugins on allowed type.
  const filteredPlugins = bundleConfig.plugins.filter(plugin =>
    allow.reduce((ok, klass) => ok || plugin instanceof klass, false),
  )

  // Compose storybook config, making use of stack webpack config.
  const cfg = {
    ...config,
    resolve: {
      fallback: { crypto: false },
      alias: {
        ...config.resolve.alias,
        ...bundleConfig.resolve.alias,
      },
    },
    output: {
      ...config.output,
      publicPath: '',
    },
    module: {
      rules: [
        ...config.module.rules,
        {
          test: /\.jsx?$/,
          exclude: /node_modules/,
          use: 'babel-loader',
        },
        styleConfig,
      ],
    },
    plugins: [
      ...config.plugins,
      ...filteredPlugins,
      new webpack.DefinePlugin({
        'process.predefined.SUPPORTED_LOCALES': JSON.stringify(supportedLocales),
      }),
    ],
  }

  return cfg
}
