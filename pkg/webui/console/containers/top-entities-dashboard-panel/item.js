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

import React, { useEffect } from 'react'
import { useDispatch } from 'react-redux'

import { APPLICATION, END_DEVICE } from '@console/constants/entities'
import { EVENT_END_DEVICE_HEARTBEAT_FILTERS_STRING } from '@console/constants/event-filters'

import { Table } from '@ttn-lw/components/table'

import PropTypes from '@ttn-lw/lib/prop-types'

import { startDeviceEventsStream, stopDeviceEventsStream } from '@console/store/actions/devices'
import {
  startApplicationEventsStream,
  stopApplicationEventsStream,
} from '@console/store/actions/applications'

const EntitiesItem = ({ entity, headers }) => {
  const { id, path, type } = entity
  const createdAt = entity.entity ? entity.entity.created_at : null
  const dispatch = useDispatch()

  // Start an app/device event stream if the entity is an end device
  // so we can update the last seen status in real-time.
  useEffect(() => {
    if (createdAt === null) {
      // Only start the stream if the entity is already fetched
      // so we avoid starting streams for entities that don't exist
      // or we don't have permissions for.
      return
    }
    if (type === END_DEVICE) {
      dispatch(startDeviceEventsStream(id, { filter: [EVENT_END_DEVICE_HEARTBEAT_FILTERS_STRING] }))
    } else if (type === APPLICATION) {
      dispatch(
        startApplicationEventsStream(id, { filter: [EVENT_END_DEVICE_HEARTBEAT_FILTERS_STRING] }),
      )
    }
    return () => {
      if (type === END_DEVICE) {
        dispatch(stopDeviceEventsStream(id))
      } else if (type === APPLICATION) {
        dispatch(stopApplicationEventsStream(id))
      }
    }
  }, [createdAt, dispatch, id, type])

  return (
    <Table.Row id={id} clickable linkTo={path} body panelStyle>
      {headers.map((header, index) => {
        const value = header.getValue ? header.getValue(entity) : entity[header.name]
        return (
          <Table.DataCell
            key={`${id}-${index}`}
            align={header.align}
            className={header.className}
            panelStyle
          >
            {headers[index].render(value, id)}
          </Table.DataCell>
        )
      })}
    </Table.Row>
  )
}

EntitiesItem.propTypes = {
  entity: PropTypes.unifiedEntity.isRequired,
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      displayName: PropTypes.shape({}),
      render: PropTypes.func,
      getValue: PropTypes.func,
      align: PropTypes.string,
    }),
  ).isRequired,
}

export default EntitiesItem
