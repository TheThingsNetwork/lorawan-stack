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

import applicationIcon from '@assets/misc/application.svg'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import SideNavigation from '@ttn-lw/components/navigation/side'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

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

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'

import {
  mayViewApplicationInfo,
  mayViewApplicationEvents,
  maySetApplicationPayloadFormatters,
  mayViewApplicationDevices,
  mayCreateOrEditApplicationIntegrations,
  mayEditBasicApplicationInfo,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayViewOrEditApplicationPackages,
  mayAddPubSubIntegrations,
  mayViewApplications,
} from '@console/lib/feature-checks'

import {
  getApplication,
  getApplicationsRightsList,
  stopApplicationEventsStream,
} from '@console/store/actions/applications'
import { getAsConfiguration } from '@console/store/actions/application-server'

import {
  selectApplicationRights,
  selectSelectedApplication,
} from '@console/store/selectors/applications'
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
  const rights = useSelector(selectApplicationRights)
  const siteName = selectApplicationSiteName()
  const natsDisabled = useSelector(selectNatsProviderDisabled)
  const mqttDisabled = useSelector(selectMqttProviderDisabled)

  const dispatch = useDispatch()
  const stopStream = React.useCallback(id => dispatch(stopApplicationEventsStream(id)), [dispatch])

  useEffect(() => () => stopStream(appId), [appId, stopStream])
  useBreadcrumbs('apps.single', [
    {
      path: `/applications/${appId}`,
      content: name,
    },
  ])

  return (
    <>
      <Breadcrumbs />
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <SideNavigation
        header={{
          icon: applicationIcon,
          iconAlt: sharedMessages.application,
          title: name,
          to: '',
        }}
      >
        {mayViewApplicationInfo.check(rights) && (
          <SideNavigation.Item title={sharedMessages.overview} path="" icon="overview" exact />
        )}
        {mayViewApplicationDevices.check(rights) && (
          <SideNavigation.Item title={sharedMessages.devices} path="devices" icon="devices" />
        )}
        {mayViewApplicationEvents.check(rights) && (
          <SideNavigation.Item title={sharedMessages.liveData} path="data" icon="data" />
        )}
        {maySetApplicationPayloadFormatters.check(rights) && (
          <SideNavigation.Item title={sharedMessages.payloadFormatters} icon="code">
            <SideNavigation.Item
              title={sharedMessages.uplink}
              path="payload-formatters/uplink"
              icon="uplink"
            />
            <SideNavigation.Item
              title={sharedMessages.downlink}
              path="payload-formatters/downlink"
              icon="downlink"
            />
          </SideNavigation.Item>
        )}
        {mayCreateOrEditApplicationIntegrations.check(rights) && (
          <SideNavigation.Item title={sharedMessages.integrations} icon="integration">
            <SideNavigation.Item
              title={sharedMessages.mqtt}
              path="integrations/mqtt"
              icon="extension"
            />
            <SideNavigation.Item
              title={sharedMessages.webhooks}
              path="integrations/webhooks"
              icon="extension"
            />
            {mayAddPubSubIntegrations.check(natsDisabled, mqttDisabled) && (
              <SideNavigation.Item
                title={sharedMessages.pubsubs}
                path="integrations/pubsubs"
                icon="extension"
              />
            )}
            {mayViewOrEditApplicationPackages.check(rights) && (
              <SideNavigation.Item
                title={sharedMessages.loraCloud}
                path="integrations/lora-cloud"
                icon="extension"
              />
            )}
          </SideNavigation.Item>
        )}
        {mayViewOrEditApplicationCollaborators.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            path="collaborators"
            icon="organization"
          />
        )}
        {mayViewOrEditApplicationApiKeys.check(rights) && (
          <SideNavigation.Item title={sharedMessages.apiKeys} path="api-keys" icon="api_keys" />
        )}
        {mayEditBasicApplicationInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.generalSettings}
            path="general-settings"
            icon="general_settings"
          />
        )}
      </SideNavigation>
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
