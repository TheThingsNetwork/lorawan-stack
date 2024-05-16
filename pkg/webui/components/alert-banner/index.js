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

import React, { useEffect, useRef } from 'react'
import classnames from 'classnames'

import Icon, { IconX } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import from from '@ttn-lw/lib/from'

import style from './alert-banner.styl'

const AlertBanner = ({
  className,
  type,
  open,
  title,
  subtitle,
  titleValues,
  subtitleValues,
  handleClose,
}) => {
  const ref = useRef(null)

  useEffect(() => {
    const handleCloseBanner = () => {
      if (window.innerWidth > 768 && open) {
        handleClose()
      }
    }

    const header = document.getElementById('header')
    if (header) {
      const headerHeight = header.offsetHeight
      ref.current.style.top = `${headerHeight}px`
    }

    window.addEventListener('resize', handleCloseBanner)
    return () => {
      window.removeEventListener('resize', handleCloseBanner)
    }
  }, [handleClose, open])

  return (
    <div
      ref={ref}
      className={classnames(
        style.alertBanner,
        className,
        style[type],
        ...from(style, {
          visible: open,
        }),
      )}
    >
      <div className="d-flex al-center j-between mb-cs-xxs">
        <Message
          className={classnames(style.title, 'fw-bold', 'fs-l')}
          content={title}
          values={titleValues}
        />
        <Icon
          className={classnames(style.closeIcon, 'cursor-pointer')}
          icon={IconX}
          onClick={handleClose}
        />
      </div>
      {subtitle && <Message content={subtitle} values={subtitleValues} />}
    </div>
  )
}

AlertBanner.propTypes = {
  className: PropTypes.string,
  handleClose: PropTypes.func.isRequired,
  open: PropTypes.bool.isRequired,
  subtitle: PropTypes.message,
  subtitleValues: PropTypes.object,
  title: PropTypes.message.isRequired,
  titleValues: PropTypes.object,
  type: PropTypes.oneOf(['info', 'success', 'warning', 'error']).isRequired,
}

AlertBanner.defaultProps = {
  className: undefined,
  subtitle: undefined,
  titleValues: undefined,
  subtitleValues: undefined,
}
export default AlertBanner
