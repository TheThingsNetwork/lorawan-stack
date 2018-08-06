// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import HtmlWebpackPlugin from 'html-webpack-plugin'
import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import AddAssetHtmlPlugin from 'add-asset-html-webpack-plugin'
import CleanWebpackPlugin from 'clean-webpack-plugin'
import ShellPlugin from 'webpack-shell-plugin'

import convert from 'koa-connect'
import history from 'connect-history-api-fallback'
import proxy from 'http-proxy-middleware'

import nib from 'nib'

const {
  CONTEXT = '.',
  CACHE_DIR = '.cache',
  PUBLIC_DIR = 'public',
  NODE_ENV = 'development',
  VERSION = '?.?.?',
  GIT_TAG,
  SUPPORT_LOCALES = 'en',
  DEFAULT_LOCALE = 'en',
} = process.env

const DEV_SERVER_BUILD = process.env.DEV_SERVER_BUILD && process.env.DEV_SERVER_BUILD === 'true'
const ASSETS_ROOT = DEV_SERVER_BUILD ? '/assets' : '{{.Root}}'

const context = path.resolve(CONTEXT)
const production = NODE_ENV === 'production'
const src = path.resolve('.', 'pkg/webui')
const include = [ src ]
const modules = [ path.resolve(context, 'node_modules') ]
const publicPath = ASSETS_ROOT
const publicPathRegExp = publicPath.replace('/', '\\/')
const publicPathScheme = new RegExp(`${publicPathRegExp}\\/.*$`)
const publicPathReplace = new RegExp(`^${publicPathRegExp}`)

const r = SUPPORT_LOCALES.split(',').map(l => new RegExp(l.trim()))

