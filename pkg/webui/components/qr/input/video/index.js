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

import React, { useCallback, createRef, useEffect, useState } from 'react'
import PropTypes from 'prop-types'
import { defineMessages } from 'react-intl'

import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import style from '../../qr.styl'

const m = defineMessages({
  fetchingCamera: 'Waiting for camera…',
})

const Video = props => {
  const { onRead, setError, setCapture } = props
  const videoRef = createRef()
  const [stream, setStream] = useState(undefined)
  const [devices, setDevices] = useState([])

  const getDevices = useCallback(async () => {
    if (!devices.length) {
      const enumerateDevices = await navigator.mediaDevices.enumerateDevices()
      setDevices(enumerateDevices)
    }
  }, [devices.length])

  const getStream = useCallback(async () => {
    if (devices.length && !stream) {
      const ua = navigator.userAgent.toLowerCase()
      const cameras = devices.filter(device => device.kind === 'videoinput')
      const videoMode =
        cameras.length > 1
          ? ua.indexOf('safari') !== -1 && ua.indexOf('chrome') === -1
            ? { facingMode: { exact: 'environment' } }
            : { deviceId: cameras[1].deviceId }
          : { facingMode: 'environment' }

      try {
        const userStream = await navigator.mediaDevices.getUserMedia({
          video: { ...videoMode },
        })
        setStream(userStream)
      } catch (error) {
        if (error instanceof DOMException && error.name === 'NotAllowedError') {
          setCapture(false)
          setError(true)
        } else {
          throw error
        }
      }
    }
  }, [devices, setCapture, setError, stream])

  useEffect(() => {
    getDevices()
    getStream()
    return () => {
      if (stream) {
        stream.getTracks().map(t => t.stop())
      }
    }
  }, [devices, getDevices, getStream, stream])

  const handleVideoFrame = useCallback(
    video => {
      const { active } = stream
      const { videoWidth: width, videoHeight: height } = video

      onRead(video, width, height)

      if (active) {
        requestAnimationFrame(() => {
          handleVideoFrame(video)
        })
      }
    },
    [onRead, stream],
  )

  useEffect(() => {
    if (devices.length && stream) {
      const video = videoRef.current
      video.srcObject = stream
      handleVideoFrame(video)
    }
  }, [devices, handleVideoFrame, stream, videoRef])

  return devices.length && stream ? (
    <video autoPlay playsInline ref={videoRef} className={style.video} data-test-id="webcam-feed" />
  ) : (
    <Spinner center>
      <Message className={style.msg} content={m.fetchingCamera} />
    </Spinner>
  )
}

Video.propTypes = {
  onRead: PropTypes.func.isRequired,
  setCapture: PropTypes.func.isRequired,
  setError: PropTypes.func.isRequired,
  stream: PropTypes.shape({
    active: PropTypes.bool,
  }),
}

Video.defaultProps = {
  stream: undefined,
}

export default Video
