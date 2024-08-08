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

import React, { useCallback, useEffect } from 'react'

import Dropdown from '@ttn-lw/components/dropdown'

import PropTypes from '@ttn-lw/lib/prop-types'

import Portal from './portal'

const AttachedDropdown = ({
  attachedRef,
  onItemsClick,
  hover,
  onOutsideClick,
  portalled,
  ...rest
}) => {
  const [open, setOpen] = React.useState(false)

  // Add event listeners to open the dropdown.
  useEffect(() => {
    if (!attachedRef.current) {
      return
    }
    const openEvent = hover ? 'mouseenter' : 'click'
    const node = attachedRef.current
    const toggleDropdown = () => {
      // Add escape key event listener to close the dropdown
      setOpen(val => !val)
    }
    const closeDropdown = () => setOpen(false)

    node.addEventListener(openEvent, toggleDropdown)
    if (hover) {
      node.addEventListener('mouseleave', closeDropdown)
    }
    return () => {
      node.removeEventListener(openEvent, toggleDropdown)
      if (hover) {
        node.removeEventListener('mouseleave', closeDropdown)
      }
    }
  }, [attachedRef, hover, open])

  // Add escape key event listener to close the dropdown.
  useEffect(() => {
    if (open) {
      const closeDropdown = e => {
        if (e.key === 'Escape') {
          setOpen(false)
        }
      }
      document.addEventListener('keydown', closeDropdown)
      return () => {
        document.removeEventListener('keydown', closeDropdown)
      }
    }
    return
  }, [open])

  const handleItemsClick = useCallback(() => {
    setOpen(false)
    onItemsClick()
  }, [onItemsClick])

  const handleOutsideClick = useCallback(
    e => {
      if (attachedRef.current && attachedRef.current.contains(e.target)) {
        // Ignore clicks on the attached element, so that toggling is possible.
        return
      }
      setOpen(false)
      onOutsideClick()
    },
    [attachedRef, onOutsideClick],
  )

  if (portalled) {
    return (
      <Portal visible={open} setOpen={setOpen} positionReference={attachedRef}>
        <Dropdown
          open={open}
          onItemsClick={handleItemsClick}
          onOutsideClick={handleOutsideClick}
          hover={hover}
          {...rest}
        />
      </Portal>
    )
  }

  return (
    <Dropdown
      open={open}
      onItemsClick={handleItemsClick}
      onOutsideClick={handleOutsideClick}
      hover={hover}
      {...rest}
    />
  )
}

AttachedDropdown.propTypes = {
  attachedRef: PropTypes.shape({ current: PropTypes.instanceOf(Element) }).isRequired,
  hover: PropTypes.bool,
  onItemsClick: PropTypes.func,
  onOutsideClick: PropTypes.func,
  portalled: PropTypes.bool,
  positionReference: PropTypes.shape({}),
}

AttachedDropdown.defaultProps = {
  onItemsClick: () => null,
  onOutsideClick: () => null,
  hover: false,
  portalled: false,
  positionReference: undefined,
}

export default AttachedDropdown
