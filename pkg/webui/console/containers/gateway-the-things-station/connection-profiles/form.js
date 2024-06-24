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

import React, { useCallback, useState } from 'react'
import { useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import SubmitButton from '@ttn-lw/components/submit-button'
import SubmitBar from '@ttn-lw/components/submit-bar'

import {
  CONNECTION_TYPES,
  getFormTypeMessage,
  getInitialProfile,
} from '@console/containers/gateway-the-things-station/utils'
import validationSchema from '@console/containers/gateway-the-things-station/connection-profiles/validation-schema'
import GatewayConnectionProfilesFormFields from '@console/containers/gateway-the-things-station/connection-profiles/connection-profiles-form-fields'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const GatewayConnectionProfilesForm = () => {
  const [error, setError] = useState(undefined)
  const { gtwId, type, profileId } = useParams()

  const isEdit = Boolean(profileId)

  const baseUrl = `/gateways/${gtwId}/the-things-station/connection-profiles/${type}`

  useBreadcrumbs(
    'gtws.single.the-things-station.connection-profiles.form',
    <Breadcrumb
      path={isEdit ? `${baseUrl}/edit/${profileId}` : `${baseUrl}/add`}
      content={getFormTypeMessage(type, profileId)}
    />,
  )

  const handleSubmit = useCallback(values => {
    try {
      console.log(values)
    } catch (e) {
      setError(e)
    }
  }, [])

  const initialValues = getInitialProfile(type)

  return (
    <>
      <PageTitle title={getFormTypeMessage(type, profileId)} />
      <Form
        error={error}
        onSubmit={handleSubmit}
        initialValues={initialValues}
        validationSchema={validationSchema}
      >
        <>
          <GatewayConnectionProfilesFormFields isEdit={isEdit} />

          <SubmitBar>
            <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
          </SubmitBar>
        </>
      </Form>
    </>
  )
}

export default GatewayConnectionProfilesForm
