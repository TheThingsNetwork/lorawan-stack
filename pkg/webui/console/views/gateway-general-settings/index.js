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

import React, { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'
import { isEqual } from 'lodash'

import toast from '@ttn-lw/components/toast'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Collapse from '@ttn-lw/components/collapse'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import Require from '@console/lib/components/require'

import diff from '@ttn-lw/lib/diff'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'

import {
  checkFromState,
  mayEditBasicGatewayInformation,
  mayDeleteGateway,
  mayEditGatewaySecrets,
  mayViewOrEditGatewayApiKeys,
  mayViewOrEditGatewayCollaborators,
} from '@console/lib/feature-checks'

import { updateGateway } from '@console/store/actions/gateways'
import { getApiKeysList } from '@console/store/actions/api-keys'
import { getIsConfiguration } from '@console/store/actions/identity-server'

import {
  selectSelectedGateway,
  selectSelectedGatewayClaimable,
  selectSelectedGatewayId,
} from '@console/store/selectors/gateways'

import LorawanSettingsForm from './lorawan-settings-form'
import BasicSettingsForm from './basic-settings-form'
import m from './messages'

const GatewayGeneralSettingsInner = () => {
  const dispatch = useDispatch()
  const { gtwId } = useParams()
  const gateway = useSelector(selectSelectedGateway)
  const mayDeleteGtw = useSelector(state => checkFromState(mayDeleteGateway, state))
  const mayEditSecrets = useSelector(state => checkFromState(mayEditGatewaySecrets, state))
  const supportsClaiming = useSelector(selectSelectedGatewayClaimable)

  const handleSubmit = useCallback(
    async values => {
      const formValues = { ...values }
      const { attributes, frequency_plan_ids } = formValues
      if (isEqual(gateway.attributes || {}, attributes)) {
        delete formValues.attributes
      }
      if (isEqual(gateway.frequency_plan_ids || {}, frequency_plan_ids)) {
        delete formValues.frequency_plan_ids
      }

      const changed = diff(gateway, formValues, {
        patchArraysItems: false,
        patchInFull: ['attributes', 'frequency_plan_ids'],
      })

      try {
        await dispatch(updateGateway(gtwId, changed))
        toast({
          title: gtwId,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        toast({
          title: gtwId,
          message: m.updateFailure,
          type: toast.types.ERROR,
        })
      }
    },
    [gateway, dispatch, gtwId],
  )

  return (
    <div className="container container--xl grid">
      <PageTitle title={sharedMessages.generalSettings} hideHeading />
      <div className="item-12">
        <Collapse
          title={m.basicTitle}
          description={m.basicDescription}
          disabled={false}
          initialCollapsed={false}
        >
          <BasicSettingsForm
            gtwId={gtwId}
            gateway={gateway}
            onSubmit={handleSubmit}
            mayDeleteGateway={mayDeleteGtw}
            mayEditSecrets={mayEditSecrets}
            supportsClaiming={supportsClaiming}
          />
        </Collapse>
        <Collapse
          title={sharedMessages.lorawanOptions}
          description={m.lorawanDescription}
          disabled={false}
          initialCollapsed
        >
          <LorawanSettingsForm gateway={gateway} onSubmit={handleSubmit} />
        </Collapse>
      </div>
    </div>
  )
}

const GatewaySettings = () => {
  const gtwId = useSelector(selectSelectedGatewayId)
  const mayDeleteGtw = useSelector(state => checkFromState(mayDeleteGateway, state))
  const mayViewApiKeys = useSelector(state => checkFromState(mayViewOrEditGatewayApiKeys, state))
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditGatewayCollaborators, state),
  )

  const loadData = useCallback(
    async dispatch => {
      if (mayDeleteGtw) {
        if (mayViewApiKeys) {
          await dispatch(attachPromise(getApiKeysList('gateway', gtwId)))
        }
        if (mayViewCollaborators) {
          await dispatch(attachPromise(getCollaboratorsList('gateway', gtwId)))
        }
      }
      dispatch(attachPromise(getIsConfiguration()))
    },
    [mayDeleteGtw, mayViewApiKeys, mayViewCollaborators, gtwId],
  )

  useBreadcrumbs(
    'gtws.single.general-settings',
    <Breadcrumb
      path={`/gateways/${gtwId}/general-settings`}
      content={sharedMessages.generalSettings}
    />,
  )

  return (
    <Require
      featureCheck={mayEditBasicGatewayInformation}
      otherwise={{ redirect: `/gateways/${gtwId}` }}
    >
      <RequireRequest requestAction={loadData}>
        <GatewayGeneralSettingsInner />
      </RequireRequest>
    </Require>
  )
}

export default GatewaySettings
