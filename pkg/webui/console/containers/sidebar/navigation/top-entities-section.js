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
import { defineMessages } from 'react-intl'

import { APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

import { IconPlus, entityIcons } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import SideNavigation from '@ttn-lw/components/sidebar/side-menu'
import SectionLabel from '@ttn-lw/components/sidebar/section-label'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getTopEntities } from '@console/store/actions/top-entities'

const m = defineMessages({
  topGateways: 'Top gateways',
  topApplications: 'Top applications',
  noTopGateways: 'No top gateways yet',
  noTopApplications: 'No top applications yet',
  noTopDevices: 'No top end devices yet',
})

const SectionError = () => (
  <Message
    className="text-center fs-s c-text-neutral-light"
    content={sharedMessages.topEntitiesError}
  />
)

const TopEntityItem = ({ id, entity = {}, path, type, forType }) => (
  <SideNavigation.Item
    title={entity?.name || (forType === 'END_DEVICE' ? id.split('/')[1] : id)}
    icon={entityIcons[type]}
    path={path}
  />
)

TopEntityItem.propTypes = {
  entity: PropTypes.shape({
    name: PropTypes.string,
  }),
  forType: PropTypes.string.isRequired,
  id: PropTypes.string.isRequired,
  path: PropTypes.string.isRequired,
  type: PropTypes.string.isRequired,
}

TopEntityItem.defaultProps = {
  entity: undefined,
}

const TopEntitiesSection = ({ topEntities, type }) => {
  const [showMore, setShowMore] = useState(false)

  const handleShowMore = useCallback(async () => {
    setShowMore(showMore => !showMore)
  }, [])

  let label = sharedMessages.topEntities
  let noneLabel = sharedMessages.noTopEntities
  if (type === GATEWAY) {
    label = m.topGateways
    noneLabel = m.noTopGateways
  } else if (type === APPLICATION) {
    label = m.topApplications
    noneLabel = m.noTopApplications
  } else if (type === END_DEVICE) {
    label = sharedMessages.topDevices
    noneLabel = m.noTopDevices
  }
  return (
    <SideNavigation>
      <SectionLabel
        label={topEntities.length === 0 ? noneLabel : label}
        icon={IconPlus}
        type={type}
      />
      <RequireRequest
        requestAction={getTopEntities()}
        spinnerProps={{ inline: true, micro: true, center: true, className: 'mt-ls-s' }}
        errorRenderFunction={SectionError}
      >
        {topEntities.slice(0, 6).map((topEntity, index) => (
          <TopEntityItem key={index} {...topEntity} forType={type} />
        ))}
        {showMore &&
          topEntities.length > 6 &&
          topEntities
            .slice(6, topEntities.length)
            .map((topEntity, index) => <TopEntityItem key={index} {...topEntity} />)}
        {topEntities.length > 6 && (
          <Button
            message={showMore ? sharedMessages.showLess : sharedMessages.showMore}
            onClick={handleShowMore}
            className="c-text-neutral-light ml-cs-xs fs-s mt-cs-xs"
          />
        )}
      </RequireRequest>
    </SideNavigation>
  )
}

TopEntitiesSection.propTypes = {
  topEntities: PropTypes.unifiedEntities,
  type: PropTypes.string,
}

TopEntitiesSection.defaultProps = {
  topEntities: [],
  type: undefined,
}

export default TopEntitiesSection
