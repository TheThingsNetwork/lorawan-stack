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

import path from "path"
import webpack from "webpack"

import HTMLPlugin from "html-webpack-plugin"
import SRIPlugin from "webpack-subresource-integrity"
import AddAssetHtmlPlugin from "add-asset-html-webpack-plugin"
import ExtractTextPlugin from "extract-text-webpack-plugin"
import CleanPlugin from "clean-webpack-plugin"

const {
  CONTEXT = ".",
  CACHE_DIR = ".cache",
  PUBLIC_DIR = "public",
  NODE_ENV = "development",
  VERSION = "?.?.?",
  GIT_TAG,
  STORYBOOK = "0",
  SUPPORT_LOCALES = "en",
} = process.env

const context = path.resolve(CONTEXT)
const production = NODE_ENV === "production"
const src = path.resolve(context, "pkg/webui")
const include = [ src ]
const modules = [ path.resolve(context, "node_modules") ]

const extractCSS = new ExtractTextPlugin({
  filename: "[name].[contenthash].css",
  disable: STORYBOOK === "1",
})

const r = SUPPORT_LOCALES.split(",").map(l => new RegExp(l.trim()))
const filterLocales = function (context, request, callback) {
  if (context.endsWith("node_modules/intl/locale-data/jsonp")) {
    const supported = r.reduce(function (acc, locale) {
      return acc || locale.test(request)
    }, false)

    if (!supported) {
      return callback(null, `commonjs ${request}`)
    }
  }
  callback()
}

export default {
  context,
  target: "web",
  externals: [ filterLocales ],
  cache: !production,
  recordsPath: path.resolve(context, ".cache", "_records"),
  entry: {
    console: [
      "./config/root.js",
      "./pkg/webui/console.js",
    ],
    oauth: [
      "./config/root.js",
      "./pkg/webui/oauth.js",
    ],
  },
  output: {
    filename: "[name].[chunkhash].js",
    chunkFilename: "[name].[chunkhash].js",
    path: path.resolve(context, PUBLIC_DIR),
    publicPath: "{{.Root}}/",
    crossOriginLoading: "anonymous",
  },
  stats: {
    version: false,
    chunks: false,
    modules: false,
    children: false,
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        loader: "babel-loader",
        include,
        options: {
          cacheDirectory: path.resolve(context, CACHE_DIR, "babel"),
          sourceMap: true,
          babelrc: true,
        },
      },
      {
        test: /\.(css|styl)$/,
        loader: extractCSS.extract({
          fallback: "style-loader",
          use: [
            {
              loader: "css-loader",
              options: {
                modules: true,
                minimize: production,
                localIdentName: env({
                  production: "[hash:base64:10]",
                  development: "[path][local]-[hash:base64:10]",
                }),
              },
            },
            {
              loader: "stylus-loader",
              options: {
                "import": [
                  path.resolve(context, "pkg/webui/include.styl"),
                ],
              },
            },
          ],
        }),
        include,
      },
      {
        test: /\.css$/,
        loader: extractCSS.extract({
          fallback: "style-loader",
          use: [
            {
              loader: "css-loader",
              options: {
                modules: false,
                minimize: production,
              },
            },
          ],
        }),
        include: modules,
      },
      {
        test: /\.json$/,
        loader: "json-loader",
        include,
      },
      {
        test: /\.(woff|woff2|ttf|eot|jpg|jpeg|png|svg)$/i,
        loader: "file-loader",
        options: {
          name: "[name].[hash].[ext]",
        },
      },
    ],
  },
  plugins: env({
    all: [
      new webpack.NamedModulesPlugin(),
      new webpack.NamedChunksPlugin(),
      new webpack.EnvironmentPlugin({
        NODE_ENV,
        VERSION: VERSION || GIT_TAG || "unknown",
      }),
      new webpack.LoaderOptionsPlugin({
        context,
        minimize: production,
        debug: !production,
      }),
      new webpack.SourceMapDevToolPlugin({
        module: production,
        columns: production,
        filename: "[file].map",
        exclude: [
          /locale\..+\./,
          /lang\..+-json/,
          /vendor\..+\.js/,
          /runtime\..+\.js/,
        ],
      }),
      extractCSS,
      new SRIPlugin({
        enabled: production,
        hashFuncNames: [
          "sha512",
        ],
      }),
      new HTMLPlugin({
        chunks: [ "vendor", "console" ],
        filename: path.resolve(PUBLIC_DIR, "console.html"),
        showErrors: false,
        template: path.resolve(src, "index.html"),
        minify: {
          html5: true,
          collapseWhitespace: true,
        },
      }),
      new HTMLPlugin({
        chunks: [ "vendor", "oauth" ],
        filename: path.resolve(PUBLIC_DIR, "oauth.html"),
        showErrors: false,
        template: path.resolve(src, "index.html"),
        minify: {
          html5: true,
          collapseWhitespace: true,
        },
      }),
      new CleanPlugin([
        path.resolve(CONTEXT, PUBLIC_DIR),
      ], {
        root: context,
        verbose: false,
        exclude: env({
          production: [],
          development: [
            "libs.bundle.js",
            "libs.bundle.js.map",
          ],
        }),
      }),
    ],
    production: [
      new webpack.optimize.OccurrenceOrderPlugin(true),
      new webpack.optimize.UglifyJsPlugin({
        minimize: true,
        parallel: true,
        sourceMap: true,
        compress: {
          warnings: false,
        },
      }),
      new webpack.optimize.CommonsChunkPlugin({
        name: "vendor",
        minChunks (module) {
          return module.resource && /node_modules/.test(module.resource) && !/node_modules\/ui/.test(module.resource)
        },
      }),
      new webpack.optimize.CommonsChunkPlugin({
        name: "runtime",
      }),
    ],
    development: [
      new webpack.DllReferencePlugin({
        context,
        manifest: require(path.resolve(context, CACHE_DIR, "dll.json")),
      }),
      new webpack.WatchIgnorePlugin([
        /node_modules/,
        new RegExp(path.resolve(context, PUBLIC_DIR)),
      ]),
      new AddAssetHtmlPlugin({
        filepath: path.resolve(context, PUBLIC_DIR, "libs.bundle.js"),
      }),
    ],
  }),
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

  if (typeof dev !== "object" || typeof prod !== "object") {
    return production ? prod : dev
  }

  return {
    ...all,
    ...(production ? prod : dev),
  }
}
