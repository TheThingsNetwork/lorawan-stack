// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import CRUDEvent from '../shared/components/crud-event'
import DefaultEvent from '../shared/components/default-event'
import ErrorEvent from '../shared/components/error-event'
import { isCRUDEvent, isErrorEvent } from '../shared/utils/types'

const renderOrganizationEvent = (event, widget = false) => {
  if (isErrorEvent(event)) {
    return <ErrorEvent event={event} widget={widget} />
  }

  if (isCRUDEvent(event)) {
    return <CRUDEvent event={event} widget={widget} />
  }

  return <DefaultEvent event={event} widget={widget} />
}

export default renderOrganizationEvent
