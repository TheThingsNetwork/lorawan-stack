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

import classNames from 'classnames'
import React, { useRef, useEffect, useState, useCallback } from 'react'
import { defineMessages, useIntl } from 'react-intl'

import { APPLICATION, END_DEVICE, GATEWAY, ORGANIZATION } from '@console/constants/entities'

import Icon, {
  IconSearch,
  IconArrowUp,
  IconArrowDown,
  IconArrowBack,
  IconX,
  entityIcons,
} from '@ttn-lw/components/icon'
import ScrollFader from '@ttn-lw/components/scroll-fader'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import useDebounce from '@ttn-lw/lib/hooks/use-debounce'

import Spinner from '../spinner'
import Overlay from '../overlay'

import PanelItem from './item'
import LookingLuke from './looking-luke'

import style from './search-panel.styl'

const m = defineMessages({
  noResultsFound: 'No results found',
  noResultsSuggestion: 'Try searching for IDs names, attributes, EUIs or descriptions of:',
  devices: 'End devices of your{lineBreak}bookmarked applications',
  searchingEntities: 'Searching applications, gateways, organizations, bookmarks',
  instructions: 'Use {arrowKeys} to choose, {enter} to select',
  fetchingTopEntities: 'Fetching top entities…',
  noTopEntities: 'Seems like you haven’t interacted with any entities yet',
  noTopEntitiesSuggestion:
    'Once you created or interacted with entities, they will show up here and you can use this panel to quickly search and navigate to them.',
})

const categoryMap = {
  [APPLICATION]: {
    title: sharedMessages.applications,
  },
  [GATEWAY]: {
    title: sharedMessages.gateways,
  },
  [ORGANIZATION]: {
    title: sharedMessages.organizations,
  },
  [END_DEVICE]: {
    title: sharedMessages.devices,
  },
  bookmarks: {
    title: sharedMessages.bookmarks,
  },
  'top-entities': {
    title: sharedMessages.topEntities,
  },
}

