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

import Spinner from '../../components/spinner'

import log, { error } from '../log'
import { withEnv } from './env'

const defaultLocale = process.predefined.DEFAULT_MESSAGES_LOCALE // Note: defined by webpack define plugin
const defaultLanguage = defaultLocale.split('-')[0] || 'en'
const xx = 'xx'
const dev = '../../dev'

/**
 * WithLocale is a component that fetches the user's preferred language and
 * sets the language in th react-intl provider context. It will asynchronously fetch
 * translated messages and polyfills window.Intl.
 *
 * The default language will be fetched from the env.
 */
@connect(state => ({
  user: state.user,
  checking: state.user.checking,
}))
@withEnv
@bind
export default class UserLocale extends React.PureComponent {
  /** @private */
  promise = null

  state = {
    messages: process.predefined.DEFAULT_MESSAGES, // Note: defined by webpack define plugin
    xx: false,
    loaded: false,
  }

  toggle() {
    this.setState(state => ({ xx: !state.xx }))
  }

  componentDidUpdate(props) {
    this.check(this.props, props)
  }

  componentDidMount() {
    this.check({ env: { config: {} }, user: {} }, this.props)

    if (dev) {
      window.addEventListener('keydown', this.onKeydown)
      log('Press alt + L to toggle the xx locale')
    }
  }

  check(prev, props) {
    const current = (prev.user && prev.user.language) || prev.env.config.language
    const next = (props.user && props.user.language) || props.env.config.language || defaultLanguage

    if (current !== next) {
      this.promise = this.load(next)
    }
  }

  success(p, withLocale) {
    const newState = { loaded: true }
    if (withLocale) {
      const frontendMessages = p[0]
      const backendMessages = p[1]

      newState.messages = { ...frontendMessages, ...backendMessages }
    }

    this.setState(newState)
  }

  fail(err) {
    error(err)
    this.setState({ messages: null, loaded: true })
  }

  onKeydown(evt) {
    if (evt.altKey && evt.code === 'KeyL') {
      this.toggle()
    }
  }

  async load(language) {
    let locale = navigator.language || navigator.browserLanguage || defaultLocale

    // if the browser locale does not match the lang on the html tag, prefer the lang
    // otherwise we get mixed languages.
    if (locale.split('-')[0] !== language && language !== xx) {
      locale = language
    }

    // load the language files if needed
    await this.setState({ loaded: false })

    let promises = []
    let withLocale = false
    if (language !== defaultLanguage) {
      withLocale = true
      promises = [
        import(/* webpackChunkName: "lang.[request]" */ `../../locales/${language}.json`), // Frontend messages
        import(/* webpackChunkName: "lang.[request]" */ `../../locales/.backend/${language}.json`), // Backend messages
      ]
    }

    if (!window.Intl.NumberFormat || !window.Intl.DateTimeFormat) {
      log(`Polyfilling locale ${locale} for language ${language}`)
      promises.push(import('intl'))
      promises.push(
        import(/* webpackChunkName: "locale.[request]" */ `intl/locale-data/jsonp/${locale}`),
      )
    }

    if (!window.Intl.PluralRules) {
      log(`Polyfilling Intl.PluralRules`)
      promises.push(import('intl-pluralrules'))
    }

    if (!window.Intl.RelativeTimeFormat) {
      log(`Polyfilling Intl.RelativeTimeFormat data for language ${language}`)
      promises.push(import('@formatjs/intl-relativetimeformat/polyfill'))
      promises.push(import(`@formatjs/intl-relativetimeformat/dist/locale-data/${language}`))
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
    const {
      user,
      children,
      env: { config },
    } = this.props

    const { messages, loaded, xx } = this.state

    if (!loaded) {
      // Not using <Message />, since we're initializing locales
      return <Spinner center>Loading locale…</Spinner>
    }

    const lang = (user && user.language) || config.language || defaultLanguage

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
