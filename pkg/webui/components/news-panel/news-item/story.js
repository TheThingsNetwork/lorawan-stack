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

import loginVisual from '@assets/img/layout/bg/login-visual.jpg'

import NewsItem from '.'

export default {
  title: 'Panel/News Panel/News Item',
  component: NewsItem,
}

export const Default = () => (
  <div style={{ width: '399px' }}>
    <NewsItem
      articleTitle="Long title of the latest post on our blog that will take more that two line to fit in here"
      articleImage={loginVisual}
      articleDate="2024-01-01T00:00:00Z"
    />
  </div>
)