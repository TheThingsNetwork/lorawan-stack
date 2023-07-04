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

import React from 'react'
import { Routes, Route, useParams } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import ErrorView from '@ttn-lw/lib/components/error-view'
import ValidateRouteParam from '@ttn-lw/lib/components/validate-route-param'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'

import ApplicationCollaboratorsList from '@console/views/application-collaborators-list'
import ApplicationCollaboratorEdit from '@console/views/application-collaborator-edit'
import SubViewError from '@console/views/sub-view-error'
import ApplicationCollaboratorAdd from '@console/views/application-collaborator-add'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { userPathId as userPathIdRegexp } from '@ttn-lw/lib/regexp'

const ApplicationCollaborators = () => {
  const { appId } = useParams()

  useBreadcrumbs(
    'apps.single.collaborators',
    <Breadcrumb
      path={`/applications/${appId}/collaborators`}
      content={sharedMessages.collaborators}
    />,
  )

  return (
    <ErrorView errorRender={SubViewError}>
      <Routes>
        <Route index Component={ApplicationCollaboratorsList} />
        <Route path="add" Component={ApplicationCollaboratorAdd} />
        <Route
          path=":collaboratorType/:collaboratorId"
          element={
            <ValidateRouteParam
              check={{
                collaboratorType: /^user$|^organization$/,
                collaboratorId: userPathIdRegexp,
              }}
              Component={ApplicationCollaboratorEdit}
            />
          }
        />
        <Route path="*" element={<GenericNotFound />} />
      </Routes>
    </ErrorView>
  )
}

export default ApplicationCollaborators
