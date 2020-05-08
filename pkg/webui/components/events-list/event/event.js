// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import PropTypes from '@ttn-lw/lib/prop-types'

import eventsContext from '../context'

import eventContext from './context'

const Event = props => {
  const { event, children, expandable } = props

  const { widget } = React.useContext(eventsContext)
  const [isOpen, setOpen] = React.useState(false)
  const handleDetailsOpen = React.useCallback(() => {
    setOpen(open => !open)
  }, [setOpen])

  return (
    <eventContext.Provider
      value={{ event, isOpen, onDetailsOpen: handleDetailsOpen, expandable, widget }}
    >
      {children}
    </eventContext.Provider>
  )
}

Event.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  event: PropTypes.event.isRequired,
  expandable: PropTypes.bool,
}

Event.defaultProps = {
  expandable: false,
}

export default Event
