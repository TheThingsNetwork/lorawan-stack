// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import DOM from 'react-dom'

import PropTypes from '@ttn-lw/lib/prop-types'

import Modal from '..'

/**
 * PortalledModal is a wrapper around the modal component that renders it into
 * a portal div with the id "modal-container". This div needs to be present
 * for the portal to be functional. This way the modal can be displayed at the
 * top of the DOM hierarchy, regardless of its position in the component
 * hierarchy.
 *
 * @param {object} props - The props of the modal component.
 * @param {boolean} props.visible - Whether the modal is currently visible.
 * @returns {object} - The modal rendered into a portal.
 */
const PortalledModal = ({ visible, ...modalProps }) =>
  DOM.createPortal(visible && <Modal {...modalProps} />, document.getElementById('modal-container'))

PortalledModal.Modal = Modal

PortalledModal.propTypes = {
  ...Modal.propTypes,
  visible: PropTypes.bool,
}

PortalledModal.defaultProps = {
  ...Modal.defaultProps,
  visible: false,
}

export default PortalledModal
