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

import fs from 'fs'
import path from 'path'
import child_process from 'child_process'

import webpack from 'webpack'
import HtmlWebpackPlugin from 'html-webpack-plugin'
import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import AddAssetHtmlPlugin from 'add-asset-html-webpack-plugin'
import { CleanWebpackPlugin } from 'clean-webpack-plugin'
import ShellPlugin from 'webpack-shell-plugin'
import CopyWebpackPlugin from 'copy-webpack-plugin'
import ReactRefreshWebpackPlugin from '@pmmmwh/react-refresh-webpack-plugin'
import nib from 'nib'

import pjson from '../package.json'

const { version } = pjson
const revision =
  child_process.execSync('git rev-parse --short HEAD').toString().trim() || 'unknown revision'

const {
  CONTEXT = '.',
  CACHE_DIR = '.cache',
  PUBLIC_DIR = 'public',
  NODE_ENV = 'production',
  MAGE = 'tools/bin/mage',
} = process.env

const WEBPACK_IS_DEV_SERVER_BUILD = process.env.WEBPACK_IS_DEV_SERVER_BUILD === 'true'
const WEBPACK_DEV_SERVER_DISABLE_HMR = process.env.WEBPACK_DEV_SERVER_DISABLE_HMR === 'true'
const WEBPACK_DEV_SERVER_USE_TLS = process.env.WEBPACK_DEV_SERVER_USE_TLS === 'true'
const WEBPACK_GENERATE_PRODUCTION_SOURCEMAPS =
  process.env.WEBPACK_GENERATE_PRODUCTION_SOURCEMAPS === 'true'
const TTN_LW_TLS_CERTIFICATE = process.env.TTN_LW_TLS_CERTIFICATE || './cert.pem'
const TTN_LW_TLS_KEY = process.env.TTN_LW_TLS_KEY || './key.pem'
const TTN_LW_TLS_ROOT_CA = process.env.TTN_LW_TLS_ROOT_CA || './ca.pem'

const ASSETS_ROOT = '/assets'

const context = path.resolve(CONTEXT)
const production = NODE_ENV !== 'development'

const src = path.resolve('.', 'pkg/webui')
const include = [src]
const modules = [path.resolve(context, 'node_modules')]

const supportedLocales = fs
  .readdirSync(path.resolve(context, 'pkg/webui/locales'))
  .filter(fn => fn.endsWith('.json'))
  .map(fn => fn.split('.')[0])

const r = supportedLocales.map(l => new RegExp(`./${l.trim()}`))

const filterLocales = (context, request, callback) => {
  if (
    context.endsWith('node_modules/intl/locale-data/jsonp') ||
    context.endsWith('node_modules/@formatjs/intl-relativetimeformat/locale-data')
  ) {
    const supported = r.reduce((acc, locale) => acc || locale.test(request), false)

    if (!supported) {
      return callback(null, `commonjs ${request}`)
    }
  }
  callback()
}

// Env selects and merges the environments for the passed object based on
// `NODE_ENV`, which can have the all, development and production keys.
const env = (obj = {}) => {
  if (!obj) {
    return obj
  }

  const all = obj.all
  const dev = obj.development
  const prod = obj.production

  if (Array.isArray(all) || Array.isArray(dev) || Array.isArray(prod)) {
    return [...(all || []), ...(production ? prod || [] : dev || [])]
  }

  if (
    (dev !== undefined && typeof dev !== 'object') ||
    (prod !== undefined && typeof prod !== 'object')
  ) {
    return production ? prod : dev
  }

  return {
    ...(all || {}),
    ...(production ? prod || {} : dev || {}),
  }
}

export const styleConfig = {
  test: /\.(styl|css)$/,
  include,
  use: [
    {
      loader: MiniCssExtractPlugin.loader,
      options: {
        publicPath: './',
      },
    },
    {
      loader: 'css-loader',
      options: {
        modules: {
          exportLocalsConvention: 'camelCase',
          localIdentName: env({
            production: '[hash:base64:10]',
            development: '[local]-[hash:base64:4]',
          }),
        },
      },
    },
    {
      loader: 'stylus-loader',
      options: {
        stylusOptions: {
          import: [path.resolve(context, 'pkg/webui/styles/include.styl')],
          use: nib(),
        },
      },
    },
  ],
}

