// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import m from './messages'

import style from './events.styl'

const EventErrorBoundary = ({ children }) => {
  const [hasErrored, setHasErrored] = useState(false)

  const handleError = useCallback(() => {
    setHasErrored(true)
  }, [])

  if (hasErrored) {
    return (
      <div className={style.cellError}>
        <Icon icon="error" className={style.eventIcon} />
        <Message content={m.errorOverviewEntry} />
      </div>
    )
  }

  return (
    <React.Fragment>
      {React.Children.map(children, child => {
        if (!child) {
          return null
        }
        return React.cloneElement(child, {
          onError: handleError,
        })
      })}
    </React.Fragment>
  )
}

EventErrorBoundary.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
}

export default EventErrorBoundary
