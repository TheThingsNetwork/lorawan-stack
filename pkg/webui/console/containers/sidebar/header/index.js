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

import React from 'react'

import SideHeader from '@ttn-lw/components/sidebar/side-header'

import { selectApplicationSiteName, selectAssetsRootPath } from '@ttn-lw/lib/selectors/env'

const Header = () => {
  const logoAlt = `${selectApplicationSiteName()} Logo`
  const logoSrc = `${selectAssetsRootPath()}/tts-logo.svg`
  const miniLogoSrc = `${selectAssetsRootPath()}/tts-logo-icon.svg`

  return (
    <SideHeader
      logo={{ src: logoSrc, alt: logoAlt }}
      miniLogo={{ src: miniLogoSrc, alt: logoAlt }}
    />
  )
}

export default Header