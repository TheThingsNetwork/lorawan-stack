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
import Icon from '@ttn-lw/components/icon'
import Switch from '@ttn-lw/components/switch'
import Tabs from '@ttn-lw/components/tabs'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import Notification from '@ttn-lw/components/notification'
import ErrorNotification from '@ttn-lw/components/error-notification'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import SubViewErrorComponent from '@console/views/sub-view-error'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { isNotEnabledError } from '@console/lib/packet-broker/utils'

import {
  registerPacketBroker,
  deregisterPacketBroker,
  getHomeNetworkDefaultRoutingPolicy,
  getHomeNetworkDefaultGatewayVisibility,
} from '@console/store/actions/packet-broker'

import {
  selectRegistered,
  selectRegisterEnabled,
  selectEnabled,
  selectListed,
  selectInfo,
  selectInfoError,
} from '@console/store/selectors/packet-broker'

import DefaultRoutingPolicyView from './default-routing-policy'
import NetworkRoutingPoliciesView from './network-routing-policies'
import DefaultGatewayVisibilityView from './default-gateway-visibility'
import m from './messages'

import style from './admin-packet-broker.styl'

const PacketBroker = () => {
  const [activeTab, setActiveTab] = useState('default-routing-policy')
  const [deregisterModalVisible, setDeregisterModalVisible] = useState(false)
  const registered = useSelector(selectRegistered)
  const registerEnabled = useSelector(selectRegisterEnabled)
  const enabled = useSelector(selectEnabled)
  const [unlistModalVisible, setUnlistModalVisible] = useState(false)
  const listed = useSelector(selectListed)
  const info = useSelector(selectInfo)
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
    { title: m.defaultRoutingPolicy, link: '/admin/packet-broker', name: 'default' },
    {
      title: m.defaultGatewayVisibility,
      link: '/admin/packet-broker/default-gateway-visibility',
      name: 'default-gateway-visibility',
    },
    {
      title: sharedMessages.networks,
      link: '/admin/packet-broker/networks',
      name: 'networks',
      exact: false,
    },
  ]

  const boldMessage = { b: msg => <b>{msg}</b> }

  return (
    <Container>
      <Row>
        <Col lg={8} md={12}>
          <PageTitle title={sharedMessages.packetBroker} />
          <div className={style.introduction}>
            <Message content={m.packetBrokerInfoText} className={style.info} />
            <img className={style.logo} src={PacketBrokerLogo} alt="Packet Broker" />
          </div>
          <div>
            <Message
              component="h4"
              content={sharedMessages.furtherResources}
              className={style.furtherResources}
            />
            <Link.DocLink path="/reference/packet-broker/" secondary>
              Packet Broker documentation
            </Link.DocLink>
            {' | '}
            <Link.Anchor href="https://www.packetbroker.net" external secondary>
              <Message content={m.packetBrokerWebsite} />
            </Link.Anchor>
            {' | '}
            <Link.Anchor href="https://status.packetbroker.net" external secondary>
              <Message content={m.packetBrokerStatusPage} />
            </Link.Anchor>
          </div>
          <hr className={style.hRule} />
          <Message content={m.registrationStatus} component="h3" />
          {!enabled && <Notification warning small content={m.packetBrokerDisabledDesc} />}
          {showError && <ErrorNotification small content={infoError} />}
          {enabled && (
            <Row gutterWidth={48}>
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
                {registered && (
                  <div className={style.featureInfo}>
                    {info.forwarder_enabled ? (
                      <span data-test-id="feature-info-forwarder-enabled">
                        <Icon icon="check" className="c-active" textPaddedRight />
                        <Message
                          content={m.forwarderEnabled}
                          values={boldMessage}
                          component="span"
                        />
                      </span>
                    ) : (
                      <span data-test-id="feature-info-forwarder-disabled">
                        <Icon icon="close" className="c-error" textPaddedRight />
                        <Message
                          content={m.forwarderDisabled}
                          values={boldMessage}
                          component="span"
                        />
                      </span>
                    )}
                    {info.home_network_enabled ? (
                      <span data-test-id="feature-info-home-network-enabled">
                        <Icon icon="check" className="c-active" textPaddedRight />
                        <Message
                          content={m.homeNetworkEnabled}
                          values={boldMessage}
                          component="span"
                        />
                      </span>
                    ) : (
                      <span data-test-id="feature-info-forwarder-disabled">
                        <Icon icon="close" className="c-error" textPaddedRight />
                        <Message
                          content={m.homeNetworkDisabled}
                          values={boldMessage}
                          component="span"
                        />
                      </span>
                    )}
                  </div>
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
            <Col lg={8} md={12}>
              <Message content={m.networkVisibility} component="h3" className={style.subTitle} />
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
              <hr className={style.hRule} />
            </Col>
            <Col md={12}>
              <Tabs tabs={tabs} active={activeTab} onTabChange={setActiveTab} divider />
              <RequireRequest
                requestAction={[
                  getHomeNetworkDefaultRoutingPolicy(),
                  getHomeNetworkDefaultGatewayVisibility(),
                ]}
                errorRenderFunction={SubViewErrorComponent}
              >
                <Routes>
                  <Route index Component={DefaultRoutingPolicyView} />
                  <Route
                    path="default-gateway-visibility"
                    Component={DefaultGatewayVisibilityView}
                  />
                  <Route path="networks/*" Component={NetworkRoutingPoliciesView} />
                  <Route path="*" component={GenericNotFound} />
                </Routes>
              </RequireRequest>
            </Col>
          </>
        )}
      </Row>
    </Container>
  )
}

export default PacketBroker
