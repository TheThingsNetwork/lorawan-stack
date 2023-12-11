// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import RequireRequest from '@ttn-lw/lib/components/require-request'

import GatewayVisibilityForm from '@console/components/gateway-visibility-form'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  getHomeNetworkDefaultGatewayVisibility,
  setHomeNetworkDefaultGatewayVisibility,
} from '@console/store/actions/packet-broker'

import { selectHomeNetworkDefaultGatewayVisibility } from '@console/store/selectors/packet-broker'

import m from './messages'

import style from './admin-packet-broker.styl'

const DefaultGatewayVisibilityView = () => {
  const dispatch = useDispatch()
  const defaultGatewayVisibility = useSelector(selectHomeNetworkDefaultGatewayVisibility)
  const initialValues = {
    visibility: defaultGatewayVisibility.visibility || {},
  }
  const [formError, setFormError] = useState(undefined)
  const handleDefaultGatewayVisibilitySubmit = useCallback(
    async ({ visibility }) => {
      try {
        await dispatch(attachPromise(setHomeNetworkDefaultGatewayVisibility(visibility)))
        toast({
          message: m.defaultGatewayVisibilitySet,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setFormError(error)
      }
    },
    [dispatch, setFormError],
  )

  return (
    <RequireRequest requestAction={getHomeNetworkDefaultGatewayVisibility()}>
      <Col md={12}>
        <Message
          content={m.gatewayVisibilityInformation}
          component="p"
          className={style.routingPolicyInformation}
        />
        <GatewayVisibilityForm
          onSubmit={handleDefaultGatewayVisibilitySubmit}
          initialValues={initialValues}
          error={formError}
        />
      </Col>
    </RequireRequest>
  )
}

export default DefaultGatewayVisibilityView
