// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useMemo, useState } from 'react'
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'
import { createSelector } from 'reselect'

import PageTitle from '@ttn-lw/components/page-title'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Tabs from '@ttn-lw/components/tabs'
import Link from '@ttn-lw/components/link'
import Select from '@ttn-lw/components/select'
import Form from '@ttn-lw/components/form'
import Button from '@ttn-lw/components/button'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import tooltipIds from '@ttn-lw/lib/constants/tooltip-ids'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'

import { selectOAuthClients, selectOAuthClientsTotalCount } from '@account/store/selectors/clients'

const m = defineMessages({
  theThingsStationConnectionProfiles: 'The Things Station connection profiles',
  wifiProfiles: 'WiFi profiles',
  ethernetProfiles: 'Ethernet profiles',
  information:
    'Connection profiles are setup to allow for multiple gateways to connect via the same settings. You can use this view to manage all your profiles or create new ones, after which you can assign them to your gateway.<br></br> <link>Learn more about gateway network connection profiles.</link>',
  showProfilesOf: 'Show profiles of',
  yourself: 'Yourself',
  addWifiProfile: 'Add WiFi profile',
  addEthernetProfile: 'Add Ethernet profile',
  profileId: 'Profile ID',
  accessPoint: 'Access point',
})

const profileOptions = [
  { value: '0', label: m.yourself },
  { value: '1', label: 'TTI' },
]

const GatewayConnectionProfiles = () => {
  const { gtwId } = useParams()
  const [activeTab, setActiveTab] = useState('wifi')

  useBreadcrumbs(
    'gtws.single.the-things-station.connection-profiles',
    <Breadcrumb
      path={`/gateways/${gtwId}/the-things-station/connection-profiles`}
      content={sharedMessages.connectionProfiles}
    />,
  )

  const tabs = [
    {
      title: m.wifiProfiles,
      name: 'wifi',
      exact: false,
    },
    {
      title: m.ethernetProfiles,
      name: 'ethernet',
    },
  ]

  const onAddProfile = useCallback(() => {}, [])

  const addProfileMessage = useMemo(() => {
    if (activeTab === 'wifi') {
      return m.addWifiProfile
    }
    if (activeTab === 'ethernet') {
      return m.addEthernetProfile
    }
    return null
  }, [activeTab])

  const headers = React.useMemo(
    () => [
      {
        name: 'ids.profile_id',
        displayName: m.profileId,
        width: 25,
        sortable: true,
      },
      {
        name: 'access_point',
        displayName: m.accessPoint,
        width: 25,
        sortable: true,
      },
      {
        name: 'created_at',
        displayName: sharedMessages.created,
        width: 50,
      },
      // {
      //   name: 'state',
      //   displayName: sharedMessages.state,
      //   width: 50,
      //   render: state => capitalizeMessage(formatMessage({ id: `enum:${state}` })),
      // },
    ],
    [],
  )

  /* Const baseDataSelector = createSelector(
    [selectOAuthClients, selectOAuthClientsTotalCount, mayAddSelector],
    (clients, totalCount, mayAdd) => ({
      clients,
      totalCount,
      mayAdd,
    }),
  )*/
  const baseDataSelector = useCallback(
    () => ({
      profiles: [],
      totalCount: 0,
      mayAdd: false,
    }),
    [],
  )

  return (
    <>
      <PageTitle title={m.theThingsStationConnectionProfiles} />
      <Tabs tabs={tabs} active={activeTab} onTabChange={setActiveTab} divider />
      <Message
        className="d-block mt-cs-l mb-cs-l"
        content={m.information}
        values={{
          link: txt => (
            <Link secondary to="#">
              {txt}
            </Link>
          ),
          br: () => <br />,
        }}
      />
      <div className="d-flex j-between al-end">
        <Form
          initialValues={{
            profiles: '0',
          }}
        >
          <Form.Field
            name="profiles"
            title={m.showProfilesOf}
            component={Select}
            options={profileOptions}
            tooltipId={tooltipIds.GATEWAY_SHOW_PROFILES}
          />
        </Form>
        <Button
          className="mb-cs-m"
          primary
          onClick={onAddProfile}
          message={addProfileMessage}
          icon="add"
        />
      </div>

      <FetchTable
        entity="clients"
        defaultOrder="-created_at"
        headers={headers}
        getItemsAction={() => {}}
        baseDataSelector={baseDataSelector}
      />
    </>
  )
}

export default GatewayConnectionProfiles