const SearchPanel = ({
  onClose,
  onSelect,
  topEntities,
  searchResults,
  inline,
  onQueryChange,
  searchResultsFetching,
  topEntitiesFetching,
  searchQuery,
}) => {
  const listRef = useRef()
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebounce(query, 350, onQueryChange)
  const { formatMessage } = useIntl()
  const lastQuery = useRef('')
  const isTopEntitiesMode = debouncedQuery === ''

  let items

  // When in top entities mode, or fetching search results
  // transitioning from top entities mode, show the top entities.
  // The second part is necessary to avoid showing old search results.
  if (isTopEntitiesMode || (lastQuery.current === '' && debouncedQuery !== searchQuery)) {
    items = topEntities || []
    // In all other cases, show the search results.
  } else {
    items = searchResults || []
  }

  const itemCount = items.reduce((acc, item) => acc + item.items.length, 0)
  const noTopEntities =
    !topEntities || topEntities.length === 0 || topEntities.every(item => item.items.length === 0)

  // Keep track of the last query to determine when to switch between top entities and search results.
  useEffect(() => {
    if (isTopEntitiesMode) {
      lastQuery.current = ''
    } else if (debouncedQuery === searchQuery) {
      lastQuery.current = searchQuery
    }
  }, [isTopEntitiesMode, debouncedQuery, searchQuery])

  // Reset selected index when search results change or when switching modes.
  useEffect(() => {
    setSelectedIndex(0)
  }, [searchResults, isTopEntitiesMode])

  useEffect(() => {
    const handleKeyDown = event => {
      const listElement = listRef.current
      let newIndex = selectedIndex

      if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
        event.preventDefault()
        newIndex =
          event.key === 'ArrowDown'
            ? (selectedIndex + 1) % itemCount
            : (selectedIndex - 1 + itemCount) % itemCount
        setSelectedIndex(newIndex)

        const item = document.getElementById(`search-item-${newIndex}`)
        if (item) {
          const itemThreshold = item.clientHeight
          if (
            item.offsetTop + item.clientHeight >
            listElement.scrollTop + listElement.clientHeight - itemThreshold
          ) {
            listElement.scrollTop =
              item.offsetTop + item.clientHeight - listElement.clientHeight + itemThreshold
          } else if (item.offsetTop < listElement.scrollTop + itemThreshold) {
            listElement.scrollTop = item.offsetTop - itemThreshold
          }
        }
      } else if (event.key === 'Escape') {
        onClose()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [itemCount, items, onClose, onSelect, selectedIndex])

  const handleInputKeyDown = useCallback(
    event => {
      if (event.key === 'Escape') {
        onClose()
      } else if (event.key === 'Enter') {
        // Get DOM item and simulate click.
        document.getElementById(`search-item-${selectedIndex}`).click()
      }
    },
    [onClose, selectedIndex],
  )

  const handleQueryChange = useCallback(event => {
    setQuery(event.target.value)
  }, [])

  let i = 0

  return (
    <div className={classNames(style.container, { 'pos-static': inline })}>
      <div className="d-flex p-vert-cs-m p-sides-cs-xl gap-cs-s al-center">
        <Icon className="c-icon-neutral-normal" icon={IconSearch} />
        <input
          className={style.input}
          onKeyDown={handleInputKeyDown}
          value={query}
          onChange={handleQueryChange}
          placeholder={formatMessage(sharedMessages.typeToSearch)}
        />
        <Icon className={style.xOut} icon={IconX} onClick={onClose} />
      </div>
      <Overlay visible={searchResultsFetching} loading spinnerMessage={sharedMessages.searching}>
        <ScrollFader className={style.list} ref={listRef}>
          {topEntitiesFetching && (
            <div className={style.loading}>
              <Spinner after={0}>
                <Message content={m.fetchingTopEntities} />
              </Spinner>
            </div>
          )}
          {!topEntitiesFetching && noTopEntities && (
            <div className={style.noResults}>
              <LookingLuke className="d-block block-center" />
              <p className="c-text-neutral-heavy mt-0 mb-0 fs-l text-center">
                <Message content={m.noTopEntities} />
              </p>
              <div>
                <p className="text-center">
                  <Message content={m.noTopEntitiesSuggestion} />
                </p>
              </div>
            </div>
          )}
          {!topEntitiesFetching && itemCount === 0 && (
            <div className={style.noResults}>
              <LookingLuke className="d-block block-center" />
              <p className="c-text-neutral-heavy mt-0 mb-0 fs-l text-center">
                <Message content={m.noResultsFound} />
              </p>
              <div>
                <p className="text-center">
                  <Message content={m.noResultsSuggestion} />
                </p>
                <ul>
                  <li>
                    <Message content={sharedMessages.applications} />
                  </li>
                  <li>
                    <Message content={sharedMessages.gateways} />
                  </li>
                  <li>
                    <Message content={sharedMessages.organizations} />
                  </li>
                  <li>
                    <Message content={m.devices} values={{ lineBreak: <br /> }} />
                  </li>
                </ul>
              </div>
            </div>
          )}
          {items.reduce((acc, item, index) => {
            if (item.items.length > 0) {
              acc.push(
                <div className={style.resultHeader} key={`header-${index}`}>
                  <Message content={categoryMap[item.category].title} />
                </div>,
              )
              item.items.forEach(subitem => {
                acc.push(
                  <PanelItem
                    icon={entityIcons[subitem.type]}
                    title={subitem?.entity?.name || subitem.id}
                    subtitle={subitem.id}
                    key={subitem.id}
                    isFocused={i === selectedIndex}
                    index={i++}
                    path={subitem.path}
                    onClick={onClose}
                    onMouseEnter={setSelectedIndex}
                  />,
                )
              })
            }
            return acc
          }, [])}
        </ScrollFader>
      </Overlay>
      <div className={style.footer}>
        <div className="d-flex al-center gap-cs-xs">
          <Message
            content={m.instructions}
            values={{
              arrowKeys: (
                <div className="d-flex gap-cs-xxs">
                  <Icon className={style.icon} icon={IconArrowUp} small />
                  <Icon className={style.icon} icon={IconArrowDown} small />
                </div>
              ),
              enter: <Icon className={style.icon} icon={IconArrowBack} small />,
            }}
          />
        </div>
        <div>
          <Message content={m.searchingEntities} component="span" />
        </div>
      </div>
    </div>
  )
}

SearchPanel.propTypes = {
  inline: PropTypes.bool,
  onClose: PropTypes.func.isRequired,
  onQueryChange: PropTypes.func.isRequired,
  onSelect: PropTypes.func,
  searchQuery: PropTypes.string.isRequired,
  searchResults: PropTypes.arrayOf(
    PropTypes.shape({
      category: PropTypes.string.isRequired,
      items: PropTypes.arrayOf(PropTypes.unifiedEntity),
    }),
  ).isRequired,
  searchResultsFetching: PropTypes.bool.isRequired,
  topEntities: PropTypes.arrayOf(
    PropTypes.shape({
      category: PropTypes.string.isRequired,
      source: PropTypes.string.isRequired,
      items: PropTypes.arrayOf(PropTypes.unifiedEntity),
    }),
  ).isRequired,
  topEntitiesFetching: PropTypes.bool.isRequired,
}

SearchPanel.defaultProps = {
  inline: false,
  onSelect: () => null,
}

export default SearchPanel
