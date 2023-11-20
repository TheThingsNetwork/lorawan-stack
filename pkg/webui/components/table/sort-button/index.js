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

import React, { useCallback } from 'react'
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './sort-button.styl'

const SortButton = ({ name, onSort, className, active, direction, title }) => {
  const handleSort = useCallback(() => {
    onSort(name)
  }, [name, onSort])

  const buttonClassNames = classnames(className, style.button, {
    [style.buttonActive]: active,
    [style.buttonDesc]: active && direction === 'desc',
  })

  return (
    <button className={buttonClassNames} type="button" onClick={handleSort}>
      <Message content={title} />
      <Icon className={style.icon} icon="sort" nudgeUp />
    </button>
  )
}

SortButton.propTypes = {
  /** A flag identifying whether the button is active. */
  active: PropTypes.bool.isRequired,
  className: PropTypes.string,
  /** The current ordering (ascending/descending/none). */
  direction: PropTypes.string,
  /** The name of the column that the sort button represents. */
  name: PropTypes.string.isRequired,
  /** Function to be called when the button is pressed. */
  onSort: PropTypes.func.isRequired,
  /** The text of the button. */
  title: PropTypes.message.isRequired,
}

SortButton.defaultProps = {
  className: undefined,
  direction: undefined,
}

export default SortButton
