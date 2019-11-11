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
import { Container, Row, Col } from 'react-grid-system'

import DateTime from '../../../lib/components/date-time'

import PropTypes from '../../../lib/prop-types'

import style from './entity-title-section.styl'

const EntityTitleSection = ({ entityName, entityId, description, creationDate, children }) => {
  return (
    <React.Fragment>
      <Container>
        <Row>
          <Col md={12}>
            <div className={style.container}>
              <h1 className={style.title}>{entityName || entityId}</h1>
              <span className={style.id}>
                <strong>ID:</strong> {entityId}
              </span>
              {description && <span className={style.description}>{description}</span>}
              <div className={style.bottom}>
                <div className={style.children}>{children}</div>
                <div className={style.creationDate}>
                  <span>
                    Created <DateTime.Relative value={creationDate} />
                  </span>
                </div>
              </div>
            </div>
          </Col>
        </Row>
      </Container>
      <hr className={style.hRule} />
    </React.Fragment>
  )
}

EntityTitleSection.propTypes = {
  entityName: PropTypes.string,
  entityId: PropTypes.string.isRequired,
  description: PropTypes.string,
  creationDate: PropTypes.string.isRequired,
  children: PropTypes.node.isRequired,
}

EntityTitleSection.defaultProps = {
  entityName: undefined,
  description: undefined,
}

export default EntityTitleSection
