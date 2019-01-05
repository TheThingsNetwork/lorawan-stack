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

import Message from './message'

/**
 * IntlHelmet is a HOC that enables usage of <Message /> components inside
 * header tags used in react-helmet.
 */
@injectIntl
export default class IntlHelmet extends React.Component {
  render () {
    const { intl, children, ...rest } = this.props

    const translatedChildren = React.Children.map(children, function (child) {
      const childrenChildren = React.Children.map(child.props.children, function (childChild) {
        let message = childChild
        if (childChild.type === Message) {
          // Evaluate the message
          message = intl.formatMessage(childChild.props.content)
        }
        return message
      })
      return <child.type>{childrenChildren.join(' ')}</child.type>
    })

    return <Helmet {...rest}>{translatedChildren}</Helmet>
  }
}
