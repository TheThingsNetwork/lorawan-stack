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

/* eslint-disable no-invalid-this */

import traverse from 'traverse'
import queryString from 'query-string'

/** Class used to marshal data shapes. Currently a stub. */
class Marshaler {
  static options (options) {
    if (Object.keys(options).length === 0) {
      return null
    }

    const query = {}

    if ('select' in options) {
      query.field_mask = {}
      query.field_mask.paths = options.select
    }

    return query
  }

  static query (params) {
    return queryString.stringify(params)
  }

  static payload (payload, wrap) {
    let res = payload

    if (wrap) {
      res = { [wrap]: res }
    }

    return res
  }

  static payloadListResponse (entity, { data = {}, headers = {}}, transform) {
    const list = data[entity]

    if (!list) {
      return { [entity]: [], totalCount: 0 }
    }

    const totalCount = parseInt(headers['x-total-count'])

    if (isNaN(totalCount)) {
      return { [entity]: [], totalCount: 0 }
    }

    const transforedList = transform ? list.map(transform) : list

    return { [entity]: transforedList, totalCount }
  }

  static unwrapApplications (result, transform) {
    return this.payloadListResponse('applications', result, transform)
  }

  static fieldMaskFromPatch (patch) {
    return traverse(patch).paths().slice(1).map( e => e.join('.') )
  }

  static fieldMask (fieldMask) {
    return { paths: fieldMask }
  }
}

export default Marshaler
