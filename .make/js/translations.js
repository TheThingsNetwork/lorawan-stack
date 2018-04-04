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

/* global process */

import fs from "fs"
import path from "path"
import yargs from "yargs"
import g from "glob"
import YAML from "js-yaml"
import xlsx from "./xls"
import xx from "./xx"

const argv = yargs.argv
const env = process.env

/**
 * flatten a string or list of strings into a list of strings,
 * split by , and trimmed.
 *
 * @param {string|Array<string>} s - The string or array to flatten.
 * @returns {Array<string>} - A flat array of strings.
 */
const flatten = function (s) {
  if (!s) {
    return []
  }

  if (Array.isArray(s)) {
    return s.map(flatten).reduce((a, n) => [ ...a, ...n ], [])
  }

  return s.split(",").map(s => s.trim())
}

const localesDir = argv.locales || env.LOCALES_DIR || "pkg/webui/locales"
const support = flatten(argv.support || env.SUPPORT_LOCALES || "en")
const messagesDir = argv.messages || env.MESSAGES_DIR || ".cache/messages"
const defaultLocale = argv.default || env.DEFAULT_LOCALE || "en"
const output = flatten(argv.output || env.OUTPUT || "messages.yml")
const updates = flatten(argv.updates || env.UPDATES)

if (argv.help) {
  console.log(`Usage: translations [opts]

  Gathers all translations into a single file and generates locale files based on them.

  This program does a couple of things:

    1. It gathers all the messages defined in the source files (as output by babel-plugin-react-intl).
    2. It incorporates updates from the updates file.
    3. It writes the messages to all the files specified by --output
    4. It writes the updated locales into one file per locale.

  Supported file types are .json, .yml and .xlsx.

Options:

  --support <locale>   list of supported locales (default: en) [$SUPPORT_LOCALES]
  --default <locale>   the default locale (that will inherit the default message) (default: en) [$DEFAULT_LOCALE]
  --locales <dir>      directory where the locales are stored (default: ./locales/) [$LOCALES_DIR]
  --messages <file>    directory where the messages are extracted to by react-intl (default: .cache/message) [$MESSAGES_DIR]
  --output <file...>   location of the output files with all messages (default: ./messages.yml) [$OUTPUT]
  --updates <file...>  location of an updated output file that will be merged into the current output file (updates are applied in order, latter files will overwrite previous ones). [$UPDATES]
  --help               show this help message
`)
}

/**
 * Find a list of files based on a glob pattern.
 *
 * @param {string} pat - THe glob pattern, eg. "./foo/*.js"
 * @returns {Promise<Array<string>>} - A promise that resolves to an array of
 *   filenames that match the pattern.
 */
const glob = function (pat) {
  return new Promise(function (resolve, reject) {
    g(pat, function (err, res) {
      if (err) {
        return reject(err)
      }
      return resolve(res)
    })
  })
}

/**
 * Read a file from disk.
 *
 * @param {string} filename - The name of the file to read.
 * @returns {Promise<string>} - The contents of the file.
 */
const read = function (filename) {
  return new Promise(function (resolve, reject) {
    console.log(`reading from ${filename}`)
    fs.readFile(filename, function (err, res) {
      if (err) {
        return reject(err)
      }
      return resolve(res)
    })
  })
}

/**
 * Write to a file.
 *
 * @param {string} filename - The file to write to.
 * @param {string} content - The content to write.
 *
 * @returns {Promise<void>} - A promise that resolves when the file has been written.
 */
const write = function (filename, content) {
  return new Promise(function (resolve, reject) {
    console.log("writing", filename)
    fs.writeFile(filename, content, function (err, res) {
      if (err) {
        return reject(err)
      }
      return resolve(res)
    })
  })
}

/**
 * Read the locales from the localesDir (specified by --locales) and parse
 * them into an object. Locales that are specified in --support but do not
 * have an corresponding file in the localesDir will be filled in.
 * Locales that are in the localesDir but not in --support will be omitted.
 *
 * @returns {object} - The locales, keyed by locale name.
 *   For example: `{ en: { ... }, ja: { ... }}`
 */
const readLocales = async function () {
  const loc = await Promise.all(support.map(async function (locale) {
    let parsed = {}
    try {
      const content = await read(`${path.resolve(localesDir)}/${locale}.json`)
      parsed = JSON.parse(content)
    } catch (err) {}

    parsed.__locale = locale

    return parsed
  }))


  return loc.reduce(function (acc, next) {
    const locale = next.__locale
    delete next.__locale

    return {
      ...acc,
      [locale]: next,
    }
  }, {})
}

/**
 * Read and parse messages that were exported by babel-plugin-react-intl, located in
 * messagesDir (specified by --messages).
 *
 * @returns {object} - The messages, keyed by message id.
 */
