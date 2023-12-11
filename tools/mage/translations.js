// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
/* eslint-disable no-alert, no-console */
/* eslint-disable import/no-commonjs */

const fs = require('fs')
const path = require('path')

const yargs = require('yargs')
const { mkdirp } = require('mkdirp')
const g = require('glob')

const argv = yargs.argv
const env = process.env

/**
 * Flatten a string or list of strings into a list of strings,
 * split by , and trimmed.
 *
 * @param {string|Array<string>} s - The string or array to flatten.
 * @returns {Array<string>} - A flat array of strings.
 */
const flatten = s => {
  if (!s) {
    return []
  }

  if (Array.isArray(s)) {
    return s.map(flatten).reduce((a, n) => [...a, ...n], [])
  }

  return s.split(',').map(s => s.trim())
}

const localesDir = argv.locales || env.LOCALES_DIR || 'pkg/webui/locales'
const support = flatten(argv.support || env.SUPPORT_LOCALES || 'en,ja')
const backendOnly = ('backend-only' in argv && argv['backend-only'] !== false) || false
const messagesDir = !backendOnly && (argv.messages || env.MESSAGES || '.cache/messages')
const backendMessages = argv.backendMessages || env.MESSAGES_BACKEND || 'config/messages.json'
const defaultLocale = argv.default || env.DEFAULT_LOCALE || 'en'
const exportForTranslation = argv['export-for-interpreters'] || env.EXPORT_FOR_INTERPRETERS || false
const verbose = ('verbose' in argv && argv.verbose !== 'false') || false
const ignoreDuplicates =
  ('ignore-duplicates' in argv && argv['ignore-duplicates'] !== 'false') || false

if (argv.help) {
  console.log(`Usage: translations [opts]

  Gathers all translations into a single file and generates locale files based on them.

  This program does a couple of things:

    1. It gathers all the messages defined in the source files (as output by babel-plugin-react-intl).
    2. It incorporates updates from the updates file.
    3. It writes the updated locales into one file per locale.

  Supported file types are .json, .yml and .xlsx.

Options:

  --support <locale>        list of supported locales (default: en) [$SUPPORT_LOCALES]
  --default <locale>        the default locale (that will inherit the default message) (default: en) [$DEFAULT_LOCALE]
  --locales <dir>           directory where the locales are stored (default: ./locales/) [$LOCALES_DIR]
  --backend-messages <file> file where the backend messages are stored (default config/messages.json) [$BACKEND_MESSAGES]
  --messages <file>         directory where the messages are extracted to by react-intl (default: .cache/message) [$MESSAGES]
  --backend-only <flag>.    Flag that determines whether only backend messages will be processed
  --verbose                 verbose output for debugging purposes
  --export-for-translation  Flag that determines whether the messages will be exported for interpreters
  --ignore-duplicates       Flag that determines whether the program will ignore duplicate messages
  --help                    show this help message
`)
}

/**
 * Find a list of files based on a glob pattern.
 *
 * @param {string} pat - The glob pattern, eg. "./foo/*.js".
 * @returns {Promise<Array<string>>} - A promise that resolves to an array of
 * filenames that match the pattern.
 */
const glob = pat =>
  new Promise((resolve, reject) => {
    g(pat, (err, res) => {
      if (err) {
        return reject(err)
      }
      return resolve(res)
    })
  })

/**
 * Read a file from disk.
 *
 * @param {string} filename - The name of the file to read.
 * @returns {Promise<string>} - The contents of the file.
 */
const read = filename =>
  new Promise((resolve, reject) => {
    if (verbose) {
      console.log(`reading from ${filename}`)
    }
    fs.readFile(filename, (err, res) => {
      if (err) {
        return reject(err)
      }
      return resolve(res)
    })
  })

/**
 * Write to a file.
 *
 * @param {string} filename - The file to write to.
 * @param {string} content - The content to write.
 *
 * @returns {Promise<undefined>} - A promise that resolves when the file has been written.
 */
const write = (filename, content) =>
  new Promise((resolve, reject) => {
    if (verbose) {
      console.log('writing', filename)
    }
    fs.writeFile(filename, content, (err, res) => {
      if (err) {
        return reject(err)
      }
      return resolve(res)
    })
  })

/**
 * Read the locales from the localesDir (specified by --locales) and parse
 * them into an object. Locales that are specified in --support but do not
 * have an corresponding file in the localesDir will be filled in.
 * Locales that are in the localesDir but not in --support will be omitted.
 *
 * @returns {object} - The locales, keyed by locale name.
 * For example: `{ en: { ... }, ja: { ... }}`.
 */
