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

import api from '@console/api'

import CodeEditor from '@ttn-lw/components/code-editor'
import ProgressBar from '@ttn-lw/components/progress-bar'
import SubmitBar from '@ttn-lw/components/submit-bar'
import Button from '@ttn-lw/components/button'
import ErrorNotification from '@ttn-lw/components/error-notification'
import Status from '@ttn-lw/components/status'

import Message from '@ttn-lw/lib/components/message'

import DeviceImportForm from '@console/components/device-import-form'

import PropTypes from '@ttn-lw/lib/prop-types'
import { selectNsConfig, selectJsConfig, selectAsConfig } from '@ttn-lw/lib/selectors/env'

import randomByteString from '@console/lib/random-bytes'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './device-importer.styl'

const m = defineMessages({
  proceed: 'Proceed',
  retry: 'Retry',
  converting: 'Converting templates…',
  creating: 'Creating end devices…',
  operationInProgress: 'Operation in progress',
  operationHalted: 'Operation halted',
  operationFinished: 'Operation finished',
  errorTitle: 'There was an error and the operation could not be completed',
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
  state => {
    const asConfig = selectAsConfig()
    const nsConfig = selectNsConfig()
    const jsConfig = selectJsConfig()
    const availableComponents = ['is']
    if (nsConfig.enabled) availableComponents.push('ns')
    if (jsConfig.enabled) availableComponents.push('js')
    if (asConfig.enabled) availableComponents.push('as')

    return {
      appId: selectSelectedApplicationId(state),
      nsConfig,
      jsConfig,
      asConfig,
      availableComponents,
    }
  },
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
export default class DeviceImporter extends Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    asConfig: PropTypes.stackComponent.isRequired,
    availableComponents: PropTypes.components.isRequired,
    jsConfig: PropTypes.stackComponent.isRequired,
    nsConfig: PropTypes.stackComponent.isRequired,
    redirectToList: PropTypes.func.isRequired,
  }

  constructor(props) {
    super(props)

    this.state = { ...initialState }
    this.editorRef = React.createRef()
  }

  componentDidUpdate(prevProps, prevState) {
    const { status } = this.state

    if (prevState.status === 'initial' && status !== 'initial') {
      // Disable undo manager of the code editor to release old logs from the heap.
      // Without this fix the browser can run out of memory when importing many end devices.
      this.editorRef.current.editor.session.setUndoManager(null)
    }
  }

  @bind
  appendToLog(message) {
    const text = typeof message !== 'string' ? JSON.stringify(message, null, 2) : message
    this.setState(({ log }) => ({ log: `${log}\n${text}` }))
  }

  @bind
  handleCreationProgress(device) {
    this.appendToLog(device)
    this.setState(({ devicesComplete }) => ({ devicesComplete: devicesComplete + 1 }))
  }

  @bind
  handleError(error) {
    const json = JSON.stringify(error, null, 2)
    this.setState(({ log }) => ({ error, status: 'error', log: `${log}\n${json}` }))
  }

  @bind
  async handleSubmit(values) {
    const { appId, jsConfig, nsConfig, asConfig } = this.props
    const {
      format_id,
      data,
      set_claim_auth_code,
      components: { js: jsSelected, as: asSelected, ns: nsSelected },
    } = values

    try {
      // Start template conversion.
      this.setState({ step: 'conversion', status: 'processing' })
      this.appendToLog('Converting end device templates…')
      const templateStream = await api.deviceTemplates.convert(format_id, data)
      const devices = await new Promise((resolve, reject) => {
        const chunks = []

        templateStream.on('chunk', message => {
          this.appendToLog(message)
          chunks.push(message)
        })
        templateStream.on('error', reject)
        templateStream.on('close', () => resolve(chunks))
      })

      // Apply default values.
      for (const deviceAndFieldMask of devices) {
        const { end_device: device, field_mask } = deviceAndFieldMask
        if (set_claim_auth_code && jsSelected) {
          device.claim_authentication_code = { value: randomByteString(4 * 2) }
          field_mask.paths.push('claim_authentication_code')
        }
        if (device.supports_join && !device.join_server_address && jsConfig.enabled && jsSelected) {
          device.join_server_address = new URL(jsConfig.base_url).hostname
          field_mask.paths.push('join_server_address')
        }
        if (!device.application_server_address && asConfig.enabled && asSelected) {
          device.application_server_address = new URL(asConfig.base_url).hostname
          field_mask.paths.push('application_server_address')
        }
        if (!device.network_server_address && nsConfig.enabled && nsSelected) {
          device.network_server_address = new URL(nsConfig.base_url).hostname
          field_mask.paths.push('network_server_address')
        }
      }

      // Start batch device creation.
      this.setState({
        step: 'creation',
        totalDevices: devices.length,
      })
      this.appendToLog('Creating end devices…')
      const createStream = api.device.bulkCreate(appId, devices)

      await new Promise((resolve, reject) => {
        createStream.on('chunk', this.handleCreationProgress)
        createStream.on('error', reject)
        createStream.on('close', resolve)
      })

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
          <ErrorNotification small content={error} title={m.errorTitle} />
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
          editorRef={this.editorRef}
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
    const { availableComponents } = this.props
    const initialValues = {
      format_id: '',
      data: '',
      set_claim_auth_code: false,
      components: availableComponents.reduce((o, c) => ({ ...o, [c]: true }), {}),
    }
    return (
      <DeviceImportForm
        components={availableComponents}
        initialValues={initialValues}
        onSubmit={this.handleSubmit}
      />
    )
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