const readMessages = async function () {
  const files = await glob(`${path.resolve(messagesDir)}/**/*.json`)
  return files
    .map(f => fs.readFileSync(f, "utf-8"))
    .map(c => JSON.parse(c))
    .reduce(function (acc, next) {
      return [ ...acc, ...next ]
    }, [])
    .reduce(function (acc, next) {
      if (next.id in acc) {
        console.warn(`message id ${next.id} seen multiple times`)
      }

      return {
        ...acc,
        [next.id]: next,
      }
    }, {})
}

/**
 * Read and parse the updates from the updates file (if given) (specified by --updates).
 *
 * @returns {object} - The updates, keyed by message id.
 */
const readUpdates = async function () {
  const contents = await updates.reduce(async function (acc, file) {
    try {
      const content = await read(file)
      const m = marshaller(file)
      return [
        ...(await acc),
        m.unmarshal(content),
      ]
    } catch (err) {
      return acc
    }
  }, [])

  return contents.reduce(function (acc, updates) {
    const res = { ...acc, ...updates }
    for (const id of Object.keys(acc)) {
      res[id].force = acc[id] && acc[id].force
    }

    return res
  }, {})
}

/**
 * get a nested key in an object or return null if not found.
 *
 * @param {object} object - The object to find the key in.
 * @param {Array<string>} pth - The path to find in the object.
 * @returns {any} - The value of the key at the path, or null if not found.
 */
const get = function (object, ...pth) {
  if (object === null) {
    return null
  }

  if (pth.length === 0) {
    return object
  }

  const [ head, ...tail ] = pth

  if (head in object) {
    return get(object[head], ...tail)
  }

  return null
}

/**
 * removes message ids (inline).
 *
 * @param {object} messages - The messages.
 * @returns {object} - The messages but with their id removed.
 */
const removeID = function (messages) {
  const res = {}
  for (const id of Object.keys(messages)) {
    const { id: _, ...rest } = messages[id]
    res[id] = rest
  }

  return res
}

/**
 * Add message ids based on keys in the map (inline).
 *
 * @param {object} messages - The messages to add the ids to.
 * @returns {object} - The messages with the added ids.
 */
const addID = function (messages) {
  const res = {}
  for (const id of Object.keys(messages)) {
    res[id] = { ...messages[id]}
    res[id].id = id
  }

  return res
}


/**
 * Create a marshaller based on the filename of the source.
 * If the extension matches yaml, returns a YAML marshaller.
 * If the extension matches json, returns a JSON marshaller.
 *
 * @param {string} filename - The filename to match.
 * @returns {object} - An object containing the marshal and unmarshal
 *   functions for the matched file type.
 */
const marshaller = function (filename) {
  if (!filename) {
    throw new Error("Illegal filename")
  }

  if (filename.endsWith(".json")) {
    return {
      unmarshal: buf => addID(JSON.parse(buf.toString())),
      marshal: o => JSON.stringify(removeID(o), null, 2),
    }
  }

  if (filename.endsWith(".yml") || filename.endsWith(".yaml")) {
    return {
      unmarshal: buf => addID(YAML.safeLoad(buf.toString())),
      marshal: o => YAML.safeDump(removeID(o)),
    }
  }

  if (filename.endsWith("xlsx")) {
    return xlsx
  }
}

/**
 * Write messages to the message file, determined by --output.
 *
 * @param {object} messages - The messages to write.
 * @returns {Promise<void>} - A promise that resolves when the messages are written.
 */
const writeMessages = function (messages) {
  return Promise.all(output.map(function (file) {
    const m = marshaller(file)
    const content = m.marshal(messages)
    return write(file, content)
  }))
}

/**
 * Write locales to their corresponding file in the localesDir (specified by --locales).
 *
 * @param {object} locales - The locales to write.
 * @returns {Promise<void>} - A promise that resolves when all locales have been written.
 */
const writeLocales = async function (locales) {
  return Promise.all(Object.keys(locales).map(async function (key) {
    const locale = locales[key]
    const content = JSON.stringify(locale, null, 2).concat("\n")
    await write(`${localesDir}/${key}.json`, content)
  }))
}

// main function.
const main = async function () {
  const [ locales, messages, updates ] = await Promise.all([ readLocales(), readMessages(), readUpdates() ])

  const updated = {}

  for (const id of Object.keys(updates)) {
    const message = updates[id]
    if (message.force && !messages[id]) {
      messages[id] = message
    }
  }

  for (const id of Object.keys(messages)) {
    const message = messages[id]
    for (const locale of support) {
      updated[locale] = updated[locale] || {}

      let msg = get(updates, id, "locales", locale)
      if (msg === null) {
        msg = get(locales, locale, id) || ""
      }

      // force default message on default locale
      if (locale === defaultLocale) {
        msg = message.defaultMessage
      }

      updated[locale][id] = msg
      message.locales = message.locales || {}
      message.locales[locale] = msg
    }

    // set xx to default locale replaced by x'es
    const x = xx(updated[defaultLocale][message.id])
    updated.xx = updated.xx || {}
    updated.xx[message.id] = x
  }

  await Promise.all([
    writeMessages(messages),
    writeLocales(updated),
  ])
}

main().catch(function (err) {
  console.error(err)
  process.exit(1)
})
