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


/* global require */

import React from 'react'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { IntlProvider } from 'react-intl'
import CancelablePromise from 'cancelable-promise'

import { withEnv } from '../../lib/env'
import log, { error } from '../../lib/log'

const defaultLanguage = 'en'
const xx = 'xx'
const dev = '../../lib/dev'

/**
 * WithLocale is a component that fetches the user's preferred language and
 * sets the langauge in th react-intl provider context. It will asynchronously fetch
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
  }

  toggle () {
    this.setState(state => ({ xx: !state.xx }))
  }

  componentWillReceiveProps (props) {
    this.check(this.props, props)
  }

  componentDidMount () {
    this.check({ env: {}}, this.props)

    if (dev) {
      window.addEventListener('keydown', this.onKeydown)
      log('Press alt + X to toggle the xx locale')
    }
  }

  check (prev, props) {
    const current = prev.user && prev.user.language || prev.env.default_language || defaultLanguage
    const next = props.user && props.user.language || props.env.default_language || defaultLanguage

    if (current !== next && next !== 'en') {
      this.promise = this.load(next)
    }
  }

  success (p) {
    this.setState({ messages: p[0]})
  }

  fail (err) {
    error(err)
    this.setState({ messages: null })
  }

  onKeydown (evt) {
    if (evt.altKey && evt.code === 'KeyX') {
      this.toggle()
    }
  }

  load (language) {
    let locale = navigator.language || navigator.browserLanguage || 'en-US'

    // if the browser locale does not match the lang on the html tag, prefer the lang
    // otherwise we get mixed languages.
    if (locale.split('-')[0] !== language && language !== xx) {
      locale = language
    }

    // load the language file
    const promises = [
      import(/* webpackChunkName: "lang.[request]" */ `../../locales/${language}.json`),
    ]

    // load locale polyfill if need be
    if (!window.Intl) {
      log(`Polyfilling locale ${locale} for language ${language}`)
      promises.push(import('intl'))
      promises.push(import(/* webpackChunkName: "locale.[request]" */ `intl/locale-data/jsonp/${locale}`))
    }

    return CancelablePromise
      .resolve(Promise.all(promises))
      .then(this.success)
      .catch(this.fail)
  }

  componentWillUnmount () {
    if (this.promise) {
      this.promise.cancel()
    }

    window.removeEventListener('onKeydown', this.onKeydown)
  }

  render () {
    const {
      user,
      children,
      env,
    } = this.props

    const { xx } = this.state

    let { messages } = this.state

    const lang = user && user.language || env.default_language || defaultLanguage

    if (dev && xx) {
      messages = require('../../locales/xx.json')
    }

    const key = `${lang}${messages ? 1 : 0}${xx ? 'xx' : ''}`
    const locale = lang === xx ? 'en' : lang

    return messages && (
      <IntlProvider key={key} messages={messages} locale={locale}>
        {children}
      </IntlProvider>
    )
  }
}
