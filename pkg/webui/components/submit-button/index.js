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

import Button from '@ttn-lw/components/button'

import PropTypes from '@ttn-lw/lib/prop-types'

const SubmitButton = ({ disabled, icon, isSubmitting, isValidating, message, ...rest }) => {
  const buttonLoading = isSubmitting || isValidating

  return (
    <Button
      primary
      {...rest}
      type="submit"
      icon={icon}
      message={message}
      disabled={disabled}
      busy={buttonLoading}
    />
  )
}

SubmitButton.propTypes = {
  disabled: PropTypes.bool,
  icon: PropTypes.icon,
  isSubmitting: PropTypes.bool.isRequired,
  isValidating: PropTypes.bool.isRequired,
  message: PropTypes.message,
}

SubmitButton.defaultProps = {
  disabled: false,
  icon: undefined,
  message: undefined,
}

export default SubmitButton
