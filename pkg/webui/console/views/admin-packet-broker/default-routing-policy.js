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
import { Col } from 'react-grid-system'
import { useSelector, useDispatch } from 'react-redux'

import toast from '@ttn-lw/components/toast'

import Message from '@ttn-lw/lib/components/message'

import RoutingPolicyForm from '@console/components/routing-policy-form'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { isValidPolicy } from '@console/lib/packet-broker/utils'

import {
  setHomeNetworkDefaultRoutingPolicy,
  deleteHomeNetworkDefaultRoutingPolicy,
} from '@console/store/actions/packet-broker'

import { selectHomeNetworkDefaultRoutingPolicy } from '@console/store/selectors/packet-broker'

import m from './messages'

import style from './admin-packet-broker.styl'

const DefaultRoutingPolicyView = () => {
  const dispatch = useDispatch()
  const defaultRoutingPolicy = useSelector(selectHomeNetworkDefaultRoutingPolicy)
  const initialValues = { _use_default_policy: isValidPolicy(defaultRoutingPolicy) }
  initialValues.policy = initialValues._use_default_policy
    ? defaultRoutingPolicy
    : { uplink: {}, downlink: {} }
  const [formError, setFormError] = useState(undefined)
  const handleDefaultRoutingPolicySubmit = useCallback(
    async ({ _use_default_policy, policy }) => {
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

  return (
    <Col md={12}>
      <Message
        content={m.routingPolicyInformation}
        component="p"
        className={style.routingPolicyInformation}
      />
      <RoutingPolicyForm
        onSubmit={handleDefaultRoutingPolicySubmit}
        initialValues={initialValues}
        error={formError}
      />
    </Col>
  )
}

export default DefaultRoutingPolicyView
