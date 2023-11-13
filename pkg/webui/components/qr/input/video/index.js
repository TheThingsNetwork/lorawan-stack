// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

const Camera = props => {
  const { onRead, setError, setCapture } = props
  const [cameras, setCameras] = useState([])
  const [deviceId, setDeviceId] = useState(undefined)

  const [hasFrontCamera, setHasFrontCamera] = useState(false)
  const [hasBackCamera, setHasBackCamera] = useState(false)

  const handleSwitchCamera = useCallback(() => {
    // The majority of mobile device will have labeled front and back cameras.
    const currentCamera = cameras.find(device => device.deviceId === deviceId)

    if (hasFrontCamera && hasBackCamera) {
      if (currentCamera.label.toLowerCase().includes('back')) {
        const frontCamera = cameras.find(device => device.label.toLowerCase().includes('front'))
        setDeviceId(frontCamera.deviceId)
      }

      if (currentCamera.label.toLowerCase().includes('front')) {
        const backCamera = cameras.find(device => device.label.toLowerCase().includes('back'))
        setDeviceId(backCamera.deviceId)
      }
    }
  }, [cameras, deviceId, hasBackCamera, hasFrontCamera])

  const handleCameraCycle = useCallback(() => {
    // If there are more than one camera then cycle through them, only if there are no camera clearly labeled and back and front
    if (cameras.length > 1) {
      const currentIndex = cameras.findIndex(device => device.deviceId === deviceId)

      const nextDevice = currentIndex !== -1 ? cameras[(currentIndex + 1) % cameras.length] : null
      setDeviceId(nextDevice.deviceId)
    }
  }, [cameras, deviceId])

  const getDevices = useCallback(async () => {
    // Depending on your device you may have access to this list on initial load
    // If you do not have access to this list then you will need to request a stream to get the list
    const enumerateDevices = await navigator.mediaDevices.enumerateDevices()
    const videoInputs = enumerateDevices.filter(device => device.kind === 'videoinput')

    if (videoInputs.length !== cameras.length) {
      setCameras(videoInputs)
    }
  }, [cameras])

  const setDeviceIdFromStream = useCallback(userStream => {
    const videoTracks = userStream.getVideoTracks()

    if (videoTracks.length > 0) {
      const { deviceId } = videoTracks[0].getSettings()
      setDeviceId(deviceId)
    }
  }, [])

  useEffect(() => {
    setHasFrontCamera(cameras.some(device => device.label.toLowerCase().includes('front')))
    setHasBackCamera(cameras.some(device => device.label.toLowerCase().includes('back')))
  }, [cameras])

  return (
    <>
      {hasFrontCamera && hasBackCamera ? (
        <Button icon="switch_camera" message={m.switchCamera} onClick={handleSwitchCamera} />
      ) : (
        cameras.length > 1 && (
          <Button icon="switch_camera" message={m.switchCamera} onClick={handleCameraCycle} />
        )
      )}

      <Stream
        deviceId={deviceId}
        getDevices={getDevices}
        setDeviceIdFromStream={setDeviceIdFromStream}
        onRead={onRead}
        setError={setError}
        setCapture={setCapture}
      />
    </>
  )
}

Camera.propTypes = {
  onRead: PropTypes.func.isRequired,
  setCapture: PropTypes.func.isRequired,
  setError: PropTypes.func.isRequired,
}

const Stream = props => {
  const { deviceId, getDevices, setDeviceIdFromStream, onRead, setCapture, setError } = props
  const [stream, setStream] = useState(undefined)

  useEffect(() => {
    const getStream = async () => {
      // Initially request the stream with the default camera for facing mode environment
      // if device id is set then create stream with that device id and display video
      try {
        if (!deviceId) {
          const userStream = await navigator.mediaDevices.getUserMedia({
            video: { facingMode: 'environment' },
          })
          // After requesting the stream, get the devices again as we now have permission to see all devices
          getDevices()
          // On initial request set the device id from the stream, this should rerender this component
          setDeviceIdFromStream(userStream)
        } else {
          const userStream = await navigator.mediaDevices.getUserMedia({
            video: { deviceId },
          })
          // Only set the stream if the device id is set
          setStream(userStream)
        }
      } catch (error) {
        if (error instanceof DOMException && error.name === 'NotAllowedError') {
          setCapture(false)
          setError(true)
        } else {
          throw error
        }
      }
    }

    getStream()
  }, [deviceId, getDevices, setCapture, setDeviceIdFromStream, setError])

  return <Video stream={stream} onRead={onRead} />
}

Stream.propTypes = {
  deviceId: PropTypes.string,
  getDevices: PropTypes.func.isRequired,
  onRead: PropTypes.func.isRequired,
  setCapture: PropTypes.func.isRequired,
  setDeviceIdFromStream: PropTypes.func.isRequired,
  setError: PropTypes.func.isRequired,
}

Stream.defaultProps = {
  deviceId: undefined,
}

const Video = props => {
  const { stream, onRead } = props
  const videoRef = createRef()

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
    if (stream) {
      const video = videoRef.current
      video.srcObject = stream
      handleVideoFrame(video)
    }

    return () => {
      // On android devices if you do not stop the tracks then the camera will error
      if (stream) {
        stream.getTracks().forEach(track => {
          track.stop()
        })
      }
    }
  }, [handleVideoFrame, stream, videoRef])

  return (
    <>
      {stream ? (
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
  stream: PropTypes.shape({
    active: PropTypes.bool,
    getTracks: PropTypes.func,
  }),
}

Video.defaultProps = {
  stream: undefined,
}

export default Camera
