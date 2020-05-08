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

import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import Input from '@ttn-lw/components/input'
import Checkbox from '@ttn-lw/components/checkbox'
import Radio from '@ttn-lw/components/radio-button'
import Select from '@ttn-lw/components/select'
import SubmitBar from '@ttn-lw/components/submit-bar'

import Message from '@ttn-lw/lib/components/message'

import DevAddrInput from '@console/containers/dev-addr-input'
import JoinEUIPrefixesInput from '@console/containers/join-eui-prefixes-input'
import { NsFrequencyPlansSelect } from '@console/containers/freq-plans-select'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectNsConfig, selectJsConfig, selectAsConfig } from '@ttn-lw/lib/selectors/env'
import errorMessages from '@ttn-lw/lib/errors/error-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import randomByteString from '@console/lib/random-bytes'

import m from './messages'
import validationSchema from './validation-schema'

const random16BytesString = () => randomByteString(32)

class DeviceDataForm extends Component {
  static propTypes = {
    mayEditKeys: PropTypes.bool.isRequired,
    onSubmit: PropTypes.func.isRequired,
    onSubmitSuccess: PropTypes.func,
  }

  static defaultProps = {
    onSubmitSuccess: () => null,
  }

  formRef = React.createRef()
  state = {
    otaa: true,
    resets_join_nonces: false,
    resets_f_cnt: false,
    external_js: true,
    lorawan_version: '',
  }

  @bind
  handleOTAASelect() {
    this.setState({ otaa: true })
  }

  @bind
  handleABPSelect() {
    this.setState({ otaa: false })
  }

  @bind
  handleResetsJoinNoncesChange(evt) {
    this.setState({ resets_join_nonces: evt.target.checked })
  }

  @bind
  handleResetsFrameCountersChange(evt) {
    this.setState({ resets_f_cnt: evt.target.checked })
  }

  @bind
  handleLorawanVersionChange(lorawan_version) {
    this.setState({ lorawan_version })
  }

  @bind
  async handleExternalJoinServerChange(evt) {
    const external_js = evt.target.checked
    await this.setState(({ resets_join_nonces }) => ({
      external_js,
      resets_join_nonces: external_js ? false : resets_join_nonces,
    }))

    const jsConfig = selectJsConfig()
    const { setValues, state } = this.formRef.current

    // Reset Join Server related entries if the device is provisined by an
    // external JS.
    if (external_js) {
      setValues({
        ...state.values,
        root_keys: {
          nwk_key: {},
          app_key: {},
        },
        resets_join_nonces: false,
        join_server_address: undefined,
        _external_js: external_js,
      })
    } else {
      let join_server_address = state.join_server_address

      // Reset `join_server_address` if is present after disabling external JS
      // provisioning.
      if (jsConfig.enabled && !Boolean(join_server_address)) {
        join_server_address = new URL(jsConfig.base_url).hostname
      }

      setValues({
        ...state.values,
        join_server_address,
        _external_js: external_js,
      })
    }
  }

  @bind
  async handleSubmit(values, { setSubmitting }) {
    const { onSubmit, onSubmitSuccess } = this.props
    const {
      _external_js,
      _may_edit_keys,
      _activation_mode,
      ...castedValues
    } = validationSchema.cast(values)
    await this.setState({ error: '' })

    try {
      const device = await onSubmit(castedValues)
      await onSubmitSuccess(device)
    } catch (error) {
      setSubmitting(false)
      const err = error instanceof Error ? errorMessages.genericError : error
      await this.setState({ error: err })
    }
  }

