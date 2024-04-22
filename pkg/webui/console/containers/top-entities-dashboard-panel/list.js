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
import { FixedSizeList as List } from 'react-window'
import InfiniteLoader from 'react-window-infinite-loader'
import AutoSizer from 'react-virtualized-auto-sizer'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import Spinner from '@ttn-lw/components/spinner'
import { Table } from '@ttn-lw/components/table'
import { IconPlus } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

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
  const itemsTotalCount = useSelector(state => itemsCountSelector(state, entity))
  const [items, setItems] = useState([])
  const showScrollIndicator = itemsTotalCount > 5
  const hasNextPage = items.length < itemsTotalCount
  const EntitiesItemComponent = EntitiesItemProp ?? EntitiesItem
  const itemCount = items.length === itemsTotalCount ? itemsTotalCount : items.length + 1
  const [fetching, setFetching] = useState(false)

  const loadNextPage = useCallback(
    async startIndex => {
      if (fetching) return
      setFetching(true)
      const nextBookmarks = allBookmarks.slice(startIndex, startIndex + 10)
      setItems([...items, ...nextBookmarks])

      setFetching(false)
    },
    [fetching, allBookmarks, items],
  )

  const isItemLoaded = useCallback(
    index => (items.length > 0 ? !hasNextPage || index < items.length : false),
    [hasNextPage, items],
  )

  const Item = ({ index, style }) =>
    isItemLoaded(index) && items[index] ? (
      <div style={style}>
        <EntitiesItemComponent
          headers={headers}
          bookmark={items[index]}
          last={index === itemsTotalCount - 1 && showScrollIndicator}
        />
      </div>
    ) : (
      <div style={style}>
        <Spinner faded micro center />
      </div>
    )

  Item.propTypes = {
    index: PropTypes.number.isRequired,
    style: PropTypes.shape({}).isRequired,
  }

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

  const minWidth = `${headers.length * 10 + 5}rem`

  return items.length === 0 && itemsTotalCount === 0 ? (
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
    <Table minWidth={minWidth}>
      <Table.Head>{columns}</Table.Head>
      <Table.Body className={styles.entityBody} emptyMessage={m.empty}>
        <AutoSizer>
          {({ width }) => (
            <InfiniteLoader
              loadMoreItems={loadNextPage}
              isItemLoaded={isItemLoaded}
              itemCount={itemsTotalCount}
              threshold={5}
            >
              {({ onItemsRendered, ref }) => (
                <>
                  <List
                    height={56 * 5}
                    width={width}
                    itemSize={56}
                    ref={ref}
                    itemCount={itemCount}
                    onItemsRendered={onItemsRendered}
                    className={styles.entityList}
                  >
                    {Item}
                  </List>
                  {showScrollIndicator && <div className={styles.entityListGradient} />}
                </>
              )}
            </InfiniteLoader>
          )}
        </AutoSizer>
      </Table.Body>
    </Table>
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
