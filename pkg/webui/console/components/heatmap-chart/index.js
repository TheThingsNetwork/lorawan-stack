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
import React, { useMemo } from 'react'
import { random } from 'lodash'

import PropTypes from '@ttn-lw/lib/prop-types'

import getOptions from './chart-utils'

const HeatmapChart = ({ showRandomValues, data, height }) => {
  const series = useMemo(() => {
    if (showRandomValues) {
      const now = new Date()
      now.setMinutes(0, 0, 0)
      const baseTimestamp = now.getTime()

      return Array.from({ length: random(10, 18) }, () => ({
        name: `${random(800, 999)}.${random(100, 999)}`,
        data: Array.from({ length: 24 }, (_, i) => ({
          x: baseTimestamp - (24 - i - 1) * 60 * 60 * 1000,
          y: random(1, 100),
        })),
      }))
    }
    return data
  }, [data, showRandomValues])

  const calculateHeight = () => {
    const baseRowHeight = 30
    const rowCount = series.length

    // Ideal full height
    const fullHeight = rowCount * baseRowHeight

    if (fullHeight <= height) {
      return fullHeight
    }

    // Shrink row height to fit within maxHeight
    const scaledRowHeight = height / rowCount
    return rowCount * scaledRowHeight
  }

  return (
    <div className="w-full">
      <ReactApexChart
        options={getOptions()}
        series={series}
        type="heatmap"
        height={calculateHeight()}
      />
    </div>
  )
}

HeatmapChart.propTypes = {
  data: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string,
      data: PropTypes.arrayOf(
        PropTypes.shape({
          x: PropTypes.number,
          y: PropTypes.number,
        }),
      ),
    }),
  ).isRequired,
  height: PropTypes.number,
  showRandomValues: PropTypes.bool.isRequired,
}

HeatmapChart.defaultProps = {
  height: 350,
}

export default HeatmapChart
