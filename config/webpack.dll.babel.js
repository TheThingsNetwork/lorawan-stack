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

import path from 'path'

import webpack from 'webpack'

const { CONTEXT = '.', CACHE_DIR = '.cache', PUBLIC_DIR = 'public' } = process.env
const mode = process.env.NODE_ENV === 'development' ? 'development' : 'production'
const WEBPACK_GENERATE_PRODUCTION_SOURCEMAPS =
  process.env.WEBPACK_GENERATE_PRODUCTION_SOURCEMAPS === 'true'

const context = path.resolve(CONTEXT)
const library = '[name]_[fullhash]'

const pkg = require(path.resolve(context, 'package.json'))
const excludeLibs = ['react-hot-loader', 'ttn-lw']
const libs = Object.keys(pkg.dependencies || {}).filter(lib => !excludeLibs.includes(lib))
const devtool =
  (mode === 'production' && WEBPACK_GENERATE_PRODUCTION_SOURCEMAPS) || mode === 'development'
    ? 'source-map'
    : false

export default {
  context,
  mode,
  target: 'web',
  stats: 'minimal',
  devtool,
  recordsPath: path.resolve(context, CACHE_DIR, '_libs_records'),
  entry: { libs },
  output: {
    filename: mode === 'production' ? '[name].[fullhash].bundle.js' : '[name].bundle.js',
    hashDigest: 'hex',
    hashDigestLength: 20,
    path: path.resolve(context, PUBLIC_DIR),
    library,
  },
  plugins: [
    new webpack.DllPlugin({
      name: library,
      path: path.resolve(context, CACHE_DIR, 'dll.json'),
    }),
  ],
  performance: {
    hints: false,
  },
  module: {
    rules: [
      {
        test: /\.(woff|woff2|ttf|eot|jpg|jpeg|png|svg)$/i,
        use: [
          {
            loader: 'file-loader',
            options: {
              name: '[name].[contenthash:20].[ext]',
            },
          },
        ],
      },
    ],
  },
}