export default {
  context,
  mode: production ? 'production' : 'development',
  externals: [filterLocales],
  stats: 'minimal',
  target: 'web',
  devtool: production ? false : 'eval-source-map',
  resolve: {
    alias: env({
      all: {
        '@ttn-lw': path.resolve(context, 'pkg/webui'),
        '@console': path.resolve(context, 'pkg/webui/console'),
        '@account': path.resolve(context, 'pkg/webui/account'),
        '@assets': path.resolve(context, 'pkg/webui/assets'),
      },
      development: {
        'ttn-lw': path.resolve(context, 'sdk/js/src'),
      },
    }),
  },
  devServer: {
    devMiddleware: {
      publicPath: `${ASSETS_ROOT}/`,
    },
    port: 8080,
    hot: !WEBPACK_DEV_SERVER_DISABLE_HMR,
    proxy: [
      {
        context: ['/console', '/account', '/oauth', '/api', '/assets/blob'],
        target: WEBPACK_DEV_SERVER_USE_TLS ? 'https://localhost:8885' : 'http://localhost:1885',
        changeOrigin: true,
        secure: false,
        ws: true,
      },
    ],
    historyApiFallback: true,
    ...(WEBPACK_DEV_SERVER_USE_TLS
      ? {
          https: {
            cert: fs.readFileSync(TTN_LW_TLS_CERTIFICATE),
            key: fs.readFileSync(TTN_LW_TLS_KEY),
            ca: fs.readFileSync(TTN_LW_TLS_ROOT_CA),
          },
        }
      : {}),
  },
  entry: {
    console: ['./config/root.js', './pkg/webui/console.js'],
    account: ['./config/root.js', './pkg/webui/account.js'],
  },
  output: {
    filename: production ? '[name].[chunkhash].js' : '[name].js',
    chunkFilename: production ? '[name].[chunkhash].js' : '[name].js',
    path: path.resolve(context, PUBLIC_DIR),
    crossOriginLoading: 'anonymous',
    publicPath: ASSETS_ROOT,
  },
  optimization: {
    splitChunks: {
      cacheGroups: {
        styles: {
          name: 'styles',
          test: /\.css$/,
          chunks: 'all',
        },
      },
    },
    removeEmptyChunks: false,
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        loader: 'babel-loader',
        include,
        options: {
          cacheDirectory: path.resolve(context, CACHE_DIR, 'babel'),
          sourceMap: true,
          babelrc: true,
          plugins: [
            !WEBPACK_DEV_SERVER_DISABLE_HMR &&
              !production &&
              require.resolve('react-refresh/babel'),
          ].filter(Boolean),
        },
      },
      {
        test: /\.(woff|woff2|ttf|eot|jpg|jpeg|png|svg)$/i,
        type: 'asset/resource',
        generator: {
          filename: '[name].[contenthash:20][ext]',
        },
      },
      styleConfig,
      {
        test: /\.css$/,
        use: [
          MiniCssExtractPlugin.loader,
          {
            loader: 'css-loader',
            options: {
              modules: false,
            },
          },
        ],
        include: modules,
      },
    ],
  },
  plugins: env({
    all: [
      new webpack.EnvironmentPlugin({
        NODE_ENV,
        VERSION: version,
        REVISION: revision,
      }),
      new webpack.DefinePlugin({
        'process.predefined.SUPPORTED_LOCALES': JSON.stringify(supportedLocales),
      }),
      new HtmlWebpackPlugin({
        inject: false,
        filename: `manifest.yaml`,
        showErrors: false,
        template: path.resolve('config', 'manifest-template.yaml'),
        minify: false,
      }),
      new MiniCssExtractPlugin({
        filename: env({
          development: '[name].css',
          production: '[name].[contenthash].css',
        }),
      }),
      new CleanWebpackPlugin({
        dry: WEBPACK_IS_DEV_SERVER_BUILD,
        verbose: false,
        cleanOnceBeforeBuildPatterns: env({
          all: ['**/*'],
          development: ['!libs.bundle.js', '!libs.bundle.js.map'],
          production: ['!libs.*.bundle.js', '!libs.*.bundle.js.map'],
        }),
      }),
      new CopyWebpackPlugin({
        patterns: [{ from: `${src}/assets/static` }],
      }),
      new webpack.DllReferencePlugin({
        context,
        manifest: path.resolve(context, CACHE_DIR, 'dll.json'),
      }),
      new AddAssetHtmlPlugin({
        glob: path.resolve(context, PUBLIC_DIR, 'libs*bundle.js'),
      }),
    ],
    production: [
      ...(WEBPACK_GENERATE_PRODUCTION_SOURCEMAPS
        ? [
            new webpack.SourceMapDevToolPlugin({
              filename: '[file].map',
              exclude: /^(?!(console|oauth).*$).*/,
            }),
          ]
        : []),
    ],
    development: [
      ...(!WEBPACK_DEV_SERVER_DISABLE_HMR ? [new ReactRefreshWebpackPlugin()] : []),
      new webpack.WatchIgnorePlugin({
        paths: [/node_modules/, /locales/, path.resolve(context, PUBLIC_DIR)],
      }),
      new ShellPlugin({
        onBuildExit: [`${MAGE} js:extractLocaleFiles`],
      }),
    ],
  }),
}
