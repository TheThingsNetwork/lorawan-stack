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
import PropTypes from 'prop-types'
import classnames from 'classnames'

import Spinner from '../spinner'
import Message from '../message'
import Icon from '../icon'

import style from './button.styl'

const Button = function ({
  message,
  danger,
  secondary,
  naked,
  icon,
  busy,
  className,
  onClick,
  error,
  ...rest
}) {

  const classname = classnames(style.button, className, {
    [style.danger]: danger,
    [style.secondary]: secondary,
    [style.naked]: naked,
    [style.busy]: busy,
    [style.withIcon]: icon !== undefined,
    [style.onlyIcon]: icon !== undefined && !message,
    [style.error]: error && !busy,
  })

  const classnameIcon = classnames({
    [style.icon]: message !== undefined,
  })

  const handleClick = function (evt) {
    if (busy || rest.disabled) {
      return
    }

    onClick(evt)
  }

  return (
    <button
      className={classname}
      onClick={handleClick}
      {...rest}
    >
      <div className={style.content}>
        {icon ? <Icon className={classnameIcon} nudgeUp icon={icon} /> : null}
        {busy ? <Spinner className={style.spinner} small after={200} /> : null}
        {message ? <Message content={message} /> : null}
      </div>
    </button>
  )
}

Button.propTypes = {
  message: PropTypes.string,
  onClick: PropTypes.func,
  danger: PropTypes.bool,
  boring: PropTypes.bool,
  busy: PropTypes.bool,
}

Button.defaultProps = {
  onClick: () => null,
}

export default Button
