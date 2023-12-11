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

import { useSelector, useDispatch } from 'react-redux'
import React, { useEffect } from 'react'
import {
  Routes,
  Route,
  BrowserRouter,
  ScrollRestoration,
  createBrowserRouter,
  RouterProvider,
  Outlet,
} from 'react-router-dom'
import { Helmet } from 'react-helmet'

import { ToastContainer } from '@ttn-lw/components/toast'

import ErrorView from '@ttn-lw/lib/components/error-view'
import FullViewError from '@ttn-lw/lib/components/full-view-error'

import Header from '@account/containers/header'

import Landing from '@account/views/landing'
import Authorize from '@account/views/authorize'

import {
  selectApplicationSiteName,
  selectApplicationSiteTitle,
  selectPageData,
} from '@ttn-lw/lib/selectors/env'
import { setStatusOnline } from '@ttn-lw/lib/store/actions/status'

import { selectUser } from '@account/store/selectors/user'

import Front from '../front'

const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()
const pageData = selectPageData()

const errorRender = error => <FullViewError error={error} header={<Header />} />
const getScrollRestorationKey = location => {
  // Preserve scroll position only when necessary.
  // E.g. we don't want to scroll to top when changing tabs of a table,
  // but we do want to scroll to top when changing pages.
  const { pathname, search } = location
  const params = new URLSearchParams(search)
  const page = params.get('page')

  return `${pathname}${page ? `?page=${page}` : ''}`
}

const Layout = () => (
  <>
    <ScrollRestoration getKey={getScrollRestorationKey} />
    <ToastContainer />
    <ErrorView errorRender={errorRender}>
      <React.Fragment>
        <Helmet
          titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
          defaultTitle={`${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
        />
        <Outlet />
      </React.Fragment>
    </ErrorView>
  </>
)

const AccountRoot = () => {
  const user = useSelector(selectUser)
  const dispatch = useDispatch()

  useEffect(() => {
    const handleConnectionStatusChange = ({ type }) => {
      dispatch(setStatusOnline(type === 'online'))
    }

    window.addEventListener('online', handleConnectionStatusChange)
    window.addEventListener('offline', handleConnectionStatusChange)

    return () => {
      window.removeEventListener('online', handleConnectionStatusChange)
      window.removeEventListener('offline', handleConnectionStatusChange)
    }
  }, [dispatch])

  if (pageData && pageData.error) {
    return (
      <BrowserRouter history={history} basename="/oauth">
        <FullViewError error={pageData.error} header={<Header />} />
      </BrowserRouter>
    )
  }

  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/authorize/*" Component={Authorize} />
        <Route path="*" Component={Boolean(user) ? Landing : Front} />
      </Route>
    </Routes>
  )
}

const router = createBrowserRouter([{ path: '*', Component: AccountRoot }], {
  basename: '/oauth',
})

const AccountApp = () => <RouterProvider router={router} />

export default AccountApp
