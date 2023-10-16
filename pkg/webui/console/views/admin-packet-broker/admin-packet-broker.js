// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Col, Row } from 'react-grid-system'
import { useSelector, useDispatch } from 'react-redux'
import { Routes, Route } from 'react-router-dom'
import classnames from 'classnames'

import PacketBrokerLogo from '@assets/misc/packet-broker.svg'

import Link from '@ttn-lw/components/link'
import PageTitle from '@ttn-lw/components/page-title'
import Switch from '@ttn-lw/components/switch'
import Tabs from '@ttn-lw/components/tabs'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import ErrorNotification from '@ttn-lw/components/error-notification'
import Radio from '@ttn-lw/components/radio-button'
import Form from '@ttn-lw/components/form'
import toast from '@ttn-lw/components/toast'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import SubViewErrorComponent from '@console/views/sub-view-error'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import Yup from '@ttn-lw/lib/yup'

import { isNotEnabledError, isValidPolicy } from '@console/lib/packet-broker/utils'

import {
  registerPacketBroker,
  deregisterPacketBroker,
  getHomeNetworkDefaultRoutingPolicy,
  setHomeNetworkDefaultRoutingPolicy,
  deleteHomeNetworkDefaultRoutingPolicy,
} from '@console/store/actions/packet-broker'

import {
  selectRegistered,
  selectRegisterEnabled,
  selectEnabled,
  selectListed,
  selectInfoError,
  selectHomeNetworkDefaultRoutingPolicy,
} from '@console/store/selectors/packet-broker'

import DefaultRoutingPolicyView from './default-routing-policy'
import NetworkRoutingPoliciesView from './network-routing-policies'
import m from './messages'

import style from './admin-packet-broker.styl'

const validationSchema = Yup.object({
  _use_default_policy: Yup.bool(),
  policy: Yup.object({
    uplink: Yup.object({}),
    downlink: Yup.object({}),
  }).when('_use_default_policy', { is: 'default', then: schema => schema.strip() }),
})