  get ABPSection() {
    const { resets_f_cnt, lorawan_version } = this.state

    const lwVersion = Boolean(lorawan_version)
      ? parseInt(lorawan_version.replace(/\D/g, '').padEnd(3, 0))
      : 0

    return (
      <>
        <Form.Field
          title={sharedMessages.devEUI}
          name="ids.dev_eui"
          type="byte"
          min={8}
          max={8}
          description={m.deviceEUIDescription}
          required={lwVersion === 104}
          component={Input}
        />
        <DevAddrInput
          title={sharedMessages.devAddr}
          name="session.dev_addr"
          placeholder={m.leaveBlankPlaceholder}
          description={m.deviceAddrDescription}
          required
        />
        <Form.Field
          title={sharedMessages.nwkSKey}
          name="session.keys.f_nwk_s_int_key.key"
          type="byte"
          min={16}
          max={16}
          description={m.nwkSKeyDescription}
          component={Input.Generate}
          onGenerateValue={random16BytesString}
          required
        />
        <Form.Field
          title={sharedMessages.appSKey}
          name="session.keys.app_s_key.key"
          type="byte"
          min={16}
          max={16}
          description={m.appSKeyDescription}
          component={Input.Generate}
          onGenerateValue={random16BytesString}
          required
        />
        {lwVersion >= 110 && (
          <>
            <Form.Field
              title={sharedMessages.sNwkSIKey}
              name="session.keys.s_nwk_s_int_key.key"
              type="byte"
              min={16}
              max={16}
              description={m.sNwkSIKeyDescription}
              component={Input.Generate}
              onGenerateValue={random16BytesString}
              required
            />
            <Form.Field
              title={sharedMessages.nwkSEncKey}
              name="session.keys.nwk_s_enc_key.key"
              type="byte"
              min={16}
              max={16}
              description={m.nwkSEncKeyDescription}
              component={Input.Generate}
              onGenerateValue={random16BytesString}
              required
            />
          </>
        )}
        <Form.Field
          title={m.resetsFCnt}
          onChange={this.handleResetsFrameCountersChange}
          warning={resets_f_cnt ? m.resetWarning : undefined}
          name="mac_settings.resets_f_cnt"
          component={Checkbox}
        />
      </>
    )
  }

  get OTAASection() {
    const { mayEditKeys } = this.props
    const { resets_join_nonces, external_js } = this.state

    return (
      <>
        <JoinEUIPrefixesInput
          title={sharedMessages.joinEUI}
          name="ids.join_eui"
          description={m.joinEUIDescription}
          required
          showPrefixes
        />
        <Form.Field
          title={sharedMessages.devEUI}
          name="ids.dev_eui"
          type="byte"
          min={8}
          max={8}
          description={m.deviceEUIDescription}
          required
          component={Input}
        />
        <Form.Field
          title={m.externalJoinServer}
          description={m.externalJoinServerDescription}
          name="_external_js"
          onChange={this.handleExternalJoinServerChange}
          component={Checkbox}
        />
        <Form.Field
          title={sharedMessages.joinServerAddress}
          placeholder={external_js ? m.external : sharedMessages.addressPlaceholder}
          name="join_server_address"
          component={Input}
          disabled={external_js}
        />
        {mayEditKeys && (
          <>
            <Form.Field
              title={sharedMessages.appKey}
              name="root_keys.app_key.key"
              type="byte"
              min={16}
              max={16}
              description={m.appKeyDescription}
              placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
              component={Input.Generate}
              disabled={external_js}
              onGenerateValue={random16BytesString}
              mayGenerateValue={!external_js && mayEditKeys}
            />
            <Form.Field
              title={sharedMessages.nwkKey}
              name="root_keys.nwk_key.key"
              type="byte"
              min={16}
              max={16}
              description={m.nwkKeyDescription}
              placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
              component={Input.Generate}
              disabled={external_js}
              onGenerateValue={random16BytesString}
              mayGenerateValue={!external_js && mayEditKeys}
            />
          </>
        )}
        <Form.Field
          title={m.resetsJoinNonces}
          onChange={this.handleResetsJoinNoncesChange}
          warning={resets_join_nonces ? m.resetWarning : undefined}
          name="resets_join_nonces"
          component={Checkbox}
          disabled={external_js}
        />
        <Form.Field
          title={m.homeNetID}
          description={m.homeNetIDDescription}
          placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
          name="net_id"
          type="byte"
          min={3}
          max={3}
          component={Input}
          disabled={external_js}
        />
        <Form.Field
          title={m.asServerID}
          name="application_server_id"
          description={m.asServerIDDescription}
          placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
          component={Input}
          disabled={external_js}
        />
        <Form.Field
          title={m.asServerKekLabel}
          name="application_server_kek_label"
          description={m.asServerKekLabelDescription}
          placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
          component={Input}
          disabled={external_js}
        />
        <Form.Field
          title={m.nsServerKekLabel}
          name="network_server_kek_label"
          description={m.nsServerKekLabelDescription}
          placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
          component={Input}
          disabled={external_js}
        />
      </>
    )
  }

