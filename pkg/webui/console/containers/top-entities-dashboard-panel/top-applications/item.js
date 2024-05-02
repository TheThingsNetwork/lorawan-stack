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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import { Table } from '@ttn-lw/components/table'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import Message from '@ttn-lw/lib/components/message'

import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getApplicationDeviceCount } from '@console/store/actions/applications'

import {
  selectApplicationDerivedLastSeen,
  selectApplicationDeviceCount,
} from '@console/store/selectors/applications'

import styles from '../top-entities-panel.styl'

const m = defineMessages({
  errorMessage: 'Not available',
})

const TopApplicationsItem = ({ bookmark, headers, last }) => {
  const { title, ids, path } = useBookmark(bookmark)
  const deviceCount = useSelector(state => selectApplicationDeviceCount(state, ids.id))
  const lastSeen = useSelector(state => selectApplicationDerivedLastSeen(state, ids.id))

  const loadDeviceCount = useCallback(
    async dispatch => {
      if (!deviceCount) {
        dispatch(getApplicationDeviceCount(ids.id))
      }
    },
    [deviceCount, ids.id],
  )

  const errorRenderFunction = () => <Message content={m.errorMessage} />

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
            : headers[index].name === 'lastSeen'
              ? lastSeen
              : deviceCount
        const entityID = ids.id
        return (
          <RequireRequest
            key={index}
            requestAction={loadDeviceCount}
            errorRenderFunction={errorRenderFunction}
          >
            <Table.DataCell
              align={header.align}
              className={classNames(styles.entityCell, {
                [styles.entityCellDivided]:
                  headers[index].name === 'lastSeen' || headers[index].name === 'name',
              })}
            >
              {headers[index].render(value, entityID)}
            </Table.DataCell>
          </RequireRequest>
        )
      })}
    </Table.Row>
  )
}

TopApplicationsItem.propTypes = {
  bookmark: PropTypes.shape({}).isRequired,
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      displayName: PropTypes.shape({}).isRequired,
      render: PropTypes.func,
      getValue: PropTypes.func,
      align: PropTypes.string,
    }),
  ).isRequired,
  last: PropTypes.bool,
}

TopApplicationsItem.defaultProps = {
  last: false,
}

export default TopApplicationsItem
