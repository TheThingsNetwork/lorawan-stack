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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
  Route,
  Routes,
  Outlet,
  createBrowserRouter,
  RouterProvider,
  ScrollRestoration,
} from 'react-router-dom'
import classnames from 'classnames'

import { ToastContainer } from '@ttn-lw/components/toast'
import Breadcrumbs from '@ttn-lw/components/breadcrumbs'
import { AlertBannerProvider } from '@ttn-lw/components/alert-banner/context'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import ErrorView from '@ttn-lw/lib/components/error-view'
import WithAuth from '@ttn-lw/lib/components/with-auth'
import FullViewError, { FullViewErrorInner } from '@ttn-lw/lib/components/full-view-error'

import Header from '@console/containers/header'
import LogBackInModal from '@console/containers/log-back-in-modal'
import Sidebar from '@console/containers/sidebar'

import OverviewRoutes from '@console/views/overview'
import Applications from '@console/views/applications'
import Gateways from '@console/views/gateways'

import { setStatusOnline } from '@ttn-lw/lib/store/actions/status'
import { selectStatusStore } from '@ttn-lw/lib/store/selectors/status'
import {
  selectApplicationSiteName,
  selectApplicationSiteTitle,
  selectPageData,
} from '@ttn-lw/lib/selectors/env'

import {
  selectUser,
  selectUserFetching,
  selectUserError,
  selectUserRights,
  selectUserIsAdmin,
} from '@console/store/selectors/logout'

import style from './app.styl'

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

const Layout = () => {
  const user = useSelector(selectUser)
  const fetching = useSelector(selectUserFetching)
  const error = useSelector(selectUserError)
  const rights = useSelector(selectUserRights)
  const isAdmin = useSelector(selectUserIsAdmin)
  const siteTitle = selectApplicationSiteTitle()
  const siteName = selectApplicationSiteName()

  // For the mobile side menu drawer functionality.
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const node = useRef()

  const openDrawer = useCallback(() => {
    setIsDrawerOpen(true)
    document.body.classList.add(style.scrollLock)
  }, [])

  const closeDrawer = useCallback(() => {
    setIsDrawerOpen(false)
    document.body.classList.remove(style.scrollLock)
  }, [])

  useEffect(() => {
    const onClickOutside = e => {
      if (isDrawerOpen && node.current && !node.current.contains(e.target)) {
        closeDrawer()
      }
    }

    if (isDrawerOpen) {
      document.addEventListener('mousedown', onClickOutside)
      return () => document.removeEventListener('mousedown', onClickOutside)
    }
  }, [isDrawerOpen, closeDrawer])

  // Pass this function to the header prop `onMenuClick`.
  const onDrawerExpandClick = useCallback(() => {
    if (!isDrawerOpen) {
      openDrawer()
    } else {
      closeDrawer()
    }
  }, [isDrawerOpen, openDrawer, closeDrawer])
  // End of mobile side menu drawer functionality

  return (
    <>
      <ScrollRestoration getKey={getScrollRestorationKey} />
      <ErrorView errorRender={errorRender}>
        <div className={style.container}>
          <IntlHelmet
            titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
            defaultTitle={siteName}
          />
          <div id="modal-container" />
          <div id="dropdown-container" className="pos-absolute-container" />
          <div className="d-flex">
            <AlertBannerProvider>
              <Sidebar isDrawerOpen={isDrawerOpen} onDrawerCloseClick={closeDrawer} />
              <div className="w-full h-vh d-flex direction-column">
                <Header onMenuClick={onDrawerExpandClick} />
                <main
                  className={classnames(style.main, 'd-flex', 'flex-column', 'h-full', 'flex-grow')}
                >
                  <WithAuth
                    user={user}
                    fetching={fetching}
                    error={error}
                    errorComponent={FullViewErrorInner}
                    rights={rights}
                    isAdmin={isAdmin}
                  >
                    <div className={style.content}>
                      <div className={style.stage} id="stage">
                        <Outlet />
                      </div>
                    </div>
                  </WithAuth>
                </main>
              </div>
            </AlertBannerProvider>
          </div>
        </div>
      </ErrorView>
    </>
  )
}
const ConsoleRoot = () => {
  const status = useSelector(selectStatusStore)
  const pageData = selectPageData()
  const dispatch = useDispatch()

  const handleConnectionStatusChange = useCallback(
    ({ type }) => {
      dispatch(setStatusOnline(type === 'online'))
    },
    [dispatch],
  )

  useEffect(() => {
    window.addEventListener('online', handleConnectionStatusChange)
    window.addEventListener('offline', handleConnectionStatusChange)
    return () => {
      window.removeEventListener('online', handleConnectionStatusChange)
      window.removeEventListener('offline', handleConnectionStatusChange)
    }
  }, [handleConnectionStatusChange])

  if (pageData && pageData.error) {
    return <FullViewError error={pageData.error} header={<Header />} />
  }

  return (
    <React.Fragment>
      {status.isLoginRequired && <LogBackInModal />}
      <ToastContainer />
      <Breadcrumbs />
      <Routes>
        <Route element={<Layout />}>
          <Route path="/*" Component={OverviewRoutes} />
          <Route path="applications/*" Component={Applications} />
          <Route path="gateways/*" Component={Gateways} />
          <Route path="*" Component={GenericNotFound} />
        </Route>
      </Routes>
    </React.Fragment>
  )
}

const router = createBrowserRouter([{ path: '*', Component: ConsoleRoot }], {
  basename: '/console',
})

const ConsoleApp = () => <RouterProvider router={router} />

export default ConsoleApp
