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

import tts from '@console/api/tts'

export default Component => props => {
  const [state, setState] = React.useState({
    devAddr: '',
    loading: false,
    error: undefined,
  })
  const { devAddr, loading, error } = state

  const handleGenerateDevAddr = React.useCallback(async () => {
    setState(prev => ({ ...prev, loading: true }))

    try {
      const { dev_addr } = await tts.Ns.generateDevAddress()

      setState({ loading: false, error: undefined, devAddr: dev_addr })
    } catch (error) {
      setState(prev => ({ ...prev, loading: false, error }))
    }
  }, [])

  return (
    <Component
      {...props}
      onGenerate={handleGenerateDevAddr}
      generatedValue={devAddr}
      generatedError={Boolean(error)}
      generatedLoading={loading}
    />
  )
}
