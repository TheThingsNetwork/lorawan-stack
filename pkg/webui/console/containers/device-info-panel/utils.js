// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import {
  IconActivityHeartbeat,
  IconAntennaBars4,
  IconArrowUpBar,
  IconArrowWaveRightUp,
  IconArrowsExchange,
  IconBattery,
  IconBounceLeft,
  IconBolt,
  IconBrandSpeedtest,
  IconBroadcast,
  IconBulb,
  IconCircuitInductor,
  IconCircuitPushbutton,
  IconCircuitResistor,
  IconClock,
  IconCloudFog,
  IconCloudRain,
  IconCloudStorm,
  IconCircleDotted,
  IconDirectionArrows,
  IconDroplet,
  IconDroplets,
  IconFilter,
  IconFlask,
  IconGauge,
  IconGps,
  IconGrain,
  IconHexagon,
  IconInputCheck,
  IconLink,
  IconMagnet,
  IconPower,
  IconPlugConnected,
  IconRadar,
  IconRipple,
  IconRotate3d,
  IconRuler,
  IconRulerMeasure,
  IconSun,
  IconTemperature,
  IconTemperatureSun,
  IconTiltShift,
  IconUserScan,
  IconUsers,
  IconVolume2,
  IconWaveSine,
  IconWaveSquare,
  IconWeight,
  IconWifi,
  IconWind,
  IconWindsock,
} from '@ttn-lw/components/icon'

const sensorIconMap = Object.freeze({
  '4-20 ma': IconArrowWaveRightUp,
  accelerometer: IconGauge,
  altitude: IconArrowUpBar,
  'analog input': IconWaveSquare,
  auxiliary: IconPlugConnected,
  barometer: IconDirectionArrows,
  battery: IconBattery,
  button: IconCircuitPushbutton,
  bvoc: IconCircleDotted,
  co: IconHexagon,
  co2: IconHexagon,
  conductivity: IconCircuitResistor,
  current: IconCircuitInductor,
  'digital input': IconInputCheck,
  'dissolved oxygen': IconDroplet,
  distance: IconRuler,
  dust: IconGrain,
  energy: IconBolt,
  gps: IconGps,
  gyroscope: IconRotate3d,
  h2s: IconHexagon,
  humidity: IconDroplets,
  iaq: IconWind,
  level: IconRulerMeasure,
  light: IconBulb,
  lightning: IconCloudStorm,
  link: IconLink,
  magnetometer: IconMagnet,
  moisture: IconDroplet,
  motion: IconBounceLeft,
  no: IconHexagon,
  no2: IconHexagon,
  o3: IconHexagon,
  'particulate matter': IconHexagon,
  ph: IconFlask,
  pir: IconUsers,
  'pm2.5': IconHexagon,
  pm10: IconHexagon,
  potentiometer: IconWaveSquare,
  power: IconPower,
  precipitation: IconCloudRain,
  pressure: IconCloudFog,
  proximity: IconUserScan,
  'pulse count': IconActivityHeartbeat,
  'pulse frequency': IconActivityHeartbeat,
  radar: IconRadar,
  rainfall: IconCloudRain,
  rssi: IconBroadcast,
  'smart valve': IconFilter,
  snr: IconAntennaBars4,
  so2: IconHexagon,
  'solar radiation': IconSun,
  sound: IconVolume2,
  strain: IconArrowsExchange,
  'surface temperature': IconTemperatureSun,
  temperature: IconTemperature,
  tilt: IconTiltShift,
  time: IconClock,
  tvoc: IconWind,
  uv: IconSun,
  'vapor pressure': IconCloudFog,
  velocity: IconBrandSpeedtest,
  vibration: IconWaveSine,
  voltage: IconBolt,
  'water potential': IconDroplet,
  water: IconRipple,
  weight: IconWeight,
  'wifi ssid': IconWifi,
  'wind direction': IconWindsock,
  'wind speed': IconWind,
})

export default sensorIconMap
