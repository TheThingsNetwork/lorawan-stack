// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { useField } from 'formik'
import { isPlainObject } from 'lodash'

import Message from '@ttn-lw/lib/components/message'

import from from '@ttn-lw/lib/from'
import PropTypes from '@ttn-lw/lib/prop-types'

import { useFormContext } from '..'

import Tooltip from './tooltip'
import FieldError from './error'

import style from './field.styl'

export const getPassThroughProps = (props, excludeProps) => {
  const rest = {}
  for (const property of Object.keys(props)) {
    if (!excludeProps[property]) {
      rest[property] = props[property]
    }
  }
  return rest
}

const isValueEmpty = value => {
  if (value === null || value === undefined) {
    return true
  }

  if (typeof value === 'object') {
    return Object.keys(value) === 0
  }

  if (typeof value === 'string') {
    return value === ''
  }

  return false
}

const extractValue = value => {
  let newValue = value
  if (typeof value === 'object' && value !== null && 'target' in value) {
    const target = value.target
    if ('type' in target && target.type === 'checkbox') {
      newValue = target.checked
    } else if ('value' in target) {
      newValue = target.value
    }
  }

  return newValue
}

const FormField = props => {
  const {
    className,
    component: Component,
    decode,
    description,
    disabled: inputDisabled,
    encode,
    fieldWidth,
    name,
    readOnly,
    required,
    title,
    titleChildren,
    tooltipId,
    warning,
    validate,
    onChange,
    onBlur,
  } = props

  const {
    disabled: formDisabled,
    validateOnBlur,
    setFieldValue,
    setFieldTouched,
  } = useFormContext()

  // Initialize field, which also takes care of registering fields in formik's internal registry.
  const [{ value: encodedValue }, { touched, error = false }] = useField({
    name,
    validate,
  })

  const handleChange = useCallback(
    async (value, enforceValidation = false) => {
      const newValue = encode(extractValue(value))
      let isSyntheticEvent = false

      if (isPlainObject(value)) {
        // Check if the value is react's synthetic event.
        isSyntheticEvent = 'target' in value

        // TODO: Remove `await` and event persist when https://github.com/jaredpalmer/formik/issues/2457
        // is resolved.
        if (typeof value.persist === 'function') {
          value.persist()
        }
      }

      await setFieldValue(name, newValue)

      if (enforceValidation) {
        setFieldTouched(name, true, true)
      }

      onChange(isSyntheticEvent ? value : encode(value))
    },
    [encode, name, onChange, setFieldTouched, setFieldValue],
  )

  const handleBlur = useCallback(
    event => {
      if (validateOnBlur) {
        const value = extractValue(event)
        setFieldTouched(name, !isValueEmpty(value))
      }

      onBlur(event)
    },
    [validateOnBlur, onBlur, setFieldTouched, name],
  )

  const value = decode(encodedValue)
  const disabled = inputDisabled || formDisabled
  const hasTooltip = Boolean(tooltipId)
  const hasTitle = Boolean(title)
  const showError = touched && Boolean(error)
  const showWarning = !Boolean(error) && Boolean(warning)
  const showDescription = !showError && !showWarning && Boolean(description)
  const tooltipIcon = hasTooltip ? <Tooltip id={tooltipId} glossaryTerm={title} /> : null
  const describedBy = showError
    ? `${name}-field-error`
    : showWarning
    ? `${name}-field-warning`
    : showDescription
    ? `${name}-field-description`
    : undefined

  const fieldMessage = showError ? (
    <div className={style.messages}>
      <FieldError content={error} title={title} error id={describedBy} />
    </div>
  ) : showWarning ? (
    <div className={style.messages}>
      <FieldError content={warning} title={title} warning id={describedBy} />
    </div>
  ) : showDescription ? (
    <Message className={style.description} content={description} id={describedBy} />
  ) : null

  const fieldComponentProps = {
    value,
    error: showError,
    warning: showWarning,
    name,
    id: name,
    disabled,
    onChange: handleChange,
    onBlur: handleBlur,
    readOnly,
  }

  const cls = classnames(
    className,
    style.field,
    from(style, {
      error: showError,
      warning: showWarning,
      [`field-width-${fieldWidth}`]: Boolean(fieldWidth),
      required,
      readOnly,
      hasTooltip,
    }),
  )

  return (
    <div className={cls} data-needs-focus={showError} data-test-id="form-field">
      {hasTitle && (
        <div className={style.label}>
          <Message
            component="label"
            content={title}
            className={style.title}
            htmlFor={fieldComponentProps.id}
          />
          {tooltipIcon}
          {titleChildren}
        </div>
      )}
      <div className={style.componentArea}>
        <Component
          aria-invalid={showError}
          aria-describedby={describedBy}
          children={!hasTitle && tooltipIcon}
          {...fieldComponentProps}
          {...getPassThroughProps(props, FormField.propTypes)}
        />
        {fieldMessage}
      </div>
    </div>
  )
}

FormField.propTypes = {
  className: PropTypes.string,
  component: PropTypes.oneOfType([
    PropTypes.func,
    PropTypes.string,
    PropTypes.shape({
      render: PropTypes.func.isRequired,
    }),
  ]).isRequired,
  decode: PropTypes.func,
  description: PropTypes.message,
  disabled: PropTypes.bool,
  encode: PropTypes.func,
  fieldWidth: PropTypes.oneOf([
    'xxs',
    'xs',
    's',
    'm',
    'l',
    'xl',
    'xxl',
    'full',
    'half',
    'third',
    'quarter',
  ]),
  name: PropTypes.string.isRequired,
  onBlur: PropTypes.func,
  onChange: PropTypes.func,
  readOnly: PropTypes.bool,
  required: PropTypes.bool,
  title: PropTypes.message,
  titleChildren: PropTypes.oneOfType([PropTypes.node, PropTypes.arrayOf(PropTypes.node)]),
  tooltipId: PropTypes.string,
  validate: PropTypes.func,
  warning: PropTypes.message,
}

FormField.defaultProps = {
  className: undefined,
  decode: value => value,
  description: '',
  disabled: false,
  encode: value => value,
  fieldWidth: undefined,
  onBlur: () => null,
  onChange: () => null,
  readOnly: false,
  required: false,
  title: undefined,
  titleChildren: null,
  tooltipId: '',
  validate: undefined,
  warning: '',
}

export default FormField
