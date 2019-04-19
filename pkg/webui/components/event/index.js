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
import classnames from 'classnames'

import Icon from '../icon'
import DateTime from '../../lib/components/date-time'
import PropTypes from '../../lib/prop-types'

import style from './event.styl'

const Event = function (props) {
  const {
    className,
    icon,
    time,
    emitter,
    content,
  } = props

  const eventIcon = React.isValidElement(icon)
    ? React.cloneElement(icon, {
      ...icon.props,
      className: classnames(style.eventIcon, icon.props.className, {
        [style.eventIconDefault]: !icon.props.className,
      }),
    })
    : (
      <Icon
        className={classnames(style.eventIcon, style.eventIconDefault)}
        icon={icon}
      />
    )

  const eventContent = React.isValidElement(content)
    ? React.cloneElement(content, {
      ...content.props,
      className: classnames(style.eventContent, content.props.className),
    })
    : <div className={style.eventContent}>{content}</div>

  return (
    <div className={classnames(className, style.event)}>
      <div className={style.eventHeader}>
        {eventIcon}
        <DateTime
          className={style.eventTime}
          value={time}
          date={false}
        />
        <span className={style.eventEmitter}>{emitter}</span>
        {eventContent}
      </div>
    </div>
  )
}

Event.propTypes = {
  /** The time of the event. */
  time: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.number,
    PropTypes.instanceOf(Date),
  ]).isRequired,
  /** The icon of the event. */
  icon: PropTypes.node,
  /** The entity identifier of the event. */
  emitter: PropTypes.string.isRequired,
  /** Custom content of the event. */
  content: PropTypes.node.isRequired,
}

Event.defaultProps = {
  icon: 'event',
  content: null,
}

export default Event
