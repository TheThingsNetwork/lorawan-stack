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

import React from 'react'
import { Col, Container, Row } from 'react-grid-system'
import { NavLink, Switch, Route, useRouteMatch } from 'react-router-dom'
import classNames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'
import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'
import { ErrorView } from '@ttn-lw/lib/components/error-view'

import SubViewError from '@console/views/sub-view-error'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './panel-view.styl'

const PanelView = ({ children, className }) => {
  const { path } = useRouteMatch()
  const firstChild = React.Children.toArray(children)[0]

  return (
    <Container className={classNames(className, style.panelView)}>
      <Row>
        <Col className={style.menu} lg={3} xl={2}>
          {React.Children.toArray(children).map(child =>
            React.cloneElement(child, { path: `${path}${child.props.path}`, _isRoute: true }),
          )}
        </Col>
        <Col className={style.panelViewMenuItem} lg={9} xl={10}>
          <ErrorView errorRender={SubViewError}>
            <Switch>
              <Route exact path={path} redirect={`${path}${firstChild.path}`} />
              {React.Children.toArray(children).map(child =>
                React.cloneElement(child, { path: `${path}${child.props.path}`, _isRoute: false }),
              )}
              <NotFoundRoute />
            </Switch>
          </ErrorView>
        </Col>
      </Row>
    </Container>
  )
}

const PanelViewItem = ({ icon, title, path, exact, component: Component, _isRoute }) => {
  if (_isRoute) {
    return (
      <NavLink to={path} className={style.link} activeClassName={style.active} exact={exact}>
        {icon && <Icon icon={icon} className="mr-cs-xs" />} <Message content={title} />
      </NavLink>
    )
  }

  return <Route path={path} component={Component} exact={exact} />
}

PanelViewItem.propTypes = {
  _isRoute: PropTypes.bool,
  component: PropTypes.func.isRequired,
  exact: PropTypes.bool,
  icon: PropTypes.string.isRequired,
  path: PropTypes.string.isRequired,
  title: PropTypes.message.isRequired,
}

PanelViewItem.defaultProps = {
  _isRoute: false,
  exact: false,
}

PanelView.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  className: PropTypes.string,
}

PanelView.defaultProps = {
  className: undefined,
}

PanelView.Item = PanelViewItem

export default PanelView