  render() {
    const { mayEditKeys } = this.props
    const { otaa, error, external_js } = this.state

    const emptyValues = {
      ids: {
        device_id: undefined,
        join_eui: undefined,
        dev_eui: undefined,
      },
      _activation_mode: 'otaa',
      lorawan_version: undefined,
      lorawan_phy_version: undefined,
      frequency_plan_id: undefined,
      supports_class_c: false,
      resets_join_nonces: false,
      root_keys: {
        nwk_key: { key: undefined },
        app_key: { key: undefined },
      },
      session: {
        dev_addr: undefined,
        keys: {
          f_nwk_s_int_key: { key: undefined },
          s_nwk_s_int_key: { key: undefined },
          nwk_s_enc_key: { key: undefined },
          app_s_key: { key: undefined },
        },
      },
      mac_settings: {
        resets_f_cnt: false,
      },
    }

    const nsConfig = selectNsConfig()
    const asConfig = selectAsConfig()
    const jsConfig = selectJsConfig()
    const joinServerAddress = jsConfig.enabled ? new URL(jsConfig.base_url).hostname : ''

    const formValues = {
      ...emptyValues,
      network_server_address: nsConfig.enabled ? new URL(nsConfig.base_url).hostname : '',
      application_server_address: asConfig.enabled ? new URL(asConfig.base_url).hostname : '',
      join_server_address: external_js ? undefined : joinServerAddress,
      _activation_mode: otaa ? 'otaa' : 'abp',
      _external_js: external_js,
      _may_edit_keys: mayEditKeys,
    }

    return (
      <Form
        error={error}
        onSubmit={this.handleSubmit}
        validationSchema={validationSchema}
        submitEnabledWhenInvalid
        initialValues={formValues}
        formikRef={this.formRef}
      >
        <Message component="h4" content={sharedMessages.generalSettings} />
        <Form.Field
          title={sharedMessages.devID}
          name="ids.device_id"
          placeholder={m.deviceIdPlaceholder}
          autoFocus
          required
          component={Input}
        />
        <Form.Field
          title={sharedMessages.devName}
          name="name"
          placeholder={m.deviceNamePlaceholder}
          component={Input}
        />
        <Form.Field
          title={sharedMessages.devDesc}
          name="description"
          type="textarea"
          placeholder={m.deviceDescPlaceholder}
          description={m.deviceDescDescription}
          component={Input}
        />
        <Message component="h4" content={m.lorawanOptions} />
        <Form.Field
          title={sharedMessages.macVersion}
          description={sharedMessages.macVersionDescription}
          name="lorawan_version"
          component={Select}
          required
          onChange={this.handleLorawanVersionChange}
          options={[
            { value: '1.0.0', label: 'MAC V1.0' },
            { value: '1.0.1', label: 'MAC V1.0.1' },
            { value: '1.0.2', label: 'MAC V1.0.2' },
            { value: '1.0.3', label: 'MAC V1.0.3' },
            { value: '1.0.4', label: 'MAC V1.0.4' },
            { value: '1.1.0', label: 'MAC V1.1' },
          ]}
        />
        <Form.Field
          title={sharedMessages.phyVersion}
          description={sharedMessages.phyVersionDescription}
          name="lorawan_phy_version"
          component={Select}
          required
          options={[
            { value: '1.0.0', label: 'PHY V1.0' },
            { value: '1.0.1', label: 'PHY V1.0.1' },
            { value: '1.0.2-a', label: 'PHY V1.0.2 REV A' },
            { value: '1.0.2-b', label: 'PHY V1.0.2 REV B' },
            { value: '1.0.3-a', label: 'PHY V1.0.3 REV A' },
            { value: '1.1.0-a', label: 'PHY V1.1 REV A' },
            { value: '1.1.0-b', label: 'PHY V1.1 REV B' },
          ]}
        />
        <NsFrequencyPlansSelect name="frequency_plan_id" required />
        <Form.Field title={m.supportsClassC} name="supports_class_c" component={Checkbox} />
        <Form.Field
          title={sharedMessages.networkServerAddress}
          placeholder={sharedMessages.addressPlaceholder}
          name="network_server_address"
          component={Input}
        />
        <Form.Field
          title={sharedMessages.applicationServerAddress}
          placeholder={sharedMessages.addressPlaceholder}
          name="application_server_address"
          component={Input}
        />
        <Message component="h4" content={m.activationSettings} />
        <Form.Field
          title={m.activationMode}
          name="_activation_mode"
          component={Radio.Group}
          disabled={!mayEditKeys}
        >
          <Radio label={m.otaa} value="otaa" onChange={this.handleOTAASelect} />
          <Radio label={m.abp} value="abp" onChange={this.handleABPSelect} />
        </Form.Field>
        {otaa ? this.OTAASection : this.ABPSection}
        <SubmitBar>
          <Form.Submit component={SubmitButton} message={m.createDevice} />
        </SubmitBar>
      </Form>
    )
  }
}

export default DeviceDataForm
