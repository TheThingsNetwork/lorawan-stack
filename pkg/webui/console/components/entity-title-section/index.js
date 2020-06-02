// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { Container, Row, Col } from 'react-grid-system'
import classnames from 'classnames'
import { defineMessages } from 'react-intl'

import Status from '@ttn-lw/components/status'

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import { IconValueTag } from '@console/components/key-value-tag'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './entity-title-section.styl'

const m = defineMessages({
  lastSeenUnavailable: 'Last seen info unavailable',
})

const EntityTitleSection = ({ entityName, entityId, description, creationDate, children }) => {
  return (
    <React.Fragment>
      <Container>
        <Row>
          <Col md={12} className={style.container}>
            <h1 className={style.title}>{entityName || entityId}</h1>
            {description && <span className={style.description}>{description}</span>}
            <div className={style.bottom}>
              <div className={style.children}>{children}</div>
              <div className={style.creationDate}>
                <span>
                  <Message content={sharedMessages.created} />{' '}
                  <DateTime.Relative value={creationDate} />
                </span>
              </div>
            </div>
          </Col>
        </Row>
      </Container>
      <hr className={style.hRule} />
    </React.Fragment>
  )
}

EntityTitleSection.propTypes = {
  children: PropTypes.node.isRequired,
  creationDate: PropTypes.string.isRequired,
  description: PropTypes.string,
  entityId: PropTypes.string.isRequired,
  entityName: PropTypes.string,
}

EntityTitleSection.defaultProps = {
  entityName: undefined,
  description: undefined,
}

EntityTitleSection.Device = ({
  deviceName,
  deviceId,
  description,
  children,
  lastSeen,
  downlinkFrameCount,
  uplinkFrameCount,
}) => {
  return (
    <Container>
      <Row>
        <Col>
          <div
            className={classnames(style.containerDevice, {
              [style.hasDescription]: Boolean(description),
            })}
          >
            <h1 className={style.title}>{deviceName || deviceId}</h1>
            <span className={style.belowTitle}>
              {Boolean(lastSeen) ? (
                <Status status="good" flipped>
                  <Message content={sharedMessages.lastSeen} />{' '}
                  <DateTime.Relative value={lastSeen} />
                </Status>
              ) : (
                <Status status="mediocre" label={m.lastSeenUnavailable} flipped />
              )}
              {Boolean(uplinkFrameCount || downlinkFrameCount) && (
                <React.Fragment>
                  <div className={style.messages}>
                    {uplinkFrameCount && (
                      <IconValueTag
                        iconClassName={style.icon}
                        icon="uplink"
                        value={uplinkFrameCount}
                        tooltipMessage={sharedMessages.uplinkFrameCount}
                      />
                    )}
                    {downlinkFrameCount && (
                      <IconValueTag
                        iconClassName={style.icon}
                        icon="downlink"
                        value={downlinkFrameCount}
                        tooltipMessage={sharedMessages.downlinkFrameCount}
                      />
                    )}
                  </div>
                </React.Fragment>
              )}
            </span>
            {description && <span className={style.description}>{description}</span>}
          </div>
          {children}
        </Col>
      </Row>
    </Container>
  )
}

EntityTitleSection.Device.propTypes = {
  children: PropTypes.node.isRequired,
  description: PropTypes.string,
  deviceId: PropTypes.string.isRequired,
  deviceName: PropTypes.string,
  downlinkFrameCount: PropTypes.number,
  lastSeen: PropTypes.string,
  uplinkFrameCount: PropTypes.number,
}

EntityTitleSection.Device.defaultProps = {
  deviceName: undefined,
  downlinkFrameCount: undefined,
  uplinkFrameCount: undefined,
  description: undefined,
  lastSeen: undefined,
}

export default EntityTitleSection
