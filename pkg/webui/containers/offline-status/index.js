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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'
import toast from '../../components/toast'
import PropTypes from '../../lib/prop-types'

import Status from '../../components/status'
import Message from '../../lib/components/message'

import sharedMessages from '../../lib/shared-messages'

import style from './offline.styl'

const m = defineMessages({
  offline: 'The Application is now offline',
  online: 'The Application is back online',
})

@bind
export default class OfflineStatus extends Component {
  static propTypes = {
    showOfflineOnly: PropTypes.bool,
    showWarnings: PropTypes.bool,
  }

  static defaultProps = {
    showOfflineOnly: false,
    showWarnings: false,
  }

  state = {
    online: true,
  }

  handleOnline() {
    this.setState({ online: true })
  }

  handleOffline() {
    this.setState({ online: false })
  }

  handleMessage(message, type) {
    toast({
      message,
      type,
    })
  }

  componentDidMount() {
    window.addEventListener('online', this.handleOnline)
    window.addEventListener('offline', this.handleOffline)
  }

  componentWillUnmount() {
    window.removeEventListener('online', this.handleOnline)
    window.removeEventListener('offline', this.handleOffline)
  }

  componentDidUpdate(prevProps) {
    const { showWarnings } = this.props
    const { online } = this.state
    if (showWarnings && online && !prevProps.online) {
      this.handleMessage(m.online, toast.types.INFO)
    } else if (showWarnings && !online) {
      this.handleMessage(m.offline, toast.types.ERROR)
    }
  }

  render() {
    const { showOfflineOnly } = this.props
    const { online } = this.state

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

    if (showOfflineOnly && online) {
      return null
    }

    return (
      <span>
        <Status className={style.status} status={statusIndicator}>
          <Message className={style.message} content={message} />
        </Status>
      </span>
    )
  }
}
