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
import { useSelector, useDispatch } from 'react-redux'
import { Container, Col, Row } from 'react-grid-system'
import { useParams } from 'react-router-dom'

import DataSheet from '@ttn-lw/components/data-sheet'
import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import toast from '@ttn-lw/components/toast'
import SafeInspector from '@ttn-lw/components/safe-inspector'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import RoutingPolicy from '@console/components/routing-policy'
import RoutingPolicyForm from '@console/components/routing-policy-form'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { isValidPolicy } from '@console/lib/packet-broker/utils'

import {
  getPacketBrokerNetwork,
  setHomeNetworkRoutingPolicy,
  getHomeNetworkDefaultRoutingPolicy,
  deleteHomeNetworkRoutingPolicy,
} from '@console/store/actions/packet-broker'

import {
  selectPacketBrokerHomeNetworkPolicyById,
  selectPacketBrokerForwarderPolicyById,
  selectPacketBrokerNetworkById,
  selectHomeNetworkDefaultRoutingPolicy,
  selectRegistered,
} from '@console/store/selectors/packet-broker'

import m from './messages'

import style from './admin-packet-broker.styl'

const NetworkRoutingPolicyViewInner = () => {
  const { netId, tenantId } = useParams()
  const dispatch = useDispatch()
  const displayNetId = parseInt(netId).toString(16).padStart(6, '0')
  const combinedId = tenantId ? `${netId}/${tenantId}` : netId
  const displayId = tenantId ? `${displayNetId}/${tenantId}` : displayNetId
  const [formError, setFormError] = useState(undefined)

  const network = useSelector(state => selectPacketBrokerNetworkById(state, combinedId))
  const homeNetwork = useSelector(state =>
    selectPacketBrokerHomeNetworkPolicyById(state, combinedId),
  )
  const defaultRoutingPolicy = useSelector(selectHomeNetworkDefaultRoutingPolicy)

  useBreadcrumbs(
    'admin-panel.packet-broker.routing-configuration.networks.single',
    <>
      <Breadcrumb
        path={'/admin-panel/packet-broker/routing-configuration/networks'}
        content={sharedMessages.networks}
      />
      <Breadcrumb
        path={`/admin-panel/packet-broker/routing-configuration/networks/${netId}${
          tenantId ? `/${tenantId}` : ''
        }`}
        content={network.name || displayId}
      />
    </>,
  )

  const handleRoutingPolicySubmit = useCallback(
    async value => {
      try {
        if (value._use_default_policy) {
          await dispatch(attachPromise(deleteHomeNetworkRoutingPolicy(combinedId)))
        } else {
          await dispatch(attachPromise(setHomeNetworkRoutingPolicy(combinedId, value.policy)))
        }
        toast({
          message: m.routingPolicySet,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setFormError(error)
      }
    },
    [dispatch, setFormError, combinedId],
  )

  const forwarder = useSelector(state =>
    selectPacketBrokerForwarderPolicyById(state, combinedId),
  ) || { policy: {} }
  const initialValues = {
    _use_default_policy: !Boolean(homeNetwork.uplink) && !Boolean(homeNetwork.downlink),
    policy: {
      uplink: {},
      downlink: {},
    },
  }
  if (isValidPolicy(homeNetwork)) {
    initialValues.policy = { uplink: homeNetwork.uplink, downlink: homeNetwork.downlink }
  }
  const hasDevAddrBlocks =
    network.dev_addr_blocks instanceof Array && network.dev_addr_blocks.length !== 0

  const homeNetworkData = [
    {
      header: sharedMessages.networkInformation,
      items: [
        {
          key: m.networkId,
          value: (
            <SafeInspector
              data={displayId}
              small
              hideable={false}
              isBytes={false}
              initiallyVisible
            />
          ),
        },
        {
          key: sharedMessages.name,
          value: network.name,
        },
        {
          key: sharedMessages.contactInformation,
          value: 'contact_info' in network ? network.contact_info.value : undefined,
        },
        {
          key: m.lastPolicyChange,
          value:
            Boolean(forwarder) && forwarder.updated_at ? (
              <DateTime value={forwarder.updated_at} />
            ) : (
              <Message content={m.noPolicySet} />
            ),
        },
      ],
    },
  ]

  return (
    <Container>
      <Row>
        <Col md={12}>
          <PageTitle title={m.network} values={{ network: network.name || displayId }}>
            <Link
              to="/admin-panel/packet-broker/routing-configuration/networks"
              secondary
              className={style.backLink}
            >
              ← <Message content={m.backToAllNetworks} />
            </Link>
          </PageTitle>
        </Col>
        <Col md={6}>
          <Message content={sharedMessages.generalInformation} component="h3" />
          <DataSheet data={homeNetworkData} />
          {hasDevAddrBlocks && (
            <>
              <Message content={m.devAddressBlocks} component="h4" />
              {network.dev_addr_blocks.map(b => (
                <div
                  className={style.deviceAddressBlockRow}
                  key={`${b.dev_addr_prefix.dev_addr}/${b.dev_addr_prefix.length}`}
                >
                  <div>
                    <span>Prefix:</span>
                    <SafeInspector
                      data={`${b.dev_addr_prefix.dev_addr}/${b.dev_addr_prefix.length}`}
                      small
                      hideable={false}
                      isBytes={false}
                      initiallyVisible
                      disableResize
                    />
                  </div>
                  {b.home_network_cluster_id && (
                    <div>
                      <span>
                        <Message content={m.homeNetworkClusterId} />:
                      </span>
                      <span>{b.home_network_cluster_id}</span>
                    </div>
                  )}
                </div>
              ))}
            </>
          )}
        </Col>
        <Col md={6}>
          <Message content={m.routingPolicyFromThisNetwork} component="h3" />
          <RoutingPolicy.Sheet policy={forwarder} />
        </Col>
      </Row>
      <Row>
        <Col md={12} lg={8} className={style.setRoutingPolicyContainer}>
          <Message content={m.routingPolicyToThisNetwork} component="h3" />
          <RoutingPolicyForm
            onSubmit={handleRoutingPolicySubmit}
            defaultPolicy={defaultRoutingPolicy}
            submitMessage={m.saveRoutingPolicy}
            initialValues={initialValues}
            error={formError}
            networkLevel
          />
        </Col>
      </Row>
    </Container>
  )
}

const NetworkRoutingPolicyView = () => {
  const { netId, tenantId } = useParams()
  const combinedId = tenantId ? `${netId}/${tenantId}` : netId
  const registered = useSelector(selectRegistered)

  return (
    <Require
      condition={registered}
      otherwise={{ redirect: '/admin-panel/packet-broker/routing-configuration' }}
    >
      <RequireRequest
        requestAction={[
          getPacketBrokerNetwork(combinedId, { fetchPolicies: true }),
          getHomeNetworkDefaultRoutingPolicy(),
        ]}
      >
        <NetworkRoutingPolicyViewInner />
      </RequireRequest>
    </Require>
  )
}

export default NetworkRoutingPolicyView
