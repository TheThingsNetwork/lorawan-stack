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

import React, { useContext } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'

import FormContext from '../context'

const FormSubmit = props => {
  const { component: Component, disabled, ...rest } = props
  const formContext = useContext(FormContext)

  const submitProps = {
    isValid: context.isValid,
    isSubmitting: formContext.isSubmitting,
    isValidating: formContext.isValidating,
    submitCount: formContext.submitCount,
    dirty: formContext.dirty,
    validateForm: formContext.validateForm,
    validateField: formContext.validateField,
    disabled: formContext.disabled || disabled,
  }

  return <Component {...rest} {...submitProps} />
}

FormSubmit.propTypes = {
  component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
  disabled: PropTypes.bool,
}

FormSubmit.defaultProps = {
  component: 'button',
  disabled: false,
}

export default FormSubmit
