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

import Spinner from '.'

export default {
  title: 'Spinner',
}

export const Default = () => <Spinner />
export const WithChildren = () => <Spinner>This is a message</Spinner>

WithChildren.story = {
  name: 'With children',
}

export const WithInlineChildren = () => <Spinner inline>This is a message</Spinner>

WithInlineChildren.story = {
  name: 'With inline children',
}

export const Centered = () => <Spinner center>This is a message</Spinner>
export const Faded = () => <Spinner faded />
export const Small = () => <Spinner small />
export const Micro = () => <Spinner micro />
