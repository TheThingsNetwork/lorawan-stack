// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

/* eslint-env node */

import path from "path"
import webpack from "webpack"

const {
  CONTEXT = ".",
  CACHE_DIR = ".cache",
  PUBLIC_DIR = "public",
} = process.env

const context = path.resolve(CONTEXT)
const library = "[name]_[hash]"

const pkg = require(path.resolve(context, "package.json"))
const libs = Object.keys(pkg.dependencies || {})

export default {
  context,
  target: "web",
  stats: "minimal",
  devtool: "module-source-map",
  recordsPath: path.resolve(context, CACHE_DIR, "_libs_records"),
  entry: { libs },
  output: {
    filename: "[name].bundle.js",
    path: path.resolve(context, PUBLIC_DIR),
    library,
  },
  plugins: [
    new webpack.DllPlugin({
      name: library,
      path: path.resolve(context, CACHE_DIR, "dll.json"),
      library,
    }),
  ],
  performance: {
    hints: false,
  },
  module: {
    loaders: [
      {
        test: /\.(woff|woff2|ttf|eot|jpg|jpeg|png|svg)$/i,
        loader: "file-loader",
        options: {
          name: "[name].[hash].[ext]",
        },
      },
    ],
  },
}
