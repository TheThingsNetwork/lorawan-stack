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
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import style from '../../qr.styl'

const m = defineMessages({
  fetchingCamera: 'Waiting for camera…',
  switchCamera: 'Switch camera',
})

const Video = props => {
  const { onRead, setError, setCapture } = props
  const videoRef = createRef()
  const [stream, setStream] = useState(undefined)
  const [devices, setDevices] = useState([])
  const [cameras, setCameras] = useState([])
  const [videoMode, setVideoMode] = useState({})
  const isMobile = window.innerWidth <= 768

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
      setCameras(cameras)
      let rearCamera = cameras.find(device => device.label.toLowerCase().includes('back'))
      if (!rearCamera) {
        rearCamera = cameras[1] ?? cameras[0]
      }
      const videoMode =
        cameras.length > 1
          ? ua.indexOf('safari') !== -1 && ua.indexOf('chrome') === -1
            ? { facingMode: { exact: 'environment' } }
            : { deviceId: rearCamera.deviceId }
          : { facingMode: 'environment' }

      setVideoMode(videoMode)
      try {
        const userStream = await navigator.mediaDevices.getUserMedia({
          video: videoMode ? { ...videoMode } : { facingMode: 'environment' },
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

  const switchStream = useCallback(async () => {
    if (videoMode.facingMode === 'environment') {
      const userStream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'user' },
      })
      setStream(userStream)
    } else if (videoMode.facingMode === 'user') {
      const userStream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' },
      })
      setStream(userStream)
    } else if ('deviceId' in videoMode) {
      let indexOfCurrentDevice = cameras.findIndex(camera => camera.deviceId === videoMode.deviceId)
      // The first item will be taken from the beginning of the array after the last item.
      const nextIndex = ++indexOfCurrentDevice % cameras.length
      const device = cameras[nextIndex]
      setVideoMode({ deviceId: device.deviceId })
      const userStream = await navigator.mediaDevices.getUserMedia({
        video: { deviceId: device.deviceId },
      })
      setStream(userStream)
    }
  }, [cameras, videoMode])

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

  return (
    <>
      {cameras.length > 1 && isMobile && (
        <Button icon="switch_camera" message={m.switchCamera} onClick={switchStream} />
      )}
      {devices.length && stream ? (
        <video
          autoPlay
          playsInline
          ref={videoRef}
          className={style.video}
          data-test-id="webcam-feed"
        />
      ) : (
        <Spinner center>
          <Message className={style.msg} content={m.fetchingCamera} />
        </Spinner>
      )}
    </>
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
