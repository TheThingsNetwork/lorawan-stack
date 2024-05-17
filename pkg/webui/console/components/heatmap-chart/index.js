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

import ReactApexChart from 'react-apexcharts'
import React from 'react'
import { random } from 'lodash'

import PropTypes from '@ttn-lw/lib/prop-types'

import options from './chart-utils'

const HeatmapChart = ({ showRandomValues, data, height }) => {
  const series = showRandomValues
    ? new Array(random(4, 18)).fill(1).map(() => ({
        name: `${random(800, 999)}.${random(100, 999)}`,
        data: Array.from({ length: 50 }, () => ({
          x: Date.now() + random(0, 1000) * 1000,
          y: random(0, 100),
        })),
      }))
    : data

  return <ReactApexChart options={options} series={series} type="heatmap" height={height} />
}

HeatmapChart.propTypes = {
  data: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string,
      data: PropTypes.arrayOf(PropTypes.number),
    }),
  ).isRequired,
  height: PropTypes.number,
  showRandomValues: PropTypes.bool.isRequired,
}

HeatmapChart.defaultProps = {
  height: 250,
}

export default HeatmapChart
