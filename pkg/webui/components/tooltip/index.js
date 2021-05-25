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

import React, { useCallback } from 'react'
import Tippy from '@tippyjs/react'

import PropTypes from '@ttn-lw/lib/prop-types'

import './tooltip.styl'

let currentInstance

const popperModifiers = [
  {
    name: 'offset',
    options: {
      offset: [-12, 8],
    },
  },
  {
    name: 'arrow',
    options: {
      element: '.tippy-arrow',
      padding: 12,
    },
  },
]

const Tooltip = props => {
  const { className, children, content, interactive, placement, delay, onShow } = props

  const handleShow = useCallback(
    instance => {
      if (currentInstance) {
        currentInstance.hide()
      }
      currentInstance = instance

      onShow(instance)
    },
    [onShow],
  )
  return (
    <Tippy
      className={className}
      content={content}
      interactive={interactive}
      placement={placement}
      popperOptions={{ modifiers: popperModifiers }}
      delay={delay}
      onShow={handleShow}
      animation="fade"
      duration={250}
    >
      {children}
    </Tippy>
  )
}

Tooltip.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  content: PropTypes.node.isRequired,
  delay: PropTypes.oneOfType([PropTypes.number, PropTypes.arrayOf([PropTypes.number])]),
  interactive: PropTypes.bool,
  onShow: PropTypes.func,
  placement: PropTypes.oneOf([
    'top',
    'top-start',
    'top-end',
    'right',
    'right-start',
    'right-end',
    'bottom',
    'bottom-start',
    'bottom-end',
    'left',
    'left-start',
    'left-end',
    'auto',
    'auto-start',
    'auto-end',
  ]),
}

Tooltip.defaultProps = {
  className: '',
  interactive: false,
  placement: 'bottom',
  delay: 300,
  onShow: () => null,
}

export default Tooltip
