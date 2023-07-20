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
import { ToastContainer as Container, cssTransition } from 'react-toastify'

import PropTypes from '@ttn-lw/lib/prop-types'

import createToast from './toast'

import './react-toastify.styl'
import style from './toast.styl'

const ToastContainer = props => (
  <Container toastClassName={style.toast} bodyClassName={style.body} {...props} />
)

ToastContainer.propTypes = {
  autoClose: PropTypes.oneOfType([PropTypes.bool, PropTypes.number]),
  closeButton: PropTypes.oneOfType([PropTypes.bool, PropTypes.element]),
  closeOnClick: PropTypes.bool,
  hideProgressBar: PropTypes.bool,
  limit: PropTypes.number,
  pauseOnFocusLoss: PropTypes.bool,
  pauseOnHover: PropTypes.bool,
  position: PropTypes.oneOf([
    'bottom-right',
    'bottom-left',
    'top-right',
    'top-left',
    'top-center',
    'bottom-center',
  ]),
  transition: PropTypes.func,
}

ToastContainer.defaultProps = {
  autoClose: undefined,
  position: 'bottom-right',
  closeButton: false,
  hideProgressBar: true,
  pauseOnHover: true,
  closeOnClick: true,
  pauseOnFocusLoss: true,
  limit: 2,
  transition: cssTransition({ enter: style.slideInRight, exit: style.slideOutRight }),
}

const toast = createToast()

export { toast as default, ToastContainer }
