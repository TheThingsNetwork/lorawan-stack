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
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'

import PropTypes from '@ttn-lw/lib/prop-types'

import eventContext from '../context'

import RawEventDetails from './raw'

import style from './details.styl'

const m = defineMessages({
  showRaw: 'Show raw payload',
  hideRaw: 'Hide raw payload',
})

const EventDetails = props => {
  const { className, children } = props
  const { isOpen, event, expandable, widget } = React.useContext(eventContext)

  const hasChildren = Boolean(children)
  const hasData = 'data' in event

  const [isRaw, setRaw] = React.useState(!hasChildren)
  const handleViewChange = React.useCallback(() => {
    setRaw(raw => !raw)
  }, [setRaw])

  if (!isOpen || !expandable || widget) {
    return null
  }

  return (
    <div className={classnames(className, style.details)}>
      <div>
        {hasChildren && hasData && (
          <Button
            className={style.rawButton}
            naked
            secondary
            message={isRaw ? m.hideRaw : m.showRaw}
            onClick={handleViewChange}
          />
        )}
      </div>
      {isRaw && hasData ? (
        <RawEventDetails details={event.data} id={`${event.time}-${event.name}`} />
      ) : (
        children
      )}
    </div>
  )
}

EventDetails.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  className: PropTypes.string,
}

EventDetails.defaultProps = {
  children: null,
  className: undefined,
}

export default EventDetails
