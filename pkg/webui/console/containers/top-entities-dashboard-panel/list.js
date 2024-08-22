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
import { defineMessages } from 'react-intl'

import { Table } from '@ttn-lw/components/table'
import ScrollFader from '@ttn-lw/components/scroll-fader'

import PropTypes from '@ttn-lw/lib/prop-types'

import EntitiesItem from './item'

import styles from './top-entities-panel.styl'

const m = defineMessages({
  empty: 'No entities yet',
})

const EntitiesList = ({
  entities,
  headers,
  renderWhenEmpty,
  EntitiesItemComponent: EntitiesItemProp,
}) => {
  const EntitiesItemComponent = EntitiesItemProp ?? EntitiesItem

  const rows = entities
    .slice(0, 10)
    .map(entity => <EntitiesItemComponent headers={headers} entity={entity} key={entity.id} />)

  const columns = (
    <Table.Row head panelStyle>
      {headers.map((header, key) => (
        <Table.HeadCell
          key={key}
          align={header.align}
          content={header.displayName}
          name={header.name}
          width={header.width}
          className={header.className}
          panelStyle
        />
      ))}
    </Table.Row>
  )

  return entities.length === 0 ? (
    renderWhenEmpty
  ) : (
    <ScrollFader className={styles.scrollFader} faderHeight="4rem" topFaderOffset="3rem" light>
      <Table className={styles.table}>
        <Table.Head className={styles.topEntitiesPanelOuterTableHeader} panelStyle>
          {columns}
        </Table.Head>
        <Table.Body emptyMessage={m.empty}>{rows}</Table.Body>
      </Table>
    </ScrollFader>
  )
}

EntitiesList.propTypes = {
  EntitiesItemComponent: PropTypes.func,
  entities: PropTypes.unifiedEntities.isRequired,
  headers: PropTypes.arrayOf(
    PropTypes.shape({
      align: PropTypes.string,
      displayName: PropTypes.message,
      name: PropTypes.string,
      width: PropTypes.string,
      className: PropTypes.string,
    }),
  ).isRequired,
  renderWhenEmpty: PropTypes.node,
}

EntitiesList.defaultProps = {
  renderWhenEmpty: null,
  EntitiesItemComponent: undefined,
}

export default EntitiesList
