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

import Message from '../../../../lib/components/message'
import Event from '../..'
import PropTypes from '../../../../lib/prop-types'
import { getEntityId } from '../../../../lib/selectors/id'

class DefaultEvent extends React.PureComponent {
  render() {
    const { className, event, widget, overviewClassName, expandedClassName } = this.props

    const entityId = getEntityId(event.identifiers[0])

    const content = <Message content={{ id: `event:${event.name}` }} />

    return (
      <Event
        className={className}
        overviewClassName={overviewClassName}
        expandedClassName={expandedClassName}
        time={event.time}
        content={content}
        emitter={entityId}
        widget={widget}
      />
    )
  }
}

DefaultEvent.propTypes = {
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool,
}

DefaultEvent.defaultProps = {
  widget: false,
}

export default DefaultEvent
