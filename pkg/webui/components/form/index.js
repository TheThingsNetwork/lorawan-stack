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
import { Formik } from 'formik'
import bind from 'autobind-decorator'
import scrollIntoView from 'scroll-into-view-if-needed'

import Notification from '../notification'
import PropTypes from '../../lib/prop-types'
import FormContext from './context'
import FormField from './field'
import FormInfoField from './field/info'
import FormSubmit from './submit'

class InnerForm extends React.PureComponent {
  constructor(props) {
    super(props)
    this.notificationRef = React.createRef()
  }

  componentDidUpdate(prevProps) {
    const { formError, isSubmitting, isValid } = this.props
    const { isSubmitting: prevIsSubmitting, formError: prevFormError } = prevProps

    // Scroll form notification into view if needed
    if (formError && !prevFormError) {
      scrollIntoView(this.notificationRef.current, { behavior: 'smooth' })
      this.notificationRef.current.focus({ preventScroll: true })
    }

    // Scroll invalid fields into view if needed and focus them
    if (prevIsSubmitting && !isSubmitting && !isValid) {
      const firstErrorNode = document.querySelectorAll('[data-needs-focus="true"]')[0]
      if (firstErrorNode) {
        scrollIntoView(firstErrorNode, { behavior: 'smooth' })
        firstErrorNode.querySelector('input,textarea').focus({ preventScroll: true })
      }
    }
  }

  render() {
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
        {(formError || formInfo) && (
          <div style={{ outline: 'none' }} ref={this.notificationRef} tabIndex="-1">
            {formError && <Notification error={formError} small />}
            {formInfo && <Notification info={formInfo} small />}
          </div>
        )}
        <FormContext.Provider
          value={{
            ...rest,
            horizontal,
          }}
        >
          {children}
        </FormContext.Provider>
      </form>
    )
  }
}

const formRenderer = ({ children, ...rest }) =>
  function(props) {
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
  render() {
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
Form.InfoField = FormInfoField
Form.Submit = FormSubmit

export default Form
