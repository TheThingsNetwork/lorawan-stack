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

import React from 'react'

import { Helmet } from 'react-helmet'
import { injectIntl } from 'react-intl'

import { warn } from '../log'

/**
 * IntlHelmet is a HOC that enables usage of i18n message objects inside the
 * props in react-helmet, which will be translated automatically.
 */
@injectIntl
export default class IntlHelmet extends React.Component {
  componentDidMount() {
    if (this.props.children) {
      warn(`Children of <IntlHelmet /> will not be translated. If you tried to
translate head elements with <Message />, use props with message objects
instead.`)
    }
  }

  render() {
    const { intl, children, values, ...rest } = this.props
    let translatedRest = {}
    for (const key in rest) {
      let prop = rest[key]
      if (typeof prop === 'object' && prop.id && prop.defaultMessage) {
        const messageValues = values || {}
        const translatedMessageValues = {}

        for (const entry in messageValues) {
          const content = messageValues[entry]
          if (typeof content === 'object' && prop.id && prop.defaultMessage) {
            translatedMessageValues[entry] = intl.formatMessage(content)
          } else {
            translatedMessageValues[entry] = messageValues[entry]
          }
        }

        prop = intl.formatMessage(prop, translatedMessageValues)
      }

      translatedRest = {
        ...translatedRest,
        [key]: prop,
      }
    }

    return <Helmet {...translatedRest}>{children}</Helmet>
  }
}
