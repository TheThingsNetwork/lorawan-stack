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

/* eslint-disable capitalized-comments */

import React, { useState, useCallback, useEffect, createContext, useMemo } from 'react'
import { IntlProvider, defineMessages, ReactIntlErrorCode } from 'react-intl'
import { uniq } from 'lodash'

import Overlay from '@ttn-lw/components/overlay'
import Spinner from '@ttn-lw/components/spinner'

import PropTypes from '@ttn-lw/lib/prop-types'
import log, { error } from '@ttn-lw/lib/log'
import isDevelopment from '@ttn-lw/lib/dev'
import { ingestError } from '@ttn-lw/lib/errors/utils'

const SUPPORTED_LOCALES = process.predefined.SUPPORTED_LOCALES // Note: defined by webpack define plugin.
const defaultLanguage = 'en'
const defaultEnglishRegion = 'US'
const defaultLocale = `en-${defaultEnglishRegion}`

const getLanguageName = (languageCode, type = 'language') => {
  const languageNames = new Intl.DisplayNames(languageCode, { type })
  return languageNames.of(languageCode)
}

// Helper function to get the appropriate available locale from the desired locale.
const getAppropriateLocale = selectedLocale => {
  const desiredLocale =
    selectedLocale ||
    localStorage.getItem('locale') ||
    navigator.language ||
    navigator.browserLanguage ||
    defaultLocale

  const desiredLanguage = desiredLocale.split('-')[0]
  let locale

  // Check if the desired locale is available.
  if (SUPPORTED_LOCALES.includes(desiredLocale)) {
    locale = desiredLocale
  } else if (supportedLanguages.includes(desiredLanguage)) {
    if (SUPPORTED_LOCALES.includes(desiredLanguage)) {
      // Use the language without region if possible.
      locale = desiredLanguage
    } else {
      // Otherwise take the first available region.
      locale = SUPPORTED_LOCALES.find(l => l.split('-')[0] === desiredLanguage)
    }
  } else {
    // If the desired language is not available, fall back to the default.
    locale = defaultLocale
  }

  // Decorate English base locale with US region.
  if (locale === defaultLanguage) {
    locale = defaultLocale
  }

  return locale
}

const supportedLanguages = uniq(SUPPORTED_LOCALES.filter(l => l.split('-')[0]))
const getSupportedLocalesMap = () =>
  SUPPORTED_LOCALES.reduce((acc, locale) => {
    const loc = locale === 'en' ? `${locale}-${defaultEnglishRegion}` : locale
    const language = loc.split('-')[0]
    const region = loc.split('-').length > 1 ? loc.split('-')[1] : undefined

    const title = region
      ? `${getLanguageName(language)} (${getLanguageName(region, 'region')})`
      : getLanguageName(language)

    acc[loc] = title
    return acc
  }, {})

const m = defineMessages({
  switchingLanguage: 'Switching language…',
})

export const LanguageContext = createContext()

// `WithLocale` is a component that fetches necessary polyfills for `window.Intl`
// This component has to mount before any other component that makes use of the `Intl` API.
const WithLocale = ({ children }) => {
  const [polyfillsLoaded, setPolyfillsLoaded] = useState(false)
  const [error, setError] = useState(undefined)

  useEffect(() => {
    const initialize = async () => {
      try {
        // Load critical polyfills.
        if (!window.Intl.Locale) {
          log('Polyfilling Intl.Locale')
          await import(
            /* webpackChunkName: "locale-display-names" */ '@formatjs/intl-locale/polyfill'
          )
        }

        if (!window.Intl.DisplayNames) {
          log('Polyfilling Intl.DisplayNames')
          await import(
            /* webpackChunkName: "locale-display-names" */ '@formatjs/intl-displaynames/polyfill'
          )
          await Promise.all(
            supportedLanguages.map(async supportedLanguage => {
              log(`Polyfilling Intl.DisplayNames for locale "${supportedLanguage}"`)
              await import(
                /* webpackChunkName: "locale-display-names.[request]" */ `@formatjs/intl-displaynames/locale-data/${supportedLanguage}`
              )
            }),
          )
        }

        if (!window.Intl.ListFormat) {
          log('Polyfilling Intl.ListFormat')
          await import(
            /* webpackChunkName: "locale-list-format" */ '@formatjs/intl-listformat/polyfill'
          )
        }

        if (!window.Intl.PluralRules) {
          log('Polyfilling Intl.PluralRules')
          await import(
            /* webpackChunkName: "locale-plural-rules" */ '@formatjs/intl-pluralrules/polyfill'
          )
        }

        if (!window.Intl.NumberFormat) {
          log('Polyfilling Intl.NumberFormat')
          await import(
            /* webpackChunkName: "locale-number-format" */ '@formatjs/intl-numberformat/polyfill'
          )
        }

        if (!window.Intl.RelativeTimeFormat) {
          log('Polyfilling Intl.RelativeTimeFormat')
          await import(
            /* webpackChunkName: "locale-date-time-format" */ '@formatjs/intl-relativetimeformat/polyfill'
          )
        }

        if (!window.Intl.DateTimeFormat) {
          log('Polyfilling Intl.DateTimeFormat')
          await import(
            /* webpackChunkName: "locale-date-time-format" */ '@formatjs/intl-datetimeformat/polyfill'
          )
        }

        setPolyfillsLoaded(true)
      } catch (error) {
        setError(error)
      }
    }

    initialize()
  }, [setError])

  if (error) {
    throw error
  }

  if (!polyfillsLoaded) {
    // Not using `<Message />`, since we're initializing locales.
    return <Spinner center>Initializing multi-language support…</Spinner>
  }

  return <LocaleLoader>{children}</LocaleLoader>
}

