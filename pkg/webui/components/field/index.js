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
import PropTypes from 'prop-types'

import from from '../../lib/from'

import Icon from '../icon'
import Input from '../input'
import Checkbox from '../checkbox'
import Message from '../message'

import style from './field.styl'

const component = function (type) {
  switch (type) {
  case 'checkbox':
    return Checkbox

  case 'text':
  case 'number':
  case 'password':
  case 'byte':
    return Input

  default:
    console.warn('No type matches', type)
    return Input
  }
}

const Field = function (props) {

  const handleChange = function (value) {
    props.setFieldValue(props.name, value)
  }

  const handleBlur = function () {
    props.setFieldTouched(props.name, true)
  }

  const {
    type = 'text',
    name = '',
    title = null,
    placeholder = props.title,
    description = null,
    error,
    value,
    warning,
    touched,
    disabled = false,
    readOnly = false,
    required = false,
    form = true,
    ...rest
  } = props

  // Underscored assignment due to naming conflict

  let _value = props.value
  let _error = props.error
  let _touched = props.touched


  if (form) {
    const {
      values = {},
      errors = {},
    } = props

    _value = values[name]
    _error = errors[name]
    _touched = touched[name]
    rest.value = _value
    rest.onChange = handleChange
    rest.onBlur = handleBlur
  }

  const hasMessages = _touched && (_error || warning)

  const classname = classnames(style.field, style[type], ...from(style, {
    error: _error,
    warning: warning && !_error,
    required,
    readOnly,
    disabled,
  }))

  const Component = component(type)

  return (
    <div className={classname}>
      <label className={style.label} htmlFor={name}>
        <Message content={title} className={style.title} />
        <span className={style.reqicon}>&middot;</span>
      </label>
      <Component
        className={style.component}
        name={name}
        id={name}
        readOnly={readOnly}
        disabled={disabled}
        error={_touched && Boolean(_error)}
        warning={Boolean(warning)}
        type={type}
        placeholder={placeholder}
        {...rest}
      />
      {hasMessages
        ? <div className={style.messages}>
          <Err error={_error} name={title} />
          <Err warning={warning} name={title} />
        </div>
        : <div className={style.description}>
          <Message content={description} />
        </div>}
    </div>
  )
}

Field.propTypes = {
  title: PropTypes.oneOfType([
    PropTypes.node,
    PropTypes.object,
  ]).isRequired,
  description: PropTypes.node,
  name: PropTypes.string.isRequired,
  type: PropTypes.oneOfType([
    PropTypes.string.isRequired,
    PropTypes.func.isRequired,
  ]),
  error: PropTypes.string,
  warning: PropTypes.string,
  required: PropTypes.bool,
  disabled: PropTypes.bool,
  readOnly: PropTypes.bool,
  form: PropTypes.bool,
}

const Err = function (props) {
  const {
    error,
    warning,
    name,
  } = props

  const content = error || warning || ''

  const icon = error ? 'error' : 'warning'

  const classname = classnames(style.message, {
    [style.show]: content && content !== '',
    [style.hide]: !content || content === '',
    [style.err]: error,
    [style.warn]: warning,
  })

  return (
    <div className={classname}>
      <Icon icon={icon} nudgeBottom className={style.icon} />
      <Message
        content={content.format || content.error_description || content.message || content}
        values={{
          name: <Message content={name} className={style.name} />,
        }}
      />
    </div>
  )
}

export default Field
