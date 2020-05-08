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

import PropTypes from '@ttn-lw/lib/prop-types'
import { getApplicationId, getDeviceId } from '@ttn-lw/lib/selectors/id'

import CRUDEvent from '../../shared/components/crud-event'

const ApplicationCRUDEvent = React.memo(props => {
  const { event, widget } = props
  const ids = event.identifiers[0]

  let id = getDeviceId(ids)
  if (!id) {
    id = getApplicationId(ids)
  }

  return <CRUDEvent event={event} entityId={id} widget={widget} />
})

ApplicationCRUDEvent.propTypes = {
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool,
}

ApplicationCRUDEvent.defaultProps = {
  widget: false,
}

export default ApplicationCRUDEvent
