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

const options = {
  chart: {
    type: 'heatmap',
    toolbar: {
      show: false,
    },
    animations: {
      enabled: false,
    },
    zoom: {
      enabled: false,
    },
    events: {
      mounted: chart => {
        chart.windowResizeHandler()
      },
    },
  },
  dataLabels: {
    enabled: false,
  },
  colors: ['#152F66'],
  grid: {
    show: false,
    strokeDashArray: 0,
    padding: {
      left: 0,
      right: 0,
      bottom: -13,
      top: -30,
    },
  },
  yaxis: {
    labels: {
      align: 'left',
      offsetX: 10,
    },
  },
  xaxis: {
    labels: {
      show: false,
    },
    axisBorder: {
      show: false,
    },
    tooltip: {
      enabled: false,
    },
  },
  tooltip: {
    x: {
      show: false,
    },
    marker: {
      show: false,
    },
    custom: ({ series, seriesIndex, dataPointIndex, w }) => {
      const val = series[seriesIndex][dataPointIndex]
      const date = new Date(w.globals.seriesX[seriesIndex][dataPointIndex])
      const formattedDate = date.toLocaleDateString('en-US', {
        month: '2-digit',
        day: '2-digit',
      })
      const formattedTime = date.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
      })
      return `<div class="pl-cs-s pr-cs-s pt-cs-xs pb-cs-xs"><strong>${formattedDate} ${formattedTime}</strong>: ${val}</div>`
    },
  },
}

export default options
