// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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


/* global require */
import React from 'react'
import { configure, addDecorator } from '@storybook/react'
import { Provider } from 'react-redux'
import { ConnectedRouter } from 'react-router-redux'
import createHistory from 'history/createMemoryHistory'

import '../../pkg/webui/styles/main.styl'
import store from '../../pkg/webui/store'

import Center from './center'

const req = require.context('../../pkg/webui/', true, /story\.js$/)
const load = () => req.keys().forEach(req)

const history = createHistory()

addDecorator(function (story) {
  return (
    <Provider store={store}>
      <ConnectedRouter history={history}>
        <Center>
          {story()}
        </Center>
      </ConnectedRouter>
    </Provider>
  )
})

configure(load, module)
