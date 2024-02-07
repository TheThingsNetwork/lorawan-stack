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

import ProfilePicture from '.'

const pp = {
  sizes: {
    256: 'https://www.gravatar.com/avatar/205e460b479e2e5b48aec07710c08d50',
  },
}

export default {
  title: 'Profile Picture',
}

export const Default = () => (
  <div style={{ height: '8rem' }}>
    <ProfilePicture profilePicture={pp} />
  </div>
)
