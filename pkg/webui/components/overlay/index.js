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
import classnames from 'classnames'
import PropTypes from '../../lib/prop-types'
import sharedMessages from '../../lib/shared-messages'

import Spinner from '../spinner'
import Message from '../../lib/components/message'

import style from './overlay.styl'

const Overlay = ({ className, visible, loading = false, children }) => (
  <div className={classnames(className, style.overlayWrapper)}>
    <div
      className={classnames(style.overlay, {
        [style.overlayVisible]: visible,
      })}
    />
    {visible && loading && (
      <Spinner center>
        <Message content={sharedMessages.fetching} />
      </Spinner>
    )}
    {children}
  </div>
)

Overlay.propTypes = {
  /** A flag specifying whether the overlay is visible or not */
  visible: PropTypes.bool.isRequired,
  /**
   * A flag specifying whether the overlay should displat the loading spinner
   */
  loading: PropTypes.bool,
}

export default Overlay
