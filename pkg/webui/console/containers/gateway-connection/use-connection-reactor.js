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

import { useEffect, useRef } from 'react'
import { useDispatch, useSelector } from 'react-redux'

import {
  isGsDownlinkSendEvent,
  isGsStatusReceiveEvent,
  isGsUplinkReceiveEvent,
} from '@ttn-lw/lib/selectors/event'

import { updateGatewayStatistics } from '@console/store/actions/gateways'

import { selectLatestGatewayEvent } from '@console/store/selectors/gateways'

const useConnectionReactor = gtwId => {
  const latestEvent = useSelector(state => selectLatestGatewayEvent(state, gtwId))
  const dispatch = useDispatch()
  const prevEvent = useRef(null)

  useEffect(() => {
    if (Boolean(latestEvent) && latestEvent !== prevEvent.current) {
      const { name } = latestEvent
      const isHeartBeatEvent =
        isGsDownlinkSendEvent(name) || isGsUplinkReceiveEvent(name) || isGsStatusReceiveEvent(name)

      if (isHeartBeatEvent) {
        dispatch(updateGatewayStatistics(gtwId))
      }
      prevEvent.current = latestEvent
    }
  }, [dispatch, gtwId, latestEvent])
}
export default useConnectionReactor
