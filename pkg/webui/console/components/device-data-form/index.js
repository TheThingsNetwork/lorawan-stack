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

import {
  parseLorawanMacVersion,
  generate16BytesKey,
  ACTIVATION_MODES,
  LORAWAN_VERSIONS,
  LORAWAN_PHY_VERSIONS,
} from '@console/lib/device-utils'

import m from './messages'
import validationSchema from './validation-schema'

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
    const { setValues, values } = this.formRef.current

    // Reset Join Server related entries if the device is provisined by an
    // external JS.
    if (external_js) {
      setValues({
        ...values,
        root_keys: {
          nwk_key: {},
          app_key: {},
        },
        resets_join_nonces: false,
        join_server_address: undefined,
        _external_js: external_js,
      })
    } else {
      let join_server_address = values.join_server_address

      // Reset `join_server_address` if is present after disabling external JS
      // provisioning.
      if (jsConfig.enabled && !Boolean(join_server_address)) {
        join_server_address = new URL(jsConfig.base_url).hostname
      }

      setValues({
        ...values,
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

    const lwVersion = parseLorawanMacVersion(lorawan_version)

    return (
      <>
        <Form.Field
          title={sharedMessages.devEUI}
          name="ids.dev_eui"
          type="byte"
          min={8}
          max={8}
          description={sharedMessages.deviceEUIDescription}
          required={lwVersion === 104}
          component={Input}
        />
        <DevAddrInput
          title={sharedMessages.devAddr}
          name="session.dev_addr"
          description={sharedMessages.deviceAddrDescription}
          required
        />
        <Form.Field
          title={sharedMessages.nwkSKey}
          name="session.keys.f_nwk_s_int_key.key"
          type="byte"
          min={16}
          max={16}
          description={sharedMessages.nwkSKeyDescription}
          component={Input.Generate}
          onGenerateValue={generate16BytesKey}
          required
        />
        <Form.Field
          title={sharedMessages.appSKey}
          name="session.keys.app_s_key.key"
          type="byte"
          min={16}
          max={16}
          description={sharedMessages.appSKeyDescription}
          component={Input.Generate}
          onGenerateValue={generate16BytesKey}
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
              description={sharedMessages.sNwkSIKeyDescription}
              component={Input.Generate}
              onGenerateValue={generate16BytesKey}
              required
            />
            <Form.Field
              title={sharedMessages.nwkSEncKey}
              name="session.keys.nwk_s_enc_key.key"
              type="byte"
              min={16}
              max={16}
              description={sharedMessages.nwkSEncKeyDescription}
              component={Input.Generate}
              onGenerateValue={generate16BytesKey}
              required
            />
          </>
        )}
        <Form.Field
          title={sharedMessages.resetsFCnt}
          onChange={this.handleResetsFrameCountersChange}
          warning={resets_f_cnt ? sharedMessages.resetWarning : undefined}
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
          description={sharedMessages.joinEUIDescription}
          required
          showPrefixes
        />
        <Form.Field
          title={sharedMessages.devEUI}
          name="ids.dev_eui"
          type="byte"
          min={8}
          max={8}
          description={sharedMessages.deviceEUIDescription}
          required
          component={Input}
        />
        <Form.Field
          title={sharedMessages.externalJoinServer}
          description={sharedMessages.externalJoinServerDescription}
          name="_external_js"
          onChange={this.handleExternalJoinServerChange}
          component={Checkbox}
        />
        <Form.Field
          title={sharedMessages.joinServerAddress}
          placeholder={external_js ? sharedMessages.external : sharedMessages.addressPlaceholder}
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
              description={sharedMessages.appKeyDescription}
              placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
              component={Input.Generate}
              disabled={external_js}
              onGenerateValue={generate16BytesKey}
              mayGenerateValue={!external_js && mayEditKeys}
            />
            <Form.Field
              title={sharedMessages.nwkKey}
              name="root_keys.nwk_key.key"
              type="byte"
              min={16}
              max={16}
              description={sharedMessages.nwkKeyDescription}
              placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
              component={Input.Generate}
              disabled={external_js}
              onGenerateValue={generate16BytesKey}
              mayGenerateValue={!external_js && mayEditKeys}
            />
          </>
        )}
        <Form.Field
          title={sharedMessages.resetsJoinNonces}
          onChange={this.handleResetsJoinNoncesChange}
          warning={resets_join_nonces ? sharedMessages.resetWarning : undefined}
          name="resets_join_nonces"
          component={Checkbox}
          disabled={external_js}
        />
        <Form.Field
          title={sharedMessages.homeNetID}
          description={sharedMessages.homeNetIDDescription}
          placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
          name="net_id"
          type="byte"
          min={3}
          max={3}
          component={Input}
          disabled={external_js}
        />
        <Form.Field
          title={sharedMessages.asServerID}
          name="application_server_id"
          description={sharedMessages.asServerIDDescription}
          placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
          component={Input}
          disabled={external_js}
        />
        <Form.Field
          title={sharedMessages.asServerKekLabel}
          name="application_server_kek_label"
          description={sharedMessages.asServerKekLabelDescription}
          placeholder={external_js ? sharedMessages.provisionedOnExternalJoinServer : undefined}
          component={Input}
          disabled={external_js}
        />
        <Form.Field
          title={sharedMessages.nsServerKekLabel}
          name="network_server_kek_label"
          description={sharedMessages.nsServerKekLabelDescription}
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
      _activation_mode: ACTIVATION_MODES.OTAA,
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
      _activation_mode: otaa ? ACTIVATION_MODES.OTAA : ACTIVATION_MODES.ABP,
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
          placeholder={sharedMessages.deviceIdPlaceholder}
          autoFocus
          required
          component={Input}
        />
        <Form.Field
          title={sharedMessages.devName}
          name="name"
          placeholder={sharedMessages.deviceNamePlaceholder}
          component={Input}
        />
        <Form.Field
          title={sharedMessages.devDesc}
          name="description"
          type="textarea"
          placeholder={sharedMessages.deviceDescPlaceholder}
          description={sharedMessages.deviceDescDescription}
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
          options={LORAWAN_VERSIONS}
        />
        <Form.Field
          title={sharedMessages.phyVersion}
          description={sharedMessages.phyVersionDescription}
          name="lorawan_phy_version"
          component={Select}
          required
          options={LORAWAN_PHY_VERSIONS}
        />
        <NsFrequencyPlansSelect name="frequency_plan_id" required />
        <Form.Field
          title={sharedMessages.supportsClassC}
          name="supports_class_c"
          component={Checkbox}
        />
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
          title={sharedMessages.activationMode}
          name="_activation_mode"
          component={Radio.Group}
          disabled={!mayEditKeys}
        >
          <Radio label={sharedMessages.otaa} value="otaa" onChange={this.handleOTAASelect} />
          <Radio label={sharedMessages.abp} value="abp" onChange={this.handleABPSelect} />
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
