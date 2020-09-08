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
import classnames from 'classnames'

import toast from '@ttn-lw/components/toast'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './offline.styl'

const m = defineMessages({
  offline: 'The application went offline',
  online: 'The application is back online',
})

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
    online: window.navigator.onLine,
  }

  @bind
  handleOnline() {
    this.setState({ online: true })
  }

  @bind
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

  componentDidUpdate(_prevProps, prevState) {
    const { showWarnings } = this.props
    const { online } = this.state
    if (showWarnings && online && !prevState.online) {
      this.handleMessage(m.online, toast.types.INFO)
    } else if (showWarnings && !online) {
      this.handleMessage(m.offline, toast.types.ERROR)
    }
  }

  render() {
    const { showOfflineOnly } = this.props
    const { online } = this.state

    if ((showOfflineOnly && online) || online === undefined) {
      return null
    }

    return (
      <span className={classnames(style.status, { [style.online]: online })}>
        <Icon className={style.icon} icon={online ? 'info' : 'error'} />
        <Message content={online ? sharedMessages.online : sharedMessages.offline} />
      </span>
    )
  }
}