const PacketBroker = () => {
  const [activeTab, setActiveTab] = useState('default-routing-policy')
  const [deregisterModalVisible, setDeregisterModalVisible] = useState(false)
  const registered = useSelector(selectRegistered)
  const registerEnabled = useSelector(selectRegisterEnabled)
  const enabled = useSelector(selectEnabled)
  const [unlistModalVisible, setUnlistModalVisible] = useState(false)
  const listed = useSelector(selectListed)
  const infoError = useSelector(selectInfoError)
  const dispatch = useDispatch()
  const showError = Boolean(infoError) && !isNotEnabledError(infoError)

  const handleRegisterChange = useCallback(() => {
    if (!registered) {
      dispatch(registerPacketBroker({}))
    } else {
      setDeregisterModalVisible(true)
    }
  }, [dispatch, registered, setDeregisterModalVisible])

  const handleDeregisterModalComplete = useCallback(
    approved => {
      setDeregisterModalVisible(false)
      if (approved) {
        dispatch(deregisterPacketBroker())
      }
    },
    [dispatch, setDeregisterModalVisible],
  )

  const handleListedChange = useCallback(() => {
    if (!listed) {
      dispatch(registerPacketBroker({ listed: true }))
    } else {
      setUnlistModalVisible(true)
    }
  }, [dispatch, listed, setUnlistModalVisible])

  const handleUnlistModalComplete = useCallback(
    approved => {
      setUnlistModalVisible(false)
      if (approved) {
        dispatch(registerPacketBroker({ listed: false }))
      }
    },
    [dispatch, setUnlistModalVisible],
  )

  const tabs = [
    { title: m.defaultRoutingPolicy, link: '/admin-panel/packet-broker', name: 'default' },
    /*     {
      title: m.defaultGatewayVisibility,
      link: '/admin-panel/packet-broker/default-gateway-visibility',
      name: 'default-gateway-visibility',
    }, */
    {
      title: sharedMessages.networks,
      link: '/admin-panel/packet-broker/networks',
      name: 'networks',
      exact: false,
    },
  ]

  const defaultRoutingPolicy = useSelector(selectHomeNetworkDefaultRoutingPolicy)
  const initialValues = {
    _use_default_policy: isValidPolicy(defaultRoutingPolicy),
    _routing_configuration: 'default',
  }
  initialValues.policy = initialValues._use_default_policy
    ? defaultRoutingPolicy
    : { uplink: {}, downlink: {} }
  const [rountingConfig, setRoutingConfig] = useState(initialValues._routing_configuration)
  const [formError, setFormError] = useState(undefined)
  const handleDefaultRoutingPolicySubmit = useCallback(
    async values => {
      const vals = validationSchema.cast(values)
      const { _use_default_policy, policy } = vals
      try {
        if (_use_default_policy) {
          await dispatch(attachPromise(setHomeNetworkDefaultRoutingPolicy(policy)))
        } else {
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
    [dispatch, setFormError],
  )

  const handleRoutingConfigChange = useCallback(
    value => {
      setRoutingConfig(value)
    },
    [setRoutingConfig],
  )

  return (
    <Container>
      <Row>
        <Col md={12}>
          <PageTitle title={sharedMessages.packetBroker} />
          <div className={style.introduction}>
            <Message content={m.packetBrokerInfoText} className={style.info} />
            <img className={style.logo} src={PacketBrokerLogo} alt="Packet Broker" />
          </div>
          <div>
            <Message component="h4" content={m.learnMore} className={style.furtherResources} />
            <Link.DocLink path="/reference/packet-broker/" secondary>
              Packet Broker
            </Link.DocLink>
            {' | '}
            <Link.Anchor href="https://www.packetbroker.net" external secondary>
              <Message content={m.packetBrokerWebsite} />
            </Link.Anchor>
          </div>
          <hr className={style.hRule} />
          <Message content={m.whyNetworkPeeringTitle} component="h3" />
          <Message content={m.whyNetworkPeeringText} className={style.info} component="p" />
          <Message content={m.enbaling} className={style.info} />
          <Message content={m.registrationStatus} component="h3" className="mt-cs-xxl" />
          {showError && <ErrorNotification small content={infoError} />}
          {enabled && (
            <Row gutterWidth={48} className="mb-cs-xl">
              <Col md={4}>
                {registerEnabled && (
                  <label
                    className={classnames(style.toggleContainer, {
                      [style.disabled]: !enabled || !registerEnabled,
                    })}
                  >
                    <Message content={m.registerNetwork} component="span" />
                    <Switch
                      onChange={handleRegisterChange}
                      checked={registered}
                      className={style.toggle}
                      disabled={!enabled}
                    />
                  </label>
                )}
              </Col>
              <Col md={8} className={style.switchInfo}>
                <Message
                  content={
                    registerEnabled
                      ? m.packetBrokerRegistrationDesc
                      : m.packetBrokerRegistrationDisabledDesc
                  }
                  component="span"
                  className={style.description}
                />
              </Col>
            </Row>
          )}
          <PortalledModal
            visible={deregisterModalVisible}
            title={m.confirmDeregister}
            buttonMessage={m.deregisterNetwork}
            onComplete={handleDeregisterModalComplete}
            danger
            approval
          >
            <Message
              content={m.deregisterModal}
              values={{ lineBreak: <br />, b: chunks => <b>{chunks}</b> }}
              component="span"
            />
          </PortalledModal>
        </Col>
        {registered && (
          <>
            <Col md={12}>
              <Row gutterWidth={48}>
                <Col md={4}>
                  <label className={style.toggleContainer}>
                    <Message content={m.listNetwork} component="span" />
                    <Switch
                      onChange={handleListedChange}
                      checked={listed}
                      className={style.toggle}
                    />
                  </label>
                </Col>
                <Col md={8} className={style.switchInfo}>
                  <Message className={style.description} content={m.listNetworkDesc} />
                </Col>
              </Row>
              <PortalledModal
                visible={unlistModalVisible}
                title={m.confirmUnlist}
                buttonMessage={m.unlistNetwork}
                onComplete={handleUnlistModalComplete}
                danger
                approval
              >
                <Message
                  content={m.unlistModal}
                  values={{ lineBreak: <br />, b: chunks => <b>{chunks}</b> }}
                  component="span"
                />
              </PortalledModal>
            </Col>
            <Col md={12} style={{ position: 'relative' }}>
              <Message
                content={'Routing configuration'}
                component="h3"
                className={style.subTitle}
              />
              <RequireRequest
                requestAction={getHomeNetworkDefaultRoutingPolicy()}
                errorRenderFunction={SubViewErrorComponent}
                spinnerProps={{ inline: true, center: true, className: 'mt-ls-s' }}
              >
                <Form
                  onSubmit={handleDefaultRoutingPolicySubmit}
                  initialValues={initialValues}
                  error={formError}
                >
                  <Form.Field
                    component={Radio.Group}
                    className="mb-cs-xl"
                    name="_routing_configuration"
                    onChange={handleRoutingConfigChange}
                  >
                    <Radio
                      label="Forward traffic to all networks registered in Packet Broker"
                      value="default"
                    />
                    <Radio
                      label="Forward traffic to The Things Stack Sandbox (community network) only"
                      value="ttn"
                    />
                    <Radio label="Use custom routing policies" value="custom" />
                  </Form.Field>
                  {rountingConfig === 'custom' && (
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
            </Col>
          </>
        )}
      </Row>
    </Container>
  )
}

export default PacketBroker
