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

import Panel from '.'

export default {
  title: 'Panel',
  component: Panel,
}

export const Default = () => (
  <div style={{ width: '32.5rem' }}>
    <Panel title="Latest news" icon="feed" buttonTitle="Visit our blog" />
  </div>
)

const Example = props => {
  const [active, setActive] = React.useState('option-1')

  const handleChange = React.useCallback(
    (_, value) => {
      setActive(value)
    },
    [setActive],
  )

  return <Panel {...props} activeToggle={active} onToggleClick={handleChange} />
}

export const WithToggle = () => {
  const options = [
    { label: 'Option 1', value: 'option-1' },
    { label: 'Option 2', value: 'option-2' },
    { label: 'Option 3', value: 'option-3' },
  ]

  return (
    <div style={{ width: '32.5rem' }}>
      <Example title="Your top entities" icon="star" toggleOptions={options} />
    </div>
  )
}