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

import PropTypes from '../../lib/prop-types'
import Button from '../button'

class SubmitButton extends React.PureComponent {
  render () {
    const {
      message,
      icon,
      isSubmitting,
      isValidating,
      disabled,
      dirty,
    } = this.props

    const buttonDisabled = disabled || isSubmitting || !dirty
    const buttonLoading = isSubmitting || isValidating

    return (
      <Button
        type="submit"
        icon={icon}
        message={message}
        disabled={buttonDisabled}
        busy={buttonLoading}
      />
    )
  }
}

SubmitButton.propTypes = {
  message: PropTypes.message.isRequired,
  isSubmitting: PropTypes.bool.isRequired,
  isValidating: PropTypes.bool.isRequired,
  dirty: PropTypes.bool.isRequired,
  disabled: PropTypes.bool,
  icon: PropTypes.string,
}

SubmitButton.defaultProps = {
  disabled: false,
}

export default SubmitButton
