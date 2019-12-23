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
import { ToastContainer as Container, Slide } from 'react-toastify'
import 'react-toastify/dist/ReactToastify.css'

import PropTypes from '../../lib/prop-types'
import createToast from './toast'

import style from './toast.styl'

class ToastContainer extends React.Component {
  static propTypes = {
    autoClose: PropTypes.oneOfType([PropTypes.bool, PropTypes.number]),
    closeButton: PropTypes.oneOfType([PropTypes.bool, PropTypes.element]),
    closeOnClick: PropTypes.bool,
    hideProgressBar: PropTypes.bool,
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

  static defaultProps = {
    position: 'bottom-right',
    autoClose: 4000,
    closeButton: false,
    hideProgressBar: true,
    pauseOnHover: true,
    closeOnClick: true,
    pauseOnFocusLoss: true,
    transition: Slide,
  }

  render() {
    return <Container toastClassName={style.toast} bodyClassName={style.body} {...this.props} />
  }
}

const toast = createToast()

export { toast as default, ToastContainer }