// `LocaleLoader` is a component that fetches the user's preferred language and
// sets the language in the react-intl provider context. It will asynchronously
// fetch translated messages.
const LocaleLoader = ({ children }) => {
  const [intlState, setIntlState] = useState({
    locale: undefined,
    messages: undefined,
  })
  const [loaded, setLoaded] = useState(false)

  const handleIntlError = useCallback(err => {
    error(err)
    if (err.code === ReactIntlErrorCode.FORMAT_ERROR && !isDevelopment) {
      ingestError(err, { ingestedBy: 'IntlFormat' })
      return
    }
  }, [])

  const setLocale = useCallback(
    async desiredLocale => {
      const storeAsPreference = Boolean(desiredLocale)
      const locale = getAppropriateLocale(desiredLocale)
      const language = locale.split('-')[0]

      // Exit if we don't need to change anything.
      if (locale === intlState.locale) {
        setLoaded(true)
        return
      }

      // Load the language files.
      setLoaded(false)

      let promises = []
      if (locale === defaultLocale) {
        promises = [
          // For the default locale (en-US), we only need to load the backend messages.
          // For the other messages, we can use the default messages (via the `defaultLocale` prop).
          import(
            /* webpackChunkName: "lang.[request]" */ `../../locales/.backend/${language}.json`
          ),
        ]
      } else {
        promises = [
          import(/* webpackChunkName: "lang.[request]" */ `../../locales/${language}.json`),
        ]
      }

      // Load locale specific polyfills if needed.
      if (window.Intl.NumberFormat.polyfilled) {
        log(`Polyfilling NumberFormat for language ${language}`)
        promises.push(
          import(
            /* webpackChunkName: "locale.[request]" */ `@formatjs/intl-numberformat/locale-data/${language}`
          ),
        )
      }

      if (window.Intl.DateTimeFormat.polyfilled) {
        log(`Polyfilling DateTimeFormat for language ${language}`)
        promises.push(
          import(/* webpackChunkName: "locale" */ '@formatjs/intl-datetimeformat/add-all-tz'),
          import(
            /* webpackChunkName: "locale.[request]" */ `@formatjs/intl-datetimeformat/locale-data/${language}`
          ),
        )
      }

      if (window.Intl.RelativeTimeFormat.polyfilled) {
        log(`Polyfilling RelativeTimeFormat for language ${language}`)
        promises.push(
          import(
            /* webpackChunkName: "locale.[request]" */ `@formatjs/intl-relativetimeformat/locale-data/${language}`
          ),
        )
      }

      try {
        const res = await Promise.all(promises)
        setIntlState({ locale, messages: res[0] })

        // Set `lang` attribute of the `html` tag.
        document.documentElement.lang = language

        if (storeAsPreference) {
          localStorage.setItem('locale', locale)
        }
      } catch (err) {
        // Log the error and fall back to default locale.
        handleIntlError(err)
        setIntlState({ locale: defaultLocale, messages: undefined })
      }
      setLoaded(true)
    },
    [setLoaded, setIntlState, handleIntlError, intlState.locale],
  )

  useEffect(() => {
    // Perform the initial locale load.
    setLocale()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const languageContext = useMemo(
    () => ({
      locale: intlState.locale,
      supportedLocales: getSupportedLocalesMap(),
      setLocale,
    }),
    [intlState.locale, setLocale],
  )

  if (!Boolean(intlState.locale)) {
    // Not using `<Message />`, since we're initializing locales.
    return <Spinner center>Loading language…</Spinner>
  }

  return (
    <LanguageContext.Provider value={languageContext}>
      <IntlProvider
        messages={intlState.messages}
        locale={intlState.locale}
        defaultLocale={defaultLocale}
        onError={handleIntlError}
      >
        <Overlay loading={!loaded} visible={!loaded} spinnerMessage={m.switchingLanguage}>
          {children}
        </Overlay>
      </IntlProvider>
    </LanguageContext.Provider>
  )
}

LocaleLoader.propTypes = WithLocale.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
}

export default WithLocale
