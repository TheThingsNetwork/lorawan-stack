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
