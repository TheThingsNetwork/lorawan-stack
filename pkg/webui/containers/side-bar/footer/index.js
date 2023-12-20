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

import SideFooter from '@ttn-lw/components/sidebar/side-footer'

import {
  selectDocumentationUrlConfig,
  selectPageStatusBaseUrlConfig,
  selectSupportLinkConfig,
} from '@ttn-lw/lib/selectors/env'

const supportLink = selectSupportLinkConfig()
const documentationBaseUrl = selectDocumentationUrlConfig()
const statusPageBaseUrl = selectPageStatusBaseUrlConfig()

const Footer = () => (
  <SideFooter
    supportLink={supportLink}
    documentationBaseUrl={documentationBaseUrl}
    statusPageBaseUrl={statusPageBaseUrl}
  />
)

export default Footer
