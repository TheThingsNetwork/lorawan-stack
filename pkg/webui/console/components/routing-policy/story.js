// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { storiesOf } from '@storybook/react'

import RoutingPolicy from '.'

storiesOf('Routing Policy/Sheet', module)
  .add('Empty', () => (
    <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
      <RoutingPolicy.Sheet
        policy={{
          uplink: {},
          downlink: {},
        }}
      />
    </div>
  ))
  .add('Full', () => (
    <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
      <RoutingPolicy.Sheet
        policy={{
          uplink: {
            join_request: true,
            mac_data: true,
            application_data: true,
            signal_quality: true,
            localization: true,
          },
          downlink: {
            join_accept: true,
            mac_data: true,
            application_data: true,
          },
        }}
      />
    </div>
  ))
  .add('Random', () => (
    <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
      <RoutingPolicy.Sheet
        policy={{
          uplink: {
            join_request: Math.random() >= 0.5,
            mac_data: Math.random() >= 0.5,
            application_data: Math.random() >= 0.5,
            signal_quality: Math.random() >= 0.5,
            localization: Math.random() >= 0.5,
          },
          downlink: {
            join_accept: Math.random() >= 0.5,
            mac_data: Math.random() >= 0.5,
            application_data: Math.random() >= 0.5,
          },
        }}
      />
    </div>
  ))

storiesOf('Routing Policy/Matrix', module)
  .add('Empty', () => (
    <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
      <RoutingPolicy.Matrix
        policy={{
          uplink: {},
          downlink: {},
        }}
      />
    </div>
  ))
  .add('Full', () => (
    <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
      <RoutingPolicy.Matrix
        policy={{
          uplink: {
            join_request: true,
            mac_data: true,
            application_data: true,
            signal_quality: true,
            localization: true,
          },
          downlink: {
            join_accept: true,
            mac_data: true,
            application_data: true,
          },
        }}
      />
    </div>
  ))
  .add('Random', () => (
    <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
      <RoutingPolicy.Matrix
        policy={{
          uplink: {
            join_request: Math.random() >= 0.5,
            mac_data: Math.random() >= 0.5,
            application_data: Math.random() >= 0.5,
            signal_quality: Math.random() >= 0.5,
            localization: Math.random() >= 0.5,
          },
          downlink: {
            join_accept: Math.random() >= 0.5,
            mac_data: Math.random() >= 0.5,
            application_data: Math.random() >= 0.5,
          },
        }}
      />
    </div>
  ))
