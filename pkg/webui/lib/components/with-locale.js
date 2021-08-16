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

import React, { useState, useCallback, useEffect, createContext, useRef, useMemo } from 'react'
import { IntlProvider, defineMessages } from 'react-intl'
import CancelablePromise from 'cancelable-promise'
import { uniq } from 'lodash'

import Overlay from '@ttn-lw/components/overlay'
import Spinner from '@ttn-lw/components/spinner'

import PropTypes from '@ttn-lw/lib/prop-types'
import log, { error } from '@ttn-lw/lib/log'

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
const supportedLocalesMap = SUPPORTED_LOCALES.reduce((acc, locale) => {
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

// `UserLocale` is a component that fetches the user's preferred language and
// sets the language in the react-intl provider context. It will asynchronously
// fetch translated messages and polyfills `window.Intl`.
const UserLocale = ({ children }) => {
  const promise = useRef()
  const [intlState, setIntlState] = useState({
    locale: undefined,
    messages: undefined,
  })
  const [loaded, setLoaded] = useState(false)

  const handleIntlError = useCallback(err => {
    error(err)
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

      // Load polyfills if needed.
      if (!window.Intl.NumberFormat || !window.Intl.DateTimeFormat) {
        log(`Polyfilling locale ${locale} for language ${language}`)
        promises.push(import('intl'))
        promises.push(
          import(/* webpackChunkName: "locale.[request]" */ `intl/locale-data/jsonp/${locale}`),
        )
      }

      if (!window.Intl.DisplayNames) {
        log(`Polyfilling Intl.DisplayNames`)
        promises.push(
          import(
            /* webpackChunkName: "locale-display-names" */ '@formatjs/intl-displaynames/polyfill'
          ),
        )
      }

      if (!window.Intl.ListFormat) {
        log(`Polyfilling Intl.ListFormat`)
        promises.push(
          import(/* webpackChunkName: "locale-list-format" */ '@formatjs/intl-listformat/polyfill'),
        )
      }

      if (!window.Intl.PluralRules) {
        log('Polyfilling Intl.PluralRules')
        promises.push(
          import(
            /* webpackChunkName: "locale-plural-rules" */ '@formatjs/intl-pluralrules/polyfill'
          ),
        )
      }

      if (!window.Intl.RelativeTimeFormat) {
        log(`Polyfilling Intl.RelativeTimeFormat data for language ${locale}`)
        promises.push(
          import(
            /* webpackChunkName: "locale-time-polyfill" */ '@formatjs/intl-relativetimeformat/polyfill'
          ),
        )
        promises.push(
          import(
            /* webpackChunkName: "locale-time-locales.[request]" */ `@formatjs/intl-relativetimeformat/locale-data/${locale}`
          ),
        )
      }

      promise.current = CancelablePromise.resolve(Promise.all(promises))
      try {
        const res = await promise.current
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
    return () => {
      if (promise.current && promise.current.cancel) {
        promise.current.cancel()
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const languageContext = useMemo(
    () => ({
      locale: intlState.locale,
      supportedLocales: supportedLocalesMap,
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

UserLocale.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
}

export default UserLocale
