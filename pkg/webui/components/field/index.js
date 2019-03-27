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
import { injectIntl } from 'react-intl'

import PropTypes from '../../lib/prop-types'
import from from '../../lib/from'
import { warn } from '../../lib/log'
import getByPath from '../../lib/get-by-path'

import Icon from '../icon'
import Input from '../input'
import Checkbox from '../checkbox'
import Select from '../select'
import Message from '../../lib/components/message'

import style from './field.styl'

const inputAllowedProps = [
  'name',
  'placeholder',
  'label',
  'value',
  'type',
  'autoFocus',
  'min',
  'max',
  'icon',
  'onFocus',
  'onBlur',
  'onChange',
  'onEnter',
  'placeholder',
  'error',
  'warning',
  'valid',
  'disabled',
  'readOnly',
  'loading',
  'autoComplete',
]

const checkboxAllowedProps = [
  'value',
  'onFocus',
  'onBlur',
  'onChange',
  'disabled',
]

const selectAllowedProps = [
  'autoFocus',
  'clearValue',
  'getStyles',
  'getValue',
  'hasValue',
  'isDisabled',
  'isMulti',
  'isRtl',
  'isSearchable',
  'onChange',
  'options',
  'placeholder',
  'selectOption',
  'selectProps',
  'setValue',
  'value',
]

const getAllowedPropsByType = function (type) {
  switch (type) {
  case 'checkbox':
    return checkboxAllowedProps
  case 'select':
    return selectAllowedProps
  default:
    return inputAllowedProps
  }
}

const filterPropsByType = function (type, props) {
  const allowedTypes = getAllowedPropsByType(type)

  const res = {}
  for (const key of allowedTypes) {
    if (key in props) {
      res[key] = props[key]
    }
  }

  return res
}

const component = function (type) {
  switch (type) {
  case 'checkbox':
    return Checkbox

  case 'text':
  case 'number':
  case 'password':
  case 'byte':
    return Input

  case 'select':
    return Select

  default:
    warn('No type matches', type)
    return Input
  }
}

const Field = function (props) {

  const handleChange = function (value) {
    props.setFieldValue(props.name, value)
    if (props.validateOnChange) {
      props.setFieldTouched(props.name, true)
    }
  }

  const handleBlur = function (e) {
    // Always regard inputs that never received a value as untouched (better UX)
    if (e.target.value !== '' && props.validateOnBlur) {
      props.setFieldTouched(props.name, true)
    }
  }

  const {
    className,
    type = 'text',
    name = '',
    title,
    placeholder = props.title,
    description = null,
    error,
    warning,
    touched,
    horizontal = false,
    disabled = false,
    readOnly = false,
    required = false,
    form = true,
    ...rest
  } = props

  // Underscored assignment due to naming conflict

  let _error = props.error
  let _touched = props.touched
  const formatMessage = content => typeof content === 'object' ? props.intl.formatMessage(content) : content

  if (form) {
    const {
      errors = {},
    } = props

    // preserve default values for different inputs
    _error = getByPath(errors, name)
    _touched = touched && touched[name]
    rest.onChange = handleChange
    rest.onBlur = handleBlur

    // Dismiss non boolean values for checkboxes
    if (type === 'checkbox') {
      rest.value = typeof rest.value === 'boolean' ? rest.value : false
    }

    // restore the rest object for future per component filtering
    rest.name = name
    rest.readOnly = readOnly
    rest.disabled = disabled
    rest.error = _touched && Boolean(_error)
    rest.warning = Boolean(warning)
    rest.type = type
    rest.placeholder = placeholder ? formatMessage(placeholder) : ''
  }

  const hasMessages = _touched && (_error || warning)

  const classname = classnames(className, style.field, style[type], ...from(style, {
    error: _error,
    warning: warning && !_error,
    horizontal,
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
        id={name}
        {...filterPropsByType(type, rest)}
      />
      {hasMessages
        ? <div className={style.messages}>
          <Err error={_error} name={title} />
          <Err warning={warning} name={title} />
        </div>
        : <Message className={style.description} content={description} />
      }
    </div>
  )
}

Field.propTypes = {
  /** The field title, displayed with the input */
  title: PropTypes.oneOfType([
    PropTypes.node,
    PropTypes.message,
  ]).isRequired,
  /** An optional helper description displayed under the input */
  description: PropTypes.oneOfType([
    PropTypes.node,
    PropTypes.message,
  ]),
  /** "name" prop applied to the input */
  name: PropTypes.string.isRequired,
  /**
   * The field type (eg. text, byte, password, checkbox), thunked values are
   * allowed as well
   */
  type: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.func,
  ]).isRequired,
  /** Error to be displayed next to input */
  error: PropTypes.error,
  /** Warning to be displayed next to input */
  warning: PropTypes.message,
  /** Whether the field is required */
  required: PropTypes.bool,
  /** Whether the field is disabled */
  disabled: PropTypes.bool,
  /** Whether the field is read only */
  readOnly: PropTypes.bool,
  /** Whether the field should be displayed in horizontal style */
  horizontal: PropTypes.bool,
  /** Whether the field is part of a Form component (default is true).
   * This is necessary to map form values correctly.
   */
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
      <Icon icon={icon} className={style.icon} />
      <Message
        content={content.format || content.error_description || content.message || content}
        values={{
          name: <Message content={name} className={style.name} />,
        }}
      />
    </div>
  )
}

export default injectIntl(Field)
export { Field, Err as FieldError }
