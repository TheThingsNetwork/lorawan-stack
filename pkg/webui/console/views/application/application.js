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

import React, { useEffect } from 'react'
import { Switch, Route } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import SideNavigation from '@ttn-lw/components/navigation/side'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

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

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

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
} from '@console/lib/feature-checks'

const Application = props => {
  const {
    appId,
    match: { url: matchedUrl, path },
    application,
    rights,
    stopStream,
    siteName,
  } = props

  const name = application.name || appId

  useBreadcrumbs('apps.single', <Breadcrumb path={`/applications/${appId}`} content={name} />)

  useEffect(() => () => stopStream(appId), [appId, stopStream])

  return (
    <React.Fragment>
      <Breadcrumbs />
      <IntlHelmet titleTemplate={`%s - ${name} - ${siteName}`} />
      <SideNavigation header={{ icon: 'application', title: name, to: matchedUrl }}>
        {mayViewApplicationInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.overview}
            path={matchedUrl}
            icon="overview"
            exact
          />
        )}
        {mayViewApplicationDevices.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.devices}
            path={`${matchedUrl}/devices`}
            icon="devices"
          />
        )}
        {mayViewApplicationEvents.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.liveData}
            path={`${matchedUrl}/data`}
            icon="data"
          />
        )}
        {maySetApplicationPayloadFormatters.check(rights) && (
          <SideNavigation.Item title={sharedMessages.payloadFormatters} icon="code">
            <SideNavigation.Item
              title={sharedMessages.uplink}
              path={`${matchedUrl}/payload-formatters/uplink`}
              icon="uplink"
            />
            <SideNavigation.Item
              title={sharedMessages.downlink}
              path={`${matchedUrl}/payload-formatters/downlink`}
              icon="downlink"
            />
          </SideNavigation.Item>
        )}
        {mayCreateOrEditApplicationIntegrations.check(rights) && (
          <SideNavigation.Item title={sharedMessages.integrations} icon="integration">
            <SideNavigation.Item
              title={sharedMessages.mqtt}
              path={`${matchedUrl}/integrations/mqtt`}
              icon="extension"
            />
            <SideNavigation.Item
              title={sharedMessages.webhooks}
              path={`${matchedUrl}/integrations/webhooks`}
              icon="extension"
            />
            <SideNavigation.Item
              title={sharedMessages.pubsubs}
              path={`${matchedUrl}/integrations/pubsubs`}
              icon="extension"
            />
            {mayViewOrEditApplicationPackages.check(rights) && (
              <SideNavigation.Item
                title={sharedMessages.loraCloud}
                path={`${matchedUrl}/integrations/lora-cloud`}
                icon="extension"
              />
            )}
          </SideNavigation.Item>
        )}
        {mayViewOrEditApplicationCollaborators.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.collaborators}
            path={`${matchedUrl}/collaborators`}
            icon="organization"
          />
        )}
        {mayViewOrEditApplicationApiKeys.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.apiKeys}
            path={`${matchedUrl}/api-keys`}
            icon="api_keys"
          />
        )}
        {mayEditBasicApplicationInfo.check(rights) && (
          <SideNavigation.Item
            title={sharedMessages.generalSettings}
            path={`${matchedUrl}/general-settings`}
            icon="general_settings"
          />
        )}
      </SideNavigation>
      <Switch>
        <Route exact path={`${path}`} component={ApplicationOverview} />
        <Route path={`${path}/general-settings`} component={ApplicationGeneralSettings} />
        <Route path={`${path}/api-keys`} component={ApplicationApiKeys} />
        <Route path={`${path}/devices`} component={Devices} />
        <Route path={`${path}/collaborators`} component={ApplicationCollaborators} />
        <Route path={`${path}/data`} component={ApplicationData} />
        <Route path={`${path}/payload-formatters`} component={ApplicationPayloadFormatters} />
        <Route path={`${path}/integrations/mqtt`} component={ApplicationIntegrationsMqtt} />
        <Route path={`${path}/integrations/webhooks`} component={ApplicationIntegrationsWebhooks} />
        <Route path={`${path}/integrations/pubsubs`} component={ApplicationIntegrationsPubsubs} />
        <Route
          path={`${path}/integrations/lora-cloud`}
          component={ApplicationIntegrationsLoRaCloud}
        />
        <NotFoundRoute />
      </Switch>
    </React.Fragment>
  )
}

Application.propTypes = {
  appId: PropTypes.string.isRequired,
  application: PropTypes.application.isRequired,
  match: PropTypes.match.isRequired,
  rights: PropTypes.rights.isRequired,
  siteName: PropTypes.string.isRequired,
  stopStream: PropTypes.func.isRequired,
}

export default Application
