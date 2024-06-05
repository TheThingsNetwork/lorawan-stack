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

import React, { useRef } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import { Table } from '@ttn-lw/components/table'
import { IconPlus } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import ScrollFader from '@ttn-lw/components/scroll-fader'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import EntitiesItem from './item'

import styles from './top-entities-panel.styl'

const m = defineMessages({
  empty: 'No entities yet',
})

const EntitiesList = ({
  itemsCountSelector,
  allBookmarks,
  headers,
  emptyMessage,
  emptyDescription,
  emptyAction,
  emptyPath,
  EntitiesItemComponent: EntitiesItemProp,
  entity,
}) => {
  const listRef = useRef()
  const itemsTotalCount = useSelector(state => itemsCountSelector(state, entity))
  const EntitiesItemComponent = EntitiesItemProp ?? EntitiesItem

  const rows = allBookmarks
    .slice(0, 10)
    .map((bookmark, index) => (
      <EntitiesItemComponent key={index} headers={headers} bookmark={bookmark} />
    ))

  const columns = (
    <Table.Row head>
      {headers.map((header, key) => (
        <Table.HeadCell
          key={key}
          align={header.align}
          content={header.displayName}
          name={header.name}
          width={header.width}
          className={header.className}
        />
      ))}
    </Table.Row>
  )

  return allBookmarks.length === 0 && itemsTotalCount === 0 ? (
    <div className="d-flex direction-column flex-grow j-center gap-cs-l">
      <div>
        <Message content={emptyMessage} className="d-block text-center fs-l fw-bold" />
        <Message content={emptyDescription} className="d-block text-center c-text-neutral-light" />
      </div>
      {emptyAction && (
        <div className="text-center">
          <Button.Link to={emptyPath} primary message={emptyAction} icon={IconPlus} />
        </div>
      )}
    </div>
  ) : (
    <ScrollFader
      className={styles.scrollFader}
      ref={listRef}
      faderHeight="4rem"
      topFaderOffset="3rem"
      light
    >
      <Table>
        <Table.Head className={styles.topEntitiesPanelOuterTableHeader}>{columns}</Table.Head>
        <Table.Body emptyMessage={m.empty}>{rows}</Table.Body>
      </Table>
    </ScrollFader>
  )
}

EntitiesList.propTypes = {
  EntitiesItemComponent: PropTypes.func,
  allBookmarks: PropTypes.arrayOf(PropTypes.object).isRequired,
  emptyAction: PropTypes.message,
  emptyDescription: PropTypes.message,
  emptyMessage: PropTypes.message,
  emptyPath: PropTypes.string,
  entity: PropTypes.string,
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      align: PropTypes.string,
      displayName: PropTypes.message,
      name: PropTypes.string,
      width: PropTypes.string,
      className: PropTypes.string,
    }),
  ).isRequired,
  itemsCountSelector: PropTypes.func.isRequired,
}

EntitiesList.defaultProps = {
  emptyDescription: undefined,
  emptyMessage: undefined,
  emptyAction: undefined,
  emptyPath: undefined,
  entity: undefined,
  EntitiesItemComponent: undefined,
}

export default EntitiesList
