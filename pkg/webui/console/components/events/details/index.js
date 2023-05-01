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

import Notification from '@ttn-lw/components/notification'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import { getEventId } from '../utils'
import messages from '../messages'

import RawEventDetails from './raw'
import SyntheticEventDetails from './synthetic'

import style from './details.styl'

const EventDetails = ({ className, children, event }) => {
  const hasChildren = Boolean(children)
  const dataFormatsDocumentationLink = (
    <Link.DocLink
      secondary
      path="/the-things-stack/concepts/data-formats/"
      title={messages.dataFormats}
    >
      <Message content={messages.dataFormats} />
    </Link.DocLink>
  )

  if (!Boolean(event)) {
    return (
      <div className={classnames(className, style.details)}>
        <Notification content={messages.eventUnavailable} warning small />
      </div>
    )
  }

  return (
    <div className={classnames(className, style.details)}>
      {!hasChildren ? (
        event.isSynthetic ? (
          <SyntheticEventDetails event={event} id={getEventId(event)} />
        ) : (
          <RawEventDetails className={style.codeEditor} details={event} id={getEventId(event)} />
        )
      ) : (
        children
      )}
      <Message
        content={messages.dataFormatsInformation}
        values={{ dataFormatsDocumentationLink }}
      />
    </div>
  )
}

EventDetails.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  className: PropTypes.string,
  event: PropTypes.event,
}

EventDetails.defaultProps = {
  children: null,
  className: undefined,
  event: undefined,
}

export default EventDetails
