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
import { useLocation } from 'react-router-dom'
import classNames from 'classnames'

import SearchButton from '@console/components/search-button'

import Switcher from './switcher'
import SideBarNavigation from './navigation'
import SideBarContext from './context'
import SideHeader from './header'
import SideFooter from './footer'

import style from './side-bar.styl'

const SideBar = () => {
  const { pathname } = useLocation()
  const [layer, setLayer] = React.useState(pathname ?? '/')

  return (
    <div
      className={classNames(
        style.sidebar,
        'd-flex pos-relative align-center direction-column gap-cs-s p-cs-s bg-tts-primary-050',
      )}
    >
      <SideHeader />
      <SideBarContext.Provider value={{ layer, setLayer }}>
        <Switcher />
        <SearchButton />
        <SideBarNavigation />
      </SideBarContext.Provider>
      <SideFooter />
    </div>
  )
}

export default SideBar
