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
import { push } from 'connected-react-router'
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import CodeEditor from '../../../components/code-editor'
import ProgressBar from '../../../components/progress-bar'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import DeviceBulkCreateForm from '../../components/device-bulk-create-form'
import SubmitBar from '../../../components/submit-bar'
import Button from '../../../components/button'
import Notification from '../../../components/notification'
import api from '../../api'
import PropTypes from '../../../lib/prop-types'
import Message from '../../../lib/components/message'
import Status from '../../../components/status'

import style from './device-bulk-creator.styl'

const m = defineMessages({
  proceed: 'Proceed',
  retry: 'Retry',
  converting: 'Converting Templates…',
  creating: 'Creating devices…',
  operationInProgress: 'Operation in progress',
  operationHalted: 'Operation halted',
  operationFinished: 'Operation finished',
  errorTitle: 'Could not complete operation',
})

const initialState = {
  log: '',
  totalDevices: undefined,
  devicesComplete: 0,
  status: 'initial',
  step: 'inital',
  error: undefined,
}

const statusMap = {
  processing: 'good',
  error: 'bad',
  finished: 'good',
}

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
  }),
  dispatch => ({
    redirectToList: appId => dispatch(push(`/applications/${appId}/devices`)),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    redirectToList: () => dispatchProps.redirectToList(stateProps.appId),
  }),
)
export default class DeviceBulkCreator extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    redirectToList: PropTypes.func.isRequired,
  }

  state = { ...initialState }

  @bind
  appendToLog(message) {
    const { log } = this.state
    const text = typeof message !== 'string' ? JSON.stringify(message, null, 2) : message
    this.setState({
      log: `${log}\n${text}`,
    })
  }

  @bind
  handleCreationProgress(device) {
    const { devicesComplete } = this.state

    this.appendToLog(device)
    this.setState({ devicesComplete: devicesComplete + 1 })
  }

  @bind
  handleError(error) {
    const { log } = this.state
    const json = JSON.stringify(error, null, 2)
    this.setState({ error, status: 'error', log: `${log}\n${json}` })
  }

  @bind
  async handleSubmit(values) {
    const { appId } = this.props
    const { format_id, data } = values

    try {
      // Start template conversion
      this.setState({ step: 'conversion', status: 'processing' })
      this.appendToLog('Converting device templates…')
      const templateStream = await api.deviceTemplates.convert(format_id, data)
      const devices = await new Promise(
        function(resolve, reject) {
          const chunks = []

          templateStream.on(
            'chunk',
            function(message) {
              this.appendToLog(message)
              chunks.push(message)
            }.bind(this),
          )
          templateStream.on('error', reject)
          templateStream.on('close', () => resolve(chunks))
        }.bind(this),
      )

      // Start batch device creation
      this.setState({
        step: 'creation',
        totalDevices: devices.length,
      })
      this.appendToLog('Creating devices…')
      const createStream = api.device.bulkCreate(appId, devices, ['is', 'as', 'js'])

      await new Promise(
        function(resolve, reject) {
          createStream.on('chunk', this.handleCreationProgress)
          createStream.on('error', reject)
          createStream.on('close', resolve)
        }.bind(this),
      )

      this.setState({ status: 'finished' })
    } catch (error) {
      this.handleError(error)
    }
  }

  @bind
  handleReset() {
    this.setState(initialState)
  }

  get processor() {
    const { log, totalDevices, devicesComplete, status, step, error } = this.state
    const hasErrored = status === 'error'
    const { redirectToList } = this.props
    const operationMessage = step === 'conversion' ? m.converting : m.creating
    let statusMessage = m.operationInProgress
    if (status === 'error') {
      statusMessage = m.operationHalted
    } else if (status === 'finished') {
      statusMessage = m.operationFinished
    }
    return (
      <div>
        <Message className={style.title} component="h4" content={operationMessage} />
        {!hasErrored ? (
          <React.Fragment>
            <Status
              label={statusMessage}
              pulse={status === 'processing'}
              status={statusMap[status] || 'unknown'}
            />
            <ProgressBar
              current={devicesComplete}
              target={totalDevices}
              showStatus
              showEstimation={!hasErrored}
              className={style.progressBar}
            />
          </React.Fragment>
        ) : (
          <Notification small error={error} title={m.errorTitle} />
        )}
        <CodeEditor
          className={style.logOutput}
          minLines={20}
          maxLines={20}
          mode="json"
          name="process_log"
          readOnly
          value={log}
          editorOptions={{ useWorker: false }}
          showGutter={false}
          scrollToBottom
        />
        <SubmitBar>
          <Button
            busy={status !== 'finished' && !hasErrored}
            message={hasErrored ? m.retry : m.proceed}
            onClick={hasErrored ? this.handleReset : redirectToList}
          />
        </SubmitBar>
      </div>
    )
  }

  get form() {
    const initialValues = {
      format_id: '',
      data: '',
    }
    return <DeviceBulkCreateForm initialValues={initialValues} onSubmit={this.handleSubmit} />
  }

  render() {
    const { step } = this.state

    switch (step) {
      case 'conversion':
      case 'creation':
        return this.processor
      case 'initial':
      default:
        return this.form
    }
  }
}
