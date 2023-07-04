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

import React from 'react'
import { Navigate } from 'react-router-dom'
import { useSelector } from 'react-redux'

import ApplicationWebhookAddForm from '@console/views/application-integrations-webhook-add-form'

import { selectWebhookTemplates } from '@console/store/selectors/webhook-templates'

const ApplicationWebhookAdd = () => {
  const hasTemplates = useSelector(state => selectWebhookTemplates(state).length > 0)

  // Forward to the template chooser, when templates have been configured.
  if (hasTemplates) {
    return <Navigate to="template" />
  }

  return <ApplicationWebhookAddForm />
}

export default ApplicationWebhookAdd
