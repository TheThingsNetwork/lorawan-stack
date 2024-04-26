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
import classNames from 'classnames'
import { useDispatch, useSelector } from 'react-redux'

import { Table } from '@ttn-lw/components/table'

import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'
import PropTypes from '@ttn-lw/lib/prop-types'

import { startGatewayStatistics, stopGatewayStatistics } from '@console/store/actions/gateways'

import { selectDeviceLastSeen } from '@console/store/selectors/devices'
import { selectApplicationDerivedLastSeen } from '@console/store/selectors/applications'
import { selectGatewayStatistics } from '@console/store/selectors/gateways'
import { selectGatewayLastSeen } from '@console/store/selectors/gateway-status'

import styles from './top-entities-panel.styl'

const EntitiesItem = ({ bookmark, headers, last }) => {
  const dispatch = useDispatch()
  const { title, ids, path, icon } = useBookmark(bookmark)
  const entityIds = bookmark.entity_ids
  const entity = Object.keys(entityIds)[0].replace('_ids', '')
  const deviceLastSeen = useSelector(state => selectDeviceLastSeen(state, ids.appId, ids.id))
  const appLastSeen = useSelector(state => selectApplicationDerivedLastSeen(state, ids.id))
  const statistics = useSelector(selectGatewayStatistics)
  const gatewayLastSeen = useSelector(selectGatewayLastSeen)

  useEffect(() => {
    dispatch(startGatewayStatistics(ids.id))
    return () => {
      dispatch(stopGatewayStatistics())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const isDisconnected = Boolean(statistics) && Boolean(statistics.disconnected_at)

  const status = {
    gatewayLastSeen,
    isDisconnected,
    disconnectedAt: statistics?.disconnected_at,
  }

  const lastSeen =
    entity === 'device' ? deviceLastSeen : entity === 'gateway' ? status : appLastSeen

  return (
    <Table.Row
      id={ids.id}
      clickable
      linkTo={path}
      body
      className={classNames(styles.entityRow, { [styles.lastRow]: last })}
    >
      {headers.map((header, index) => {
        const value =
          headers[index].name === 'name'
            ? title
            : headers[index].name === 'type'
              ? icon
              : headers[index].name === 'lastSeen'
                ? lastSeen
                : ''
        const entityID = ids.id
        return (
          <Table.DataCell
            key={index}
            align={header.align}
            className={classNames(styles.entityCell, {
              [styles.entityCellExtended]: index === 1 && headers[index].name === 'name',
              [styles.entityCellSmall]: headers[index].name === 'type',
            })}
          >
            {headers[index].render(value, entityID)}
          </Table.DataCell>
        )
      })}
    </Table.Row>
  )
}

EntitiesItem.propTypes = {
  bookmark: PropTypes.shape({
    entity_ids: PropTypes.shape({}).isRequired,
  }).isRequired,
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      displayName: PropTypes.shape({}),
      render: PropTypes.func,
      getValue: PropTypes.func,
      align: PropTypes.string,
    }),
  ).isRequired,
  last: PropTypes.bool,
}

EntitiesItem.defaultProps = {
  last: false,
}

export default EntitiesItem
