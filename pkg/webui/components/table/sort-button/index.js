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

import style from './sort-button.styl'

const SortButton = function ({
  className,
  active,
  title,
  onSort,
  name,
  direction,
  ...rest
}) {
  const buttonClassNames = classnames(className, style.button, {
    [style.buttonActive]: active,
    [style.buttonDesc]: active && direction === 'desc',
  })

  return (
    <button
      className={buttonClassNames}
      type="button"
      onClick={onSort}
      {...rest}
    >
      <Message content={title} />
      <Icon className={style.icon} icon="sort" />
    </button>
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
