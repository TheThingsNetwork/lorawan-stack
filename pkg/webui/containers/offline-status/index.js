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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import PropTypes from '../../lib/prop-types'

import sharedMessages from '../../lib/shared-messages'
import { selectOfflineStatus } from '../../lib/selectors/offline'

import Offline from '../../components/offline'

@connect(state => ({ online: selectOfflineStatus(state) }))
export default class OfflineStatus extends Component {
  static propTypes = {
    online: PropTypes.bool,
  }

  static defaultProps = {
    online: undefined,
  }

  render() {
    const { online } = this.props

    let statusIndicator = null
    let message = null

    if (online === undefined) {
      return null
    } else if (online) {
      message = sharedMessages.online
      statusIndicator = 'good'
    } else {
      statusIndicator = 'bad'
      message = sharedMessages.offline
    }

    return <Offline status={statusIndicator} content={message} />
  }
}
