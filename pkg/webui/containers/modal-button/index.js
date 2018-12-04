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
import { connect } from 'react-redux'
import bind from 'autobind-decorator'

import Button from '../../components/button'

import { setModal } from '../../actions/modal'
import PropTypes from '../../lib/prop-types'

/**
 * ModalButton is a button which needs a modal confirmation to complete the
 * action. It can be used as an easy way to get the users explicit confirmation
 * before doing an action, e.g. deleting a resource.
 */
@connect()
@bind
class ModalButton extends React.Component {
  handleClick (e) {
    const { dispatch, modalData, message, onApprove, onCancel } = this.props

    if (!modalData) {
      // No modal data likely means a faulty implementation, so since it's
      // likely best to not do anything in this case
      return
    }

    dispatch(setModal({
      approval: true,
      danger: true,
      buttonMessage: message,
      title: message,
      onComplete (confirmed) {
        if (confirmed) {
          onApprove(e)
        } else {
          onCancel(e)
        }
      },
      ...modalData,
    }))
  }

  render () {
    const { dispatch, modalData, onApprove, onCancel, ...rest } = this.props

    return (
      <Button onClick={this.handleClick} {...rest} />
    )
  }
}

ModalButton.defaultProps = {
  onApprove: () => null,
  onCancel: () => null,
}

ModalButton.propTypes = {
  onApprove: PropTypes.func,
  onCancel: PropTypes.func,
  modalData: PropTypes.object.isRequired,
}

export default ModalButton
