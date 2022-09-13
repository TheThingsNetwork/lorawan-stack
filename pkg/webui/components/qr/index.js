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

import React, { useCallback, useState } from 'react'
import * as jsQR from 'jsqr'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import Video from './input/video'
import RequirePermission from './require-permission'

import style from './qr.styl'

const QR = props => {
  const { onChange } = props
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  const [value, setValue] = useState(null)
  const [capture, setCapture] = useState(false)
  const [error, setError] = useState(false)

  const handleRead = useCallback(
    (media, width, height) => {
      if (!width && !height) {
        return
      }

      canvas.width = width
      canvas.height = height
      ctx.drawImage(media, 0, 0, width, height)

      const { data } = ctx.getImageData(0, 0, width, height)
      const qr = jsQR(data, width, height, {
        // !Important dontInvert fixes a ~50% performance hit.
        inversionAttempts: 'dontInvert',
      })

      if (qr && qr.data && qr.data !== value) {
        setValue({ value: qr.data })
        onChange(qr.data, true)
      }
    },
    [canvas.height, canvas.width, ctx, onChange, value],
  )

  const cls = classnames(style.wrapper, {
    [style.capture]: Boolean(capture),
  })

  return (
    <div className={cls}>
      <RequirePermission
        onRead={handleRead}
        useCapture={capture}
        setCapture={setCapture}
        videoError={error}
      >
        <Video onRead={handleRead} setError={setError} setCapture={setCapture} />
      </RequirePermission>
    </div>
  )
}
QR.propTypes = {
  onChange: PropTypes.func.isRequired,
}

export default QR
