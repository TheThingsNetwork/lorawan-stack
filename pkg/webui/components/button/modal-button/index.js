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
import bind from 'autobind-decorator'

import Button from '../'
import PortalledModal from '../../../components/modal/portalled'

import PropTypes from '../../../lib/prop-types'

/**
 * ModalButton is a button which needs a modal confirmation to complete the
 * action. It can be used as an easy way to get the users explicit confirmation
 * before doing an action, e.g. deleting a resource.
 */
@bind
class ModalButton extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      modalVisible: false,
    }
  }
  handleClick() {
    const { modalData } = this.props

    if (!modalData) {
      // No modal data likely means a faulty implementation, so since it's
      // likely best to not do anything in this case
      return
    }

    this.setState({ modalVisible: true })
  }

  handleComplete(confirmed) {
    const { onApprove, onCancel } = this.props

    if (confirmed) {
      onApprove()
    } else {
      onCancel()
    }
    this.setState({ modalVisible: false })
  }

  render() {
    const { modalData, message, onApprove, onCancel, ...rest } = this.props

    const modalComposedData = {
      approval: true,
      danger: true,
      buttonMessage: message,
      title: message,
      onComplete: this.handleComplete,
      ...modalData,
    }

    return (
      <React.Fragment>
        <PortalledModal visible={this.state.modalVisible} modal={modalComposedData} />
        <Button onClick={this.handleClick} message={message} {...rest} />
      </React.Fragment>
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
