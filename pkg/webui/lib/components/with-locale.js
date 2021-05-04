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

/* global require */

import React from 'react'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { IntlProvider } from 'react-intl'
import CancelablePromise from 'cancelable-promise'

import Spinner from '@ttn-lw/components/spinner'

import PropTypes from '@ttn-lw/lib/prop-types'
import log, { error } from '@ttn-lw/lib/log'
import { selectLanguageConfig } from '@ttn-lw/lib/selectors/env'

const defaultLocale = process.predefined.DEFAULT_MESSAGES_LOCALE // Note: defined by webpack define plugin.
const envLocale = selectLanguageConfig()
const defaultLanguage = defaultLocale.split('-')[0] || 'en'
const xx = 'xx'
const dev = '../../dev'

/**
 * WithLocale is a component that fetches the user's preferred language and
 * sets the language in th react-intl provider context. It will asynchronously
 * fetch translated messages and polyfills `window.Intl`.
 * The default language will be fetched from the env.
 */
@connect(state => ({
  user: state.user,
  checking: state.user.checking,
}))
export default class UserLocale extends React.PureComponent {
  static propTypes = {
    children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
    user: PropTypes.shape({
      language: PropTypes.string,
    }),
  }

  static defaultProps = {
    user: undefined,
  }
  /** @private */
  promise = null

  state = {
    messages: process.predefined.DEFAULT_MESSAGES, // Note: defined by webpack define plugin.
    xx: false,
    loaded: false,
  }

  @bind
  toggle() {
    this.setState(state => ({ xx: !state.xx }))
  }

  componentDidUpdate(props) {
    this.check(this.props, props)
  }

  componentDidMount() {
    this.check({ user: {} }, this.props, true)

    if (dev) {
      window.addEventListener('keydown', this.onKeydown)
      log('Press alt + L to toggle the xx locale')
    }
  }

  @bind
  check(prev, props, enforce = false) {
    const current = (prev.user && prev.user.language) || envLocale
    const next = (props.user && props.user.language) || envLocale || defaultLanguage

    if (enforce || current !== next) {
      this.promise = this.load(next)
    }
  }

  @bind
  success(p, withLocale) {
    const newState = { loaded: true }
    if (withLocale) {
      const frontendMessages = p[0]
      const backendMessages = p[1]

      newState.messages = { ...frontendMessages, ...backendMessages }
    }

    this.setState(newState)
  }

  @bind
  fail(err) {
    error(err)
    this.setState({ messages: null, loaded: true })
  }

  @bind
  onKeydown(evt) {
    if (evt.altKey && evt.code === 'KeyL') {
      this.toggle()
    }
  }

  @bind
  async load(language) {
    let locale = navigator.language || navigator.browserLanguage || defaultLocale

    // If the browser locale does not match the lang on the html tag, prefer the
    // lang otherwise we get mixed languages.
    if (locale.split('-')[0] !== language && language !== xx) {
      locale = language
    }

    // Load the language files if needed.
    await this.setState({ loaded: false })

    let promises = []
    let withLocale = false
    if (language !== defaultLanguage) {
      withLocale = true
      promises = [
        import(/* WebpackChunkName: "lang.[request]" */ `../../locales/${language}.json`), // Frontend messages
        import(/* WebpackChunkName: "lang.[request]" */ `../../locales/.backend/${language}.json`), // Backend messages
      ]
    }

    if (!window.Intl.NumberFormat || !window.Intl.DateTimeFormat) {
      log(`Polyfilling locale ${locale} for language ${language}`)
      promises.push(import('intl'))
      promises.push(
        import(/* WebpackChunkName: "locale.[request]" */ `intl/locale-data/jsonp/${locale}`),
      )
    }

    if (!window.Intl.PluralRules) {
      log(`Polyfilling Intl.PluralRules`)
      promises.push(import('intl-pluralrules'))
    }

    if (!window.Intl.RelativeTimeFormat) {
      log(`Polyfilling Intl.RelativeTimeFormat data for language ${language}`)
      promises.push(import('@formatjs/intl-relativetimeformat/polyfill'))
      promises.push(import(`@formatjs/intl-relativetimeformat/locale-data/${language}`))
    }

    return CancelablePromise.resolve(Promise.all(promises))
      .then(result => this.success(result, withLocale))
      .catch(this.fail)
  }

  componentWillUnmount() {
    if (this.promise) {
      this.promise.cancel()
    }

    window.removeEventListener('onKeydown', this.onKeydown)
  }

  render() {
    const { user, children } = this.props

    const { messages, loaded, xx } = this.state

    if (!loaded) {
      // Not using <Message />, since we're initializing locales.
      return <Spinner center>Loading locale…</Spinner>
    }

    const lang = (user && user.language) || envLocale || defaultLanguage

    if (dev && xx) {
      messages = {
        ...require('../../locales/xx.json'),
        ...require('../../locales/.backend/xx.json'),
      }
    }

    const key = `${lang}${messages ? 1 : 0}${xx ? 'xx' : ''}`
    const locale = lang === xx ? 'en' : lang

    return (
      messages && (
        <IntlProvider key={key} messages={messages} locale={locale}>
          {children}
        </IntlProvider>
      )
    )
  }
}
