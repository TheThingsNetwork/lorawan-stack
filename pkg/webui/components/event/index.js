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
import bind from 'autobind-decorator'
import { defineMessages } from 'react-intl'

import Icon from '../icon'
import Message from '../../lib/components/message'
import DateTime from '../../lib/components/date-time'
import PropTypes from '../../lib/prop-types'
import CodeEditor from '../code-editor'
import CRUDEvent from './types/crud'
import DefaultEvent from './types/default'
import MessageEvent from './types/message'

import style from './event.styl'

const m = defineMessages({
  eventData: 'Event Data',
})

@bind
class Event extends React.PureComponent {
  state = {
    expanded: false,
  }

  handleEventExpand() {
    this.setState(prev => ({
      expanded: !prev.expanded,
    }))
  }

  get overview() {
    const { icon, time, emitter, content, data, widget, overviewClassName } = this.props
    const { expanded } = this.state

    const eventIcon = React.isValidElement(icon) ? (
      React.cloneElement(icon, {
        ...icon.props,
        className: classnames(style.overviewIcon, icon.props.className, {
          [style.overviewIconDefault]: !icon.props.className,
        }),
      })
    ) : (
      <Icon className={classnames(style.overviewIcon, style.overviewIconDefault)} icon={icon} />
    )

    const eventContent = React.isValidElement(content) ? (
      React.cloneElement(content, {
        ...content.props,
        className: classnames(style.overviewContent, content.props.className),
      })
    ) : (
      <div className={style.overviewContent}>{content}</div>
    )

    const expandable = !widget && data

    let expandProps = {}
    let expandIcon = null
    if (expandable) {
      expandProps = { role: 'button', onClick: this.handleEventExpand }
      const iconCls = classnames(style.overviewIcon, style.overviewIconExpand)
      expandIcon = <Icon className={iconCls} icon={expanded ? 'expand_less' : 'expand_more'} />
    }

    const cls = classnames(style.overview, overviewClassName, {
      [style.expandable]: expandable,
    })

    return (
      <div className={cls} {...expandProps}>
        {eventIcon}
        <DateTime className={style.overviewTime} value={time} date={false} />
        <span className={style.overviewEmitter}>{emitter}</span>
        {eventContent}
        {expandIcon}
      </div>
    )
  }

  get expanded() {
    const { data, widget, emitter, time, expandedClassName } = this.props
    const { expanded } = this.state

    if (widget || !data || !expanded) {
      return null
    }

    const formattedData = JSON.stringify(data, null, 2)

    return (
      <div className={expandedClassName}>
        <Message content={m.eventData} component="h4" />
        <CodeEditor readOnly name={`${emitter}-${time}`} language="json" value={formattedData} />
      </div>
    )
  }

  render() {
    const { className } = this.props

    return (
      <div className={className}>
        {this.overview}
        {this.expanded}
      </div>
    )
  }
}

Event.propTypes = {
  /** The time of the event. */
  time: PropTypes.oneOfType([PropTypes.string, PropTypes.number, PropTypes.instanceOf(Date)])
    .isRequired,
  /** The icon of the event. */
  icon: PropTypes.node,
  /** The entity identifier of the event. */
  emitter: PropTypes.string.isRequired,
  /** Custom content of the event. */
  content: PropTypes.node.isRequired,
  /**
   * A flag identifying whether the event is displayed within the
   * events widget component. This disabled the expanded view.
   */
  widget: PropTypes.bool,
  /**
   * A stringified data of the event to be displayed in the
   * expanded view.
   */
  data: PropTypes.object,
  /** Additional styling for the event overview */
  overviewClassName: PropTypes.string,
  /** Additional styling for the event expanded view */
  expandedClassName: PropTypes.string,
}

Event.defaultProps = {
  icon: 'event',
  content: null,
  widget: false,
  data: null,
}

Event.CRUD = CRUDEvent
Event.Default = DefaultEvent
Event.Message = MessageEvent

export default Event
