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

import React, { useEffect } from 'react'
import { Routes, Route, useParams } from 'react-router-dom'
import { useDispatch, useSelector } from 'react-redux'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import RequireRequest from '@ttn-lw/lib/components/require-request'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import Require from '@console/lib/components/require'

import ApplicationOverview from '@console/views/application-overview'
import ApplicationGeneralSettings from '@console/views/application-general-settings'
import ApplicationApiKeys from '@console/views/application-api-keys'
import ApplicationCollaborators from '@console/views/application-collaborators'
import ApplicationData from '@console/views/application-data'
import ApplicationPayloadFormatters from '@console/views/application-payload-formatters'
import ApplicationIntegrationsWebhooks from '@console/views/application-integrations-webhooks'
import ApplicationIntegrationsPubsubs from '@console/views/application-integrations-pubsubs'
import ApplicationIntegrationsMqtt from '@console/views/application-integrations-mqtt'
import ApplicationIntegrationsLoRaCloud from '@console/views/application-integrations-lora-cloud'
import Devices from '@console/views/devices'

import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

import { mayAddPubSubIntegrations, mayViewApplications } from '@console/lib/feature-checks'

import {
  getApplication,
  getApplicationsRightsList,
  stopApplicationEventsStream,
} from '@console/store/actions/applications'
import { getAsConfiguration } from '@console/store/actions/application-server'

import { selectSelectedApplication } from '@console/store/selectors/applications'
import {
  selectMqttProviderDisabled,
  selectNatsProviderDisabled,
} from '@console/store/selectors/application-server'

const Application = () => {
  const { appId } = useParams()
  const actions = [
    getApplication(
      appId,
      'name,description,attributes,dev_eui_counter,network_server_address,application_server_address,join_server_address,administrative_contact,technical_contact',
    ),
    getApplicationsRightsList(appId),
    getAsConfiguration(),
  ]

  // Check whether application still exists after it has been possibly deleted.
  const application = useSelector(selectSelectedApplication)
  const hasApplication = Boolean(application)

  return (
    <Require featureCheck={mayViewApplications} otherwise={{ redirect: '/' }}>
      <RequireRequest requestAction={actions}>
        {hasApplication && <ApplicationInner />}
      </RequireRequest>
    </Require>
  )
}

const ApplicationInner = () => {
  const { appId } = useParams()
  const application = useSelector(selectSelectedApplication)
  const name = application.name || appId
  const siteName = selectApplicationSiteName()
  const natsDisabled = useSelector(selectNatsProviderDisabled)
  const mqttDisabled = useSelector(selectMqttProviderDisabled)

  const dispatch = useDispatch()
  const stopStream = React.useCallback(id => dispatch(stopApplicationEventsStream(id)), [dispatch])

  useEffect(() => () => stopStream(appId), [appId, stopStream])
  useBreadcrumbs('apps.single', <Breadcrumb path={`/applications/${appId}`} content={name} />)

  return (
    <>
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <Routes>
        <Route index Component={ApplicationOverview} />
        <Route path="general-settings" Component={ApplicationGeneralSettings} />
        <Route path="api-keys/*" Component={ApplicationApiKeys} />
        <Route path="devices/*" Component={Devices} />
        <Route path="collaborators/*" Component={ApplicationCollaborators} />
        <Route path="data" Component={ApplicationData} />
        <Route path="payload-formatters/*" Component={ApplicationPayloadFormatters} />
        <Route path="integrations/mqtt" Component={ApplicationIntegrationsMqtt} />
        <Route path="integrations/webhooks/*" Component={ApplicationIntegrationsWebhooks} />
        {mayAddPubSubIntegrations.check(natsDisabled, mqttDisabled) && (
          <Route path="integrations/pubsubs/*" Component={ApplicationIntegrationsPubsubs} />
        )}
        <Route path="integrations/lora-cloud" Component={ApplicationIntegrationsLoRaCloud} />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </>
  )
}

export default Application
