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
import { Formik, getIn } from 'formik'
import bind from 'autobind-decorator'

import scrollIntoView from 'scroll-into-view-if-needed'
import Notification from '../notification'
import PropTypes from '../../lib/prop-types'
import FormContext from './context'
import FormField from './field'
import FormSubmit from './submit'

@bind
class InnerForm extends React.PureComponent {

  formError = React.createRef()
  fields = {}
  firstField = undefined

  _getInvalidField () {
    const { errors } = this.props

    let field
    for (const fieldName of Object.keys(this.fields)) {
      const err = getIn(errors, fieldName)
      if (err) {
        field = this.fields[fieldName]
        break
      }
    }

    return field
  }

  _scroll (node, delay = 200) {
    if (node) {
      scrollIntoView(node, {
        scrollMode: 'if-needed',
        behavior: 'smooth',
        duration: delay,
      })
    }
  }

  _focus (node, delay = 200) {
    setTimeout(() => node.focus(), delay)
  }

  componentDidUpdate (prevProps) {
    const { formError, isSubmitting, isValid } = this.props

    if (formError && formError !== prevProps.formError) {
      this._scroll(this.formError.current)
      this._focus(this.firstField)
    } else if (prevProps.isSubmitting && !isSubmitting && !isValid) {
      const invalidField = this._getInvalidField()
      if (invalidField && invalidField.focus) {
        this._focus(invalidField, 100)
      } else if (this.firstField && this.firstField.focus) {
        this._focus(this.firstField, 100)
      }
    }
  }

  registerField (name, field) {
    const { registerField } = this.props

    this.fields[name] = field
    registerField(name, field)

    if (!this.firstField && !field.props.disabled) {
      this.firstField = field
    }
  }

  unregisterField (name) {
    const { unregisterField } = this.props

    const field = this.fields[name]

    if (field === this.firstField) {
      this.firstField = undefined
    }

    delete this.fields[name]
    unregisterField(name)
  }

  render () {
    const {
      className,
      children,
      formError,
      formInfo,
      horizontal,
      handleSubmit,
      ...rest
    } = this.props

    return (
      <form className={className} onSubmit={handleSubmit}>
        {formError && <Notification error={formError} small ref={this.formError} />}
        {formInfo && <Notification info={formInfo} small /> }
        <FormContext.Provider value={{
          ...rest,
          horizontal,
          registerField: this.registerField,
          unregisterField: this.unregisterField,
        }}
        >
          {children}
        </FormContext.Provider>
      </form>
    )
  }
}

const formRenderer = ({ children, ...rest }) => function (props) {
  const { className, horizontal, error, info, disabled } = rest
  const { handleSubmit, ...restFormikProps } = props

  return (
    <InnerForm
      className={className}
      horizontal={horizontal}
      formError={error}
      formInfo={info}
      handleSubmit={handleSubmit}
      disabled={disabled}
      {...restFormikProps}
    >
      {children}
    </InnerForm>
  )
}

@bind
class Form extends React.PureComponent {
  render () {
    const {
      onSubmit,
      onReset,
      initialValues,
      isInitialValid,
      validateOnBlur,
      validateOnChange,
      validationSchema,
      formikRef,
      ...rest
    } = this.props
    return (
      <Formik
        ref={formikRef}
        render={formRenderer(rest)}
        onSubmit={onSubmit}
        onReset={onReset}
        initialValues={initialValues}
        isInitialValid={isInitialValid}
        validateOnBlur={validateOnBlur}
        validateOnChange={validateOnChange}
        validationSchema={validationSchema}
      />
    )
  }
}

Form.propTypes = {
  // formik props
  onSubmit: PropTypes.func.isRequired,
  onReset: PropTypes.func,
  initialValues: PropTypes.object.isRequired,
  validateOnBlur: PropTypes.bool,
  validateOnChange: PropTypes.bool,
  validationSchema: PropTypes.object,
  isInitialValid: PropTypes.bool,
  formikRef: PropTypes.shape({ current: PropTypes.instanceOf(Formik) }),
  // custom props
  horizontal: PropTypes.bool,
  className: PropTypes.string,
  error: PropTypes.error,
  disabled: PropTypes.bool,
}

Form.defaultProps = {
  className: null,
  submitEnabledWhenInvalid: false,
  validateOnBlur: true,
  validateOnChange: false,
  validationSchema: {},
  isInitialValid: false,
  onReset: () => null,
  error: '',
  horizontal: true,
  disabled: false,
}

Form.Field = FormField
Form.Submit = FormSubmit

export default Form
