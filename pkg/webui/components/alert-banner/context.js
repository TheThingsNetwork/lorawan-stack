// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useContext, useRef, useState } from 'react'

import AlertBanner from '@ttn-lw/components/alert-banner/index'

import PropTypes from '@ttn-lw/lib/prop-types'

const AlertBannerContext = React.createContext()
const { Provider } = AlertBannerContext

const AlertBannerProvider = ({ children }) => {
  const [type, setType] = useState('info')
  const [open, setOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [subtitle, setSubtitle] = useState(undefined)
  const [titleValues, setTitleValues] = useState(undefined)
  const [subtitleValues, setSubtitleValues] = useState(undefined)
  const timeoutRef = useRef(null)
  const showBanner = ({ type, title, duration, subtitle, titleValues, subtitleValues }) => {
    if (window.innerWidth <= 768) {
      clearTimeout(timeoutRef.current)
      setType(type)
      setTitle(title)
      setSubtitle(subtitle)
      setTitleValues(titleValues)
      setSubtitleValues(subtitleValues)
      setOpen(true)
      if (duration) {
        timeoutRef.current = setTimeout(() => {
          setOpen(false)
        }, duration)
      }
    }
  }
  const closeBanner = () => {
    setOpen(false)
    clearTimeout(timeoutRef.current)
  }
  const value = {
    showBanner,
  }

  return (
    <Provider value={value}>
      <AlertBanner
        open={open}
        handleClose={closeBanner}
        type={type}
        title={title}
        titleValues={titleValues}
        subtitle={subtitle}
        subtitleValues={subtitleValues}
      />
      {children}
    </Provider>
  )
}

AlertBannerProvider.propTypes = {
  children: PropTypes.node.isRequired,
}

const useAlertBanner = () => {
  const context = useContext(AlertBannerContext)
  if (!context) {
    throw new Error('useAlertBanner must be used within a AlertBannerProvider')
  }
  return context
}

export { AlertBannerProvider, useAlertBanner }
