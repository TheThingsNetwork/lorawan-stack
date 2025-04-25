// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

import { capitalize } from 'lodash'

import {
  IconArrowsDoubleSwNe,
  IconBattery4,
  IconClock,
  IconCurrentLocation,
  IconDroplet,
  IconPropeller,
  IconRulerMeasure,
  IconSeedling,
  IconWind,
} from '@ttn-lw/components/icon'

import sharedMessages from '@ttn-lw/lib/shared-messages'

export const directions = {
  N: 'North',
  NNE: 'North North East',
  NE: 'North East',
  ENE: 'East North East',
  E: 'East',
  ESE: 'East South East',
  SE: 'South East',
  SSE: 'South South East',
  S: 'South',
  SSW: 'South South West',
  SW: 'South West',
  WSW: 'West South West',
  W: 'West',
  WNW: 'West North West',
  NW: 'North West',
  NNW: 'North North West',
}

export const degreesToCompass = degrees => {
  const directions = [
    'N',
    'NNE',
    'NE',
    'ENE',
    'E',
    'ESE',
    'SE',
    'SSE',
    'S',
    'SSW',
    'SW',
    'WSW',
    'W',
    'WNW',
    'NW',
    'NNW',
  ]
  const index = Math.round(degrees / 22.5) % 16
  return directions[index]
}

export const getPayloadSections = (payload, m) => {
  const sections = []

  if (payload.time) {
    sections.push({
      title: sharedMessages.time,
      icon: IconClock,
      groups: [...(payload.time ? [{ value: new Date(payload.time).toLocaleString() }] : [])],
    })
  }

  if (payload.battery) {
    sections.push({
      title: m.battery,
      icon: IconBattery4,
      groups: [...(payload.battery ? [{ value: `${payload.battery}V` }] : [])],
    })
  }

  if (payload.soil) {
    sections.push({
      title: sharedMessages.normalizedPayloadSoil,
      icon: IconSeedling,
      groups: [
        ...(payload.soil.depth ? { value: `${payload.soil.depth} cm depth` } : []),
        ...(payload.soil.moisture ? { value: `${payload.soil.moisture}% moisture` } : []),
        ...(payload.soil.temperature ? { value: `${payload.soil.temperature}°C` } : []),
        ...(payload.soil.ec ? { value: `${payload.soil.ec} dS/m` } : []),
        ...(payload.soil.pH ? { value: `pH ${payload.soil.pH}` } : []),
        ...(payload.soil.n ? { value: `Nitrogen ${payload.soil.n} ppm` } : []),
        ...(payload.soil.p ? { value: `Phosporus ${payload.soil.p} ppm` } : []),
        ...(payload.soil.k ? { value: `Potassium ${payload.soil.k} ppm` } : []),
      ],
    })
  }

  if (payload.air) {
    sections.push({
      title: sharedMessages.normalizedPayloadAir,
      icon: IconPropeller,
      groups: [
        ...(payload.air.location ? { value: capitalize(payload.air.location) } : []),
        ...(payload.air.temperature ? { value: `${payload.air.temperature}°C` } : []),
        ...(payload.air.relativeHumidity
          ? {
              value: `${payload.air.relativeHumidity}% relative humidity`,
            }
          : []),
        ...(payload.air.pressure ? { value: `${payload.air.pressure} hPa` } : []),
        ...(payload.air.co2 ? { value: `CO2 ${payload.air.co2} ppm` } : []),
        ...(payload.air.lightIntensity
          ? {
              value: `${payload.air.lightIntensity} lux`,
            }
          : []),
      ],
    })
  }

  if (payload.wind) {
    sections.push({
      title: sharedMessages.normalizedPayloadWind,
      icon: IconWind,
      groups: [
        ...(payload.wind.speed ? { value: `${payload.wind.speed} m/s` } : []),
        ...(payload.wind.direction
          ? { value: `${directions[degreesToCompass(payload.wind.direction)]}` }
          : []),
      ],
    })
  }

  if (payload.water) {
    sections.push({
      title: m.water,
      icon: IconDroplet,
      groups: [
        ...(payload.water.leak !== undefined
          ? {
              value: payload.water.leak ? 'Leak detected' : [],
            }
          : []),
        ...(payload.water.temperature?.min
          ? {
              value: `Min ${payload.water.temperature.min}°C`,
            }
          : []),
        ...(payload.water.temperature?.max
          ? {
              value: `Max ${payload.water.temperature.max}°C`,
            }
          : []),
        ...(payload.water.temperature?.avg
          ? {
              value: `Average ${payload.water.temperature.avg}°C`,
            }
          : []),
        ...(payload.water.temperature?.current
          ? {
              value: `Current ${payload.water.temperature.current}°C`,
            }
          : []),
      ],
    })
  }

  if (payload.metering?.water?.total !== undefined) {
    sections.push({
      title: m.metering,
      icon: IconRulerMeasure,
      groups: [
        ...(payload.metering.water.total
          ? {
              value: `${payload.metering.water.total}L water`,
            }
          : []),
      ],
    })
  }

  if (payload.action) {
    sections.push({
      title: m.action,
      icon: IconArrowsDoubleSwNe,
      groups: [
        ...(payload.action.motion?.detected !== undefined
          ? {
              value: payload.action.motion.detected ? 'Motion detected' : [],
            }
          : []),
        ...(payload.action.motion?.count !== undefined
          ? {
              value: `${payload.action.motion.count} motion events`,
            }
          : []),
        ...(payload.action.contactState ? { value: `Sensor ${payload.action.contactState}` } : []),
      ],
    })
  }

  if (payload.position) {
    sections.push({
      title: m.position,
      icon: IconCurrentLocation,
      groups: [
        ...(payload.position.latitude ? { value: `Latitude: ${payload.position.latitude}°` } : []),
        ...(payload.position.longitude
          ? { value: `Longitude: ${payload.position.longitude}` }
          : []),
      ],
    })
  }
  return sections
}
