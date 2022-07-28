// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import PropTypes from 'prop-types'
import { defineMessages } from 'react-intl'

import FileInput from '../../../file-input'
import style from '../../qr.styl'

const m = defineMessages({
  uploadImage: 'Upload a Photo',
  qrCodeNotFound: 'QR code not found',
})

const Capture = props => {
  const { onRead } = props

  const handleChange = useCallback(
    data => {
      const image = new Image()
      image.src = data
      image.onload = () => {
        onRead(image, image.width, image.height)
      }
    },
    [onRead],
  )

  const handleDataTransform = useCallback(content => content, [])

  return (
    <div className={style.captureWrapper}>
      <FileInput
        name="captureFileInput"
        id="captureFileInput"
        onChange={handleChange}
        message={m.uploadImage}
        dataTransform={handleDataTransform}
        providedMessage={m.uploadImage}
        image
        warningSize={0}
        largeFileWarningMessage={m.qrCodeNotFound}
        center
      />
    </div>
  )
}

Capture.propTypes = {
  onRead: PropTypes.func.isRequired,
}

export default Capture
