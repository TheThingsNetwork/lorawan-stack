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

import LogoComponent from '@ttn-lw/components/logo'

import { selectAssetsRootPath, selectApplicationSiteName } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'

const logo = {
  src: `${selectAssetsRootPath()}/tts-logo.svg`,
  alt: `${selectApplicationSiteName()} Logo`,
}
const miniLogo = {
  src: `${selectAssetsRootPath()}/tts-logo-icon.svg`,
  alt: `${selectApplicationSiteName()} Logo`,
}
const whiteLogo = {
  src: `${selectAssetsRootPath()}/logo-tts-horizontal-white.svg`,
  alt: `${selectApplicationSiteName()} Logo white`,
}
const whiteMiniLogo = {
  src: `${selectAssetsRootPath()}/tts-logo-icon-white.svg`,
  alt: `${selectApplicationSiteName()} Logo white`,
}
const Logo = props => {
  const { dark } = props
  return (
    <LogoComponent
      logo={dark ? whiteLogo : logo}
      miniLogo={dark ? whiteMiniLogo : miniLogo}
      {...props}
    />
  )
}

Logo.propTypes = {
  dark: PropTypes.bool,
}

Logo.defaultProps = {
  dark: false,
}

export default Logo
