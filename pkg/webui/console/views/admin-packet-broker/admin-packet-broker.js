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
import { Switch as RouteSwitch, Route } from 'react-router'
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
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { isNotEnabledError } from '@console/lib/packet-broker/utils'

import {
  registerPacketBroker,
  deregisterPacketBroker,
  getHomeNetworkDefaultRoutingPolicy,
} from '@console/store/actions/packet-broker'

import {
  selectRegistered,
  selectInfo,
  selectInfoError,
  selectEnabled,
} from '@console/store/selectors/packet-broker'

import DefaultRoutingPolicyView from './default-routing-policy'
import NetworkRoutingPoliciesView from './network-routing-policies'
import m from './messages'

import style from './admin-packet-broker.styl'

const PacketBroker = ({ match }) => {
  const [activeTab, setActiveTab] = useState('default-routing-policy')
  const [modalVisible, setModalVisible] = useState(false)
  const registered = useSelector(selectRegistered)
  const enabled = useSelector(selectEnabled)
  const info = useSelector(selectInfo)
  const infoError = useSelector(selectInfoError)
  const dispatch = useDispatch()
  const { url } = match
  const showError = Boolean(infoError) && !isNotEnabledError(infoError)

  const handleRegisterChange = useCallback(() => {
    if (!registered) {
      dispatch(registerPacketBroker())
    } else {
      setModalVisible(true)
    }
  }, [dispatch, registered, setModalVisible])

  const handleModalComplete = useCallback(
    approved => {
      setModalVisible(false)
      if (approved) {
        dispatch(deregisterPacketBroker())
      }
    },
    [dispatch, setModalVisible],
  )

  const tabs = React.useMemo(
    () => [
      { title: m.defaultRoutingPolicy, link: url, name: 'default' },
      { title: sharedMessages.networks, link: `${url}/networks`, name: 'networks', exact: false },
    ],
    [url],
  )

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
              Packet Broker
            </Link.DocLink>
            {' | '}
            <Link.Anchor href="https://www.packetbroker.org" external secondary>
              <Message content={m.packetBrokerWebsite} />
            </Link.Anchor>
          </div>
          <hr className={style.hRule} />
          <Message content={m.registerThisNetwork} component="h3" />
          {!enabled && <Notification info small content={m.packetBrokerDisabledDesc} />}
          {showError && <ErrorNotification small content={infoError} />}
          <label className={classnames(style.toggleContainer, { [style.disabled]: !enabled })}>
            <Message content={m.registerNetwork} component="span" />
            <Switch
              onChange={handleRegisterChange}
              checked={registered}
              className={style.toggle}
              disabled={!enabled}
            />
          </label>
          <PortalledModal
            visible={modalVisible}
            title={m.confirmDeregister}
            buttonMessage={m.deregisterNetwork}
            onComplete={handleModalComplete}
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
        {enabled && registered ? (
          <Col md={12}>
            <div className={style.featureInfo}>
              {info.forwarder_enabled ? (
                <span>
                  <Icon icon="check" />
                  <Message content={m.forwarderEnabled} />
                </span>
              ) : (
                <span>
                  <Icon icon="close" className="c-error" />
                  <Message content={m.forwarderDisabled} />
                </span>
              )}
              {info.home_network_enabled ? (
                <span>
                  <Icon icon="check" />
                  <Message content={m.homeNetworkEnabled} />
                </span>
              ) : (
                <span>
                  <Icon icon="close" className="c-error" />
                  <Message content={m.homeNetworkDisabled} />
                </span>
              )}
            </div>
            <Tabs tabs={tabs} active={activeTab} onTabChange={setActiveTab} divider />
            <RequireRequest requestAction={getHomeNetworkDefaultRoutingPolicy()}>
              <RouteSwitch>
                <Route path={url} exact component={DefaultRoutingPolicyView} />
                <Route path={`${url}/networks`} exact component={NetworkRoutingPoliciesView} />
                <NotFoundRoute />
              </RouteSwitch>
            </RequireRequest>
          </Col>
        ) : (
          <Col lg={8} md={12}>
            <Message content={m.packetBrokerRegistrationDesc} component="p" />
          </Col>
        )}
      </Row>
    </Container>
  )
}

PacketBroker.propTypes = {
  match: PropTypes.match.isRequired,
}

export default PacketBroker
