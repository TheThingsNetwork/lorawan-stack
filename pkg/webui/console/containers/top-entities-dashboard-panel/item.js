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

import React from 'react'
import classNames from 'classnames'

import { Table } from '@ttn-lw/components/table'

import useBookmark from '@ttn-lw/lib/hooks/use-bookmark'
import PropTypes from '@ttn-lw/lib/prop-types'

import styles from './top-entities-panel.styl'

const EntitiesItem = ({ bookmark, headers, last }) => {
  const { title, ids, path, icon } = useBookmark(bookmark)

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
          headers[index].name === 'name' ? title : headers[index].name === 'type' ? icon : ''
        const entityID = headers[index].name === 'name' ? ids.id : undefined
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
  bookmark: PropTypes.shape({}).isRequired,
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string.isRequired,
      displayName: PropTypes.string,
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
