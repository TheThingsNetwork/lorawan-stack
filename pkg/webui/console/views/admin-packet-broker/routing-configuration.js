// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState } from 'react'
import { Route, Routes } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'
import { isEqual } from 'lodash'

import Form from '@ttn-lw/components/form'
import Radio from '@ttn-lw/components/radio-button'
import Tabs from '@ttn-lw/components/tabs'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import SubViewErrorComponent from '@console/views/sub-view-error'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import Yup from '@ttn-lw/lib/yup'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { isValidPolicy } from '@console/lib/packet-broker/utils'

import {
  deleteAllHomeNetworkRoutingPolicies,
  deleteHomeNetworkDefaultRoutingPolicy,
  getHomeNetworkDefaultRoutingPolicy,
  getHomeNetworkRoutingPolicies,
  setHomeNetworkDefaultRoutingPolicy,
  setHomeNetworkRoutingPolicy,
} from '@console/store/actions/packet-broker'

import {
  selectHomeNetworkDefaultRoutingPolicy,
  selectPacketBrokerHomeNetworkPoliciesStore,
} from '@console/store/selectors/packet-broker'

import DefaultRoutingPolicyView from './default-routing-policy'
import NetworkRoutingPoliciesView from './network-routing-policies'
import m from './messages'

const validationSchema = Yup.object({
  _routing_configuration: Yup.string().oneOf(['all_networks', 'ttn', 'custom']),
  policy: Yup.object({
    uplink: Yup.object({}),
    downlink: Yup.object({}),
  }),
})

const fullyPermissiveNetworkPolicy = {
  uplink: {
    join_request: true,
    mac_data: true,
    application_data: true,
    signal_quality: true,
    localization: true,
  },
  downlink: {
    join_accept: true,
    mac_data: true,
    application_data: true,
  },
}

const TTN_NET_ID = '19'
const peerWithEveryNetwork = policy =>
  isEqual(policy.uplink, fullyPermissiveNetworkPolicy.uplink) &&
  isEqual(policy.downlink, fullyPermissiveNetworkPolicy.downlink)

const onlyTtn = policies =>
  Object.keys(policies).length === 1 && Object.keys(policies).includes(`${TTN_NET_ID}/ttn`)

const RoutingConfigurationView = () => {
  const dispatch = useDispatch()
  const [activeTab, setActiveTab] = useState('default-routing-policy')
  const tabs = [
    {
      title: m.defaultRoutingPolicy,
      link: '/admin-panel/packet-broker/routing-configuration',
      name: 'default',
    },
    {
      title: sharedMessages.networks,
      link: '/admin-panel/packet-broker/routing-configuration/networks',
      name: 'networks',
      exact: false,
    },
  ]

  const defaultRoutingPolicy = useSelector(selectHomeNetworkDefaultRoutingPolicy)
  const routingPolicies = useSelector(selectPacketBrokerHomeNetworkPoliciesStore)

  const initialValues = {
    _routing_configuration:
      peerWithEveryNetwork(defaultRoutingPolicy) &&
      Object.keys(routingPolicies).length === 0 &&
      isValidPolicy(defaultRoutingPolicy)
        ? 'all_networks'
        : onlyTtn(routingPolicies) && !isValidPolicy(defaultRoutingPolicy)
        ? 'ttn'
        : 'custom',
    _use_default_policy: isValidPolicy(defaultRoutingPolicy),
  }
  initialValues.policy = isValidPolicy(defaultRoutingPolicy)
    ? defaultRoutingPolicy
    : { uplink: {}, downlink: {} }

  const [routingConfig, setRoutingConfig] = useState(undefined)
  const [formError, setFormError] = useState(undefined)

  const handleDefaultRoutingPolicySubmit = useCallback(
    async values => {
      const vals = validationSchema.cast(values)
      const { _routing_configuration, _use_default_policy, policy } = vals
      const ids = Object.keys(routingPolicies)

      try {
        if (_routing_configuration === 'ttn') {
          await dispatch(attachPromise(deleteHomeNetworkDefaultRoutingPolicy()))
          await dispatch(attachPromise(deleteAllHomeNetworkRoutingPolicies(ids)))
          await dispatch(attachPromise(setHomeNetworkRoutingPolicy(`${TTN_NET_ID}/ttn`, policy)))
        } else if (_routing_configuration === 'all_networks') {
          await dispatch(attachPromise(deleteAllHomeNetworkRoutingPolicies(ids)))
          await dispatch(attachPromise(setHomeNetworkDefaultRoutingPolicy(policy)))
        } else if (_routing_configuration === 'custom' && _use_default_policy) {
          await dispatch(attachPromise(setHomeNetworkDefaultRoutingPolicy(policy)))
        } else if (_routing_configuration === 'custom' && !_use_default_policy) {
          await dispatch(attachPromise(deleteHomeNetworkDefaultRoutingPolicy()))
        }

        toast({
          message: m.defaultRoutingPolicySet,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setFormError(error)
      }
    },
    [dispatch, setFormError, routingPolicies],
  )

  const handleRoutingConfigChange = useCallback(
    value => {
      setRoutingConfig(value)
    },
    [setRoutingConfig],
  )

  const handleSetPolicies = useCallback(({ setValues }, { value }) => {
    if (value !== 'custom') {
      return setValues(values => ({
        ...values,
        _routing_configuration: value,
        policy: fullyPermissiveNetworkPolicy,
      }))
    }

    return setValues(values => ({
      ...values,
      _routing_configuration: value,
    }))
  }, [])

  const showPolicyCheckboxes = routingConfig
    ? routingConfig === 'custom'
    : initialValues._routing_configuration === 'custom'

  return (
    <RequireRequest
      requestAction={[getHomeNetworkDefaultRoutingPolicy(), getHomeNetworkRoutingPolicies()]}
      errorRenderFunction={SubViewErrorComponent}
      spinnerProps={{ inline: true, center: true, className: 'mt-ls-s' }}
    >
      <Form
        onSubmit={handleDefaultRoutingPolicySubmit}
        initialValues={initialValues}
        error={formError}
        className="mt-cs-l"
      >
        <Form.Field
          component={Radio.Group}
          className="mb-cs-xl"
          name="_routing_configuration"
          onChange={handleRoutingConfigChange}
          valueSetter={handleSetPolicies}
        >
          <Radio
            label="Forward traffic to all networks registered in Packet Broker"
            value="all_networks"
          />
          <Radio
            label="Forward traffic to The Things Stack Sandbox (community network) only"
            value="ttn"
          />
          <Radio label="Use custom routing policies" value="custom" />
        </Form.Field>
        {showPolicyCheckboxes && (
          <>
            <Tabs tabs={tabs} active={activeTab} onTabChange={setActiveTab} divider />
            <Routes>
              <Route index Component={DefaultRoutingPolicyView} />
              <Route path="networks/*" Component={NetworkRoutingPoliciesView} />
              <Route path="*" component={GenericNotFound} />
            </Routes>
          </>
        )}
        <SubmitBar align="end">
          <Form.Submit component={SubmitButton} message={'Save routing configuration'} />
        </SubmitBar>
      </Form>
    </RequireRequest>
  )
}

export default RoutingConfigurationView