export default {
  context,
  mode: production ? 'production' : 'development',
  externals: [ filterLocales ],
  stats: 'minimal',
  devServer: {
    stats: 'minimal',
  },
  entry: {
    console: [
      './config/root.js',
      './pkg/webui/console.js',
    ],
    oauth: [
      './config/root.js',
      './pkg/webui/oauth.js',
    ],
  },
  output: {
    filename: '[name].[hash].js',
    chunkFilename: '[name].[hash].js',
    path: path.resolve(context, PUBLIC_DIR),
    crossOriginLoading: 'anonymous',
    publicPath,
  },
  optimization: {
    splitChunks: {
      cacheGroups: {
        styles: {
          name: 'styles',
          test: /\.css$/,
          chunks: 'all',
          enforce: true,
        },
      },
    },
    removeAvailableModules: false,
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        loader: 'babel-loader',
        include,
        options: {
          cacheDirectory: path.resolve(context, CACHE_DIR, 'babel'),
          sourceMap: true,
          babelrc: true,
        },
      },
      {
        test: /\.(woff|woff2|ttf|eot|jpg|jpeg|png|svg)$/i,
        loader: 'file-loader',
        options: {
          name: '[name].[hash].[ext]',
        },
      },
      {
        test: /\.(styl|css)$/,
        include,
        use: [
          'css-hot-loader',
          {
            loader: MiniCssExtractPlugin.loader,
            options: {
              publicPath: './',
            },
          },
          {
            loader: 'css-loader',
            options: {
              modules: true,
              minimize: production,
              camelCase: true,
              localIdentName: env({
                production: '[hash:base64:10]',
                development: '[path][local]-[hash:base64:10]',
              }),
            },
          },
          {
            loader: 'stylus-loader',
            options: {
              'import': [
                path.resolve(context, 'pkg/webui/styles/include.styl'),
              ],
              use: [ nib() ],
            },
          },
        ],
      },
      {
        test: /\.css$/,
        use: [
          MiniCssExtractPlugin.loader,
          {
            loader: 'css-loader',
            options: {
              modules: false,
              minimize: production,
            },
          },
        ],
        include: modules,
      },
    ],
  },
  plugins: env({
    all: [
      new webpack.NamedModulesPlugin(),
      new webpack.NamedChunksPlugin(),
      new webpack.EnvironmentPlugin({
        NODE_ENV,
        VERSION: VERSION || GIT_TAG || 'unknown',
      }),
      new webpack.DefinePlugin({
        'process.predefined.DEFAULT_MESSAGES':
          JSON.stringify({
            ...require(`${src}/locales/${DEFAULT_LOCALE}`),
            ...require(`${src}/locales/.backend/${DEFAULT_LOCALE}`),
          }),
      }),
      new HtmlWebpackPlugin({
        chunks: [ 'vendor', 'console' ],
        title: 'console',
        filename: path.resolve(PUBLIC_DIR, 'console.html'),
        showErrors: false,
        template: path.resolve(src, 'index.html'),
        devServerBuild: DEV_SERVER_BUILD,
        minify: {
          html5: true,
          collapseWhitespace: true,
        },
      }),
      new HtmlWebpackPlugin({
        chunks: [ 'vendor', 'oauth' ],
        title: 'oauth',
        filename: path.resolve(PUBLIC_DIR, 'oauth.html'),
        showErrors: false,
        template: path.resolve(src, 'index.html'),
        devServerBuild: DEV_SERVER_BUILD,
        minify: {
          html5: true,
          collapseWhitespace: true,
        },
      }),
      new MiniCssExtractPlugin({
        filename: env({
          development: '[name].css',
          production: '[name].[contenthash].css',
        }),
      }),
      new CleanWebpackPlugin(path.resolve(CONTEXT, PUBLIC_DIR), {
        root: context,
        verbose: false,
        dry: DEV_SERVER_BUILD,
        exclude: env({
          production: [],
          development: [
            'libs.bundle.js',
            'libs.bundle.js.map',
          ],
        }),
      }),
    ],
    development: [
      new webpack.DllReferencePlugin({
        context,
        manifest: require(path.resolve(context, CACHE_DIR, 'dll.json')),
      }),
      new webpack.WatchIgnorePlugin([
        /node_modules/,
        /locales/,
        new RegExp(path.resolve(context, PUBLIC_DIR)),
      ]),
      new AddAssetHtmlPlugin({
        filepath: path.resolve(context, PUBLIC_DIR, 'libs.bundle.js'),
      }),
      new ShellPlugin({
        onBuildExit: [ 'make js.gather-locales' ],
      }),
    ],
  }),
  serve: {
    content: 'public',
    devMiddleware: { stats: 'minimal' },
    hotClient: { allEntries: true },
    add (app, middleware, options) {
      // Add new api routes here, to proxy them
      app.use(convert(proxy([
        '/api',
        '/console/api',
        '/console/oauth',
        '/console/auth',
        '/oauth/api',
        '/oauth/authorize',
        '/oauth/token',
      ],
      { target: 'http://localhost:1885' })))
      app.use(convert(history({
        rewrites: [
          { from: publicPathScheme, to: ({ parsedUrl }) => (`${parsedUrl.pathname.replace(publicPathReplace, '')}`) },
          { from: /\/console(\/[a-zA-Z0-9_/.-]*)?$/, to: '/console.html' },
          { from: /\/oauth(\/[a-zA-Z0-9_/.-]*)?$/, to: '/oauth.html' },
        ],
      })))
    },
  },
}

function filterLocales (context, request, callback) {
  if (context.endsWith('node_modules/intl/locale-data/jsonp')) {
    const supported = r.reduce(function (acc, locale) {
      return acc || locale.test(request)
    }, false)

    if (!supported) {
      return callback(null, `commonjs ${request}`)
    }
  }
  callback()
}

// env selects and merges the environments for the passed object based on NODE_ENV, which
// can have the all, development and production keys.
function env (obj = {}) {
  if (!obj) {
    return obj
  }

  const all = obj.all || {}
  const dev = obj.development || {}
  const prod = obj.production || {}

  if (Array.isArray(all) || Array.isArray(dev) || Array.isArray(prod)) {
    return [
      ...all,
      ...(production ? prod : dev),
    ]
  }

  if (typeof dev !== 'object' || typeof prod !== 'object') {
    return production ? prod : dev
  }

  return {
    ...all,
    ...(production ? prod : dev),
  }
}
