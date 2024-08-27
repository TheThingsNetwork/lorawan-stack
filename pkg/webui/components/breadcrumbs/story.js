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

import Breadcrumb from './breadcrumb'
import { Breadcrumbs } from './breadcrumbs'

export default {
  title: 'Breadcrumbs',
  component: [Breadcrumbs, Breadcrumb],
}

export const Default = () => {
  const breadcrumbs = [
    <Breadcrumb key="1" path="/applications" content="Applications" />,
    <Breadcrumb key="2" path="/applications/test-app" content="test-app" />,
    <Breadcrumb key="3" path="/applications/test-app/devices" content="Devices" />,
  ]

  return <Breadcrumbs breadcrumbs={breadcrumbs} />
}
