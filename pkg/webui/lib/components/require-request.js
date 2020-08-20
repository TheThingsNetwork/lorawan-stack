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

import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import useRequest from '@ttn-lw/lib/hooks/use-request'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

// `<RequireRequest />` is a utility component that can wrap a component tree
// and dispatch a request action, rendering a loading spinner until the request
// has been resolved.
const RequireRequest = ({ requestAction, children }) => {
  const [fetching] = useRequest(requestAction)
  if (fetching) {
    return (
      <Spinner>
        <Message content={sharedMessages.fetching} />
      </Spinner>
    )
  }

  return children
}

RequireRequest.propTypes = {
  children: PropTypes.node.isRequired,
  requestAction: PropTypes.shape({}).isRequired,
}

export default RequireRequest
