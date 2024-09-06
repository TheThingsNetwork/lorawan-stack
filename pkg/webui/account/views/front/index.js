// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route, Navigate, useLocation } from 'react-router-dom'

import authRoutes from '@account/constants/auth-routes'

import Login from '@account/views/login'
import TokenLogin from '@account/views/token-login'
import CreateAccount from '@account/views/create-account'
import ForgotPassword from '@account/views/forgot-password'
import UpdatePassword from '@account/views/update-password'
import FrontNotFound from '@account/views/front-not-found'
import Validate from '@account/views/validate'

import { selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'

const FrontView = () => {
  const location = useLocation()

  return (
    <Routes>
      <Route path="/login" Component={Login} />
      <Route path="/token-login" Component={TokenLogin} />
      <Route path="/register" Component={CreateAccount} />
      <Route path="/forgot-password" Component={ForgotPassword} />
      <Route path="/update-password" Component={UpdatePassword} />
      <Route path="/validate" Component={Validate} />
      <Route index element={<Navigate to="/login" />} />
      {authRoutes.map(({ path }) => (
        <Route
          path={path}
          key={path}
          element={<Navigate to={`/login?n=${selectApplicationRootPath()}${location.pathname}`} />}
        />
      ))}
      <Route Component={FrontNotFound} />
    </Routes>
  )
}

export default FrontView
