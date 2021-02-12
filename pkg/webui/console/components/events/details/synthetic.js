// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import Notification from '@ttn-lw/components/notification'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import { eventMessages } from '@console/lib/events/definitions'

import { getEventId } from '../utils'
import messages from '../messages'

import RawEventDetails from './raw'

import style from './details.styl'

const SyntheticEventDetails = ({ event }) => (
  <>
    <Notification content={messages.syntheticEvent} info small />
    {eventMessages[`${event.name}:details`] && (
      <div>
        <Message
          component="h3"
          content={eventMessages[`${event.name}:type`]}
          className={style.eventType}
        />
        <Message
          component="p"
          content={eventMessages[`${event.name}:details`]}
          className={style.eventDescription}
        />
      </div>
    )}
    <Message component="h4" content={messages.rawEvent} />
    <RawEventDetails className={style.codeEditor} details={event} id={getEventId(event)} />
  </>
)

SyntheticEventDetails.propTypes = {
  event: PropTypes.event.isRequired,
}

export default SyntheticEventDetails
