// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import PropTypes from '../../../lib/prop-types'
import Message from '../../../lib/components/message'
import Icon from '../../icon'
import { TableConsumer } from '../context'

import style from './sort-button.styl'

const SortButton = function ({
  className,
  active,
  title,
  name,
  direction,
  ...rest
}) {
  const buttonClassNames = classnames(className, style.button, {
    [style.buttonActive]: active,
  })
  const iconClassNames = classnames(style.icon, {
    [style.iconAsc]: active && direction === 'asc',
  })

  return (
    <TableConsumer>
      {
        ({ onRequestSort }) =>
          (
            <button
              className={buttonClassNames}
              type="button"
              onClick={function () {
                onRequestSort(name)
              }}
              {...rest}
            >
              <Message content={title} />
              <Icon className={iconClassNames} icon="sort" />
            </button>
          )
      }
    </TableConsumer>
  )
}

SortButton.propTypes = {
  /** A flag identifying whether the button is active */
  active: PropTypes.bool.isRequired,
  /** The text of the button */
  title: PropTypes.message.isRequired,
  /** Function to be called when the button is pressed */
  onSort: PropTypes.func.isRequired,
  /** The current ordering (ascending/descending/none) */
  direction: PropTypes.string,
}

export default SortButton
