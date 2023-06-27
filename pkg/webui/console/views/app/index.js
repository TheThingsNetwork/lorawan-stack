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

import React, { useCallback, useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Route, Routes, BrowserRouter } from 'react-router-dom'
import classnames from 'classnames'

import { ToastContainer } from '@ttn-lw/components/toast'
import sidebarStyle from '@ttn-lw/components/navigation/side/side.styl'

import Footer from '@ttn-lw/containers/footer'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import ErrorView from '@ttn-lw/lib/components/error-view'
import ScrollToTop from '@ttn-lw/lib/components/scroll-to-top'
import WithAuth from '@ttn-lw/lib/components/with-auth'
import FullViewError, { FullViewErrorInner } from '@ttn-lw/lib/components/full-view-error'

import Header from '@console/containers/header'
import LogBackInModal from '@console/containers/log-back-in-modal'

import Overview from '@console/views/overview'
import Applications from '@console/views/applications'
import Gateways from '@console/views/gateways'
import Organizations from '@console/views/organizations'
import AdminPanel from '@console/views/admin-panel'
import User from '@console/views/user'

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

const ConsoleApp = () => {
  const user = useSelector(selectUser)
  const fetching = useSelector(selectUserFetching)
  const error = useSelector(selectUserError)
  const rights = useSelector(selectUserRights)
  const isAdmin = useSelector(selectUserIsAdmin)
  const status = useSelector(selectStatusStore)
  const siteTitle = selectApplicationSiteTitle()
  const pageData = selectPageData()
  const siteName = selectApplicationSiteName()
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
    return (
      <BrowserRouter history={history} basename="/console">
        <FullViewError error={pageData.error} header={<Header />} />
      </BrowserRouter>
    )
  }

  return (
    <React.Fragment>
      {status.isLoginRequired && <LogBackInModal />}
      <ToastContainer />
      <BrowserRouter history={history} basename="/console">
        <ScrollToTop />
        <ErrorView errorRender={errorRender}>
          <div className={style.app}>
            <IntlHelmet
              titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
              defaultTitle={siteName}
            />
            <div id="modal-container" />
            <Header />
            <main className={style.main}>
              <WithAuth
                user={user}
                fetching={fetching}
                error={error}
                errorComponent={FullViewErrorInner}
                rights={rights}
                isAdmin={isAdmin}
              >
                <div className={classnames('breadcrumbs', style.mobileBreadcrumbs)} />
                <div id="sidebar" className={sidebarStyle.container} />
                <div className={style.content}>
                  <div className={classnames('breadcrumbs', style.desktopBreadcrumbs)} />
                  <div className={style.stage} id="stage">
                    <Routes>
                      <Route index Component={Overview} />
                      <Route path="/applications/*" Component={Applications} />
                      <Route path="/gateways/*" Component={Gateways} />
                      <Route path="/organizations/*" Component={Organizations} />
                      <Route path="/admin-panel/*" Component={AdminPanel} />
                      <Route path="/user/*" Component={User} />
                      <Route path="*" Component={GenericNotFound} />
                    </Routes>
                  </div>
                </div>
              </WithAuth>
            </main>
            <Footer className={style.footer} />
          </div>
        </ErrorView>
      </BrowserRouter>
    </React.Fragment>
  )
}

export default ConsoleApp
