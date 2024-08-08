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

import AppListSideNavigation from './app-list-side-navigation'
import AppSideNavigation from './app-side-navigation'
import GtwListSideNavigation from './gtw-list-side-navigation'
import GtwSideNavigation from './gtw-side-navigation'
import GeneralSideNavigation from './general-side-navigation'

const SidebarNavigation = () => {
  const { pathname } = useLocation()

  const isApplicationsPath = pathname.startsWith('/applications')
  const isGatewaysPath = pathname.startsWith('/gateways')
  const isSingleAppPath =
    isApplicationsPath &&
    /\/applications\/[a-z0-9]+([-]?[a-z0-9]+)*\/?/i.test(pathname) &&
    !pathname.endsWith('applications/add')
  const isSingleGatewayPath =
    isGatewaysPath &&
    /\/gateways\/[a-z0-9]+([-]?[a-z0-9]+)*\/?/i.test(pathname) &&
    !pathname.endsWith('gateways/add')

  return (
    <>
      {!isApplicationsPath && !isGatewaysPath && <GeneralSideNavigation />}
      {isApplicationsPath && !isSingleAppPath && <AppListSideNavigation />}
      {isSingleAppPath && <AppSideNavigation />}
      {isGatewaysPath && !isSingleGatewayPath && <GtwListSideNavigation />}
      {isSingleGatewayPath && <GtwSideNavigation />}
    </>
  )
}

export default SidebarNavigation