const readLocales = async () => {
  const loc = await Promise.all(
    support.map(async locale => {
      let parsed = {}
      let content
      try {
        content = await read(`${path.resolve(localesDir)}/${locale}.json`)
        parsed = JSON.parse(content)
      } catch (err) {
        if (err instanceof SyntaxError) {
          // Check if there are merge conflicts in the JSON file
          if (content.includes('<<<<<<<') && content.includes('>>>>>>>')) {
            throw new SyntaxError(
              `Error parsing ${locale}.json: merge conflicts found.\n\nPlease resolve merge conflicts before continuing.`,
            )
          }
          throw err
        }
      }

      parsed.__locale = locale

      // Detect duplicate messages in the English locale and store them per message
      // this will cause a non-zero exit code so that this issue can be caught in CI
      if (locale === defaultLocale) {
        const duplicates = Object.entries(parsed)
          .reduce((acc, [id, message]) => {
            const existing = acc.find(d => d.message === message)
            if (existing) {
              existing.ids.push(id)
            } else {
              acc.push({ message, ids: [id] })
            }
            return acc
          }, [])
          .filter(d => d.ids.length > 1)

        if (duplicates.length > 0) {
          // Write duplicates to a file
          await write(
            `${path.resolve(localesDir)}/${locale}-duplicates.json`,
            JSON.stringify(duplicates, null, 2),
          )

          console.warn(
            `Duplicate messages found in the "${locale}" locale. This means that the same message text is used more than once. See ${locale}-duplicates.json for details and resolve the issue before continuing.`,
          )

          if (!ignoreDuplicates) {
            process.exit(1)
          }
        }
      }

      return parsed
    }),
  )

  return loc.reduce((acc, next) => {
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
const readMessages = async () => {
  if (!messagesDir) {
    return {}
  }
  const files = await glob(`${path.resolve(messagesDir)}/**/*.json`)
  return files
    .map(f => fs.readFileSync(f, 'utf-8'))
    .map(c => JSON.parse(c))
    .reduce((acc, next) => [...acc, ...next], [])
    .reduce((acc, next) => {
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
 * Read and parse (and marshal) the backend messages, coming from `tools/bin/mage go:messages`.
 *
 * @returns {object} - The backend messages, keyed by message id.
 */
const readBackendMessages = async () => {
  if (!backendMessages) {
    return {}
  }
  const backend = JSON.parse(await read(`${path.resolve(backendMessages)}`))
  return Object.keys(backend).reduce(
    (acc, id) => ({
      ...acc,
      [id]: {
        id,
        defaultMessage: backend[id].translations[defaultLocale],
        locales: backend[id].translations,
        description: backend[id].description,
      },
    }),
    {},
  )
}

/**
 * Get a nested key in an object or return null if not found.
 *
 * @param {object} object - The object to find the key in.
 * @param {Array<string>} pth - The path to find in the object.
 * @returns {any} - The value of the key at the path, or null if not found.
 */
const get = (object, ...pth) => {
  if (object === null) {
    return null
  }

  if (pth.length === 0) {
    return object
  }

  const [head, ...tail] = pth

  if (head in object) {
    return get(object[head], ...tail)
  }

  return null
}

/**
 * Write locales to their corresponding file in the localesDir (specified by --locales).
 *
 * @param {object} locales - The locales to write.
 * @param {Array} backendIds - The ids of the backend messages.
 * @returns {Promise<undefined>} - A promise that resolves when all locales have been written.
 */
const writeLocales = async (locales, backendIds) =>
  Promise.all(
    Object.keys(locales).map(async key => {
      const locale = locales[key]
      let cleaned = locale
      if (key === defaultLocale && !backendOnly) {
        // Remove backend keys from the default locale
        cleaned = Object.keys(locale).reduce((acc, next) => {
          if (!backendIds.includes(next)) {
            return {
              ...acc,
              [next]: locale[next],
            }
          }
          return acc
        }, {})
      }
      const content = JSON.stringify(cleaned, null, 2).concat('\n')
      await mkdirp(localesDir)
      await write(`${localesDir}/${key}.json`, content)
    }),
  )

/**
 * Write messages out for supported locales as CSV, for use when translating.
 *
 * @param {object} locales - The locales to write.
 * @param {string} targetLanguage - The language (locale) to translate to.
 * @returns {Promise<undefined>} - A promise that resolves when the file has been written.
 */
const writeInterpreterFile = async (locales, targetLanguage) => {
  // Write the data as CSV, the first column being the message id
  // the second column being the English message, and the third being the target language message.

  const locale = locales[targetLanguage]
  const content = Object.keys(locale)
    .filter(id => !locale[id])
    .map(id => {
      const message = locale[id] || ''
      // Add proper escapes
      const english = locales[defaultLocale][id].replace(/"/g, '""')
      const target = message.replace ? message.replace(/"/g, '""') : message
      return `"${id}","${english}","${target}"`
    })
    .join('\n')
  await write(`./${targetLanguage}.csv`, `id,default,${targetLanguage}\n${content}`)
}

// Main function.
const main = async () => {
  const [locales, messages, backend] = await Promise.all([
    readLocales(),
    readMessages(),
    readBackendMessages(),
  ])
  const updated = {}

  // Merge backend messages into messages
  for (const id of Object.keys(backend)) {
    const message = backend[id]
    if (!messages[id]) {
      messages[id] = message
    }
  }

  const backendIds = Object.keys(backend)

  // Walk through messages via id
  for (const id of Object.keys(messages)) {
    const message = messages[id]

    // Create new unified translation object with per locale structure
    for (const locale of support) {
      updated[locale] = updated[locale] || {}

      // If not in there, try to find it in locales
      let msg = get(locales, locale, id) || ''

      // Force default message on default locale
      if (locale === defaultLocale) {
        msg = message.defaultMessage
      }

      updated[locale][id] = msg
      message.locales = message.locales || {}
      message.locales[locale] = msg
    }
  }

  await writeLocales(updated, backendIds)

  if (exportForTranslation) {
    const supportWithoutDefault = support.filter(s => s !== defaultLocale)
    supportWithoutDefault.forEach(async targetLanguage => {
      await writeInterpreterFile(updated, targetLanguage)
    })
  }
  console.log('Locale files updated.')
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})
