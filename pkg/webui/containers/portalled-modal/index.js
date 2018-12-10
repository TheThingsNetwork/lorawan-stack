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

import React, { Component } from 'react'
import DOM from 'react-dom'
import bind from 'autobind-decorator'

import Modal from '../../components/modal'

@bind
class PortalledModal extends Component {
  handleComplete (result) {
    const { modal } = this.props
    if (modal && modal.onComplete) {
      modal.onComplete(result)
    }
  }

  render () {
    const { modal, visible, ...rest } = this.props

    if (!modal) {
      return null
    }

    const { onComplete, ...modalRest } = modal
    const props = { ...rest, ...modalRest }

    return DOM.createPortal(
      visible && <Modal onComplete={this.handleComplete} {...props} />,
      document.getElementById('modal-container')
    )
  }
}

export default PortalledModal
