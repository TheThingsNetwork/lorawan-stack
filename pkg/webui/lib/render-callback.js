// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

/**
 * Allows react components to use the render prop design pattern without explicitly
 * specifying `render`, `children` and `children` as a function props.
 *
 * @param {object} props - The props of a component.
 * @param {Function} props.render - The render function.
 * @param {object} props.children - The component children.
 * @param {object} context - The context to be passed to children.
 *
 * @returns {*} - Processed children by the provided `render`, `children` function or just `children`
 * depending on the props.
 */
const renderCallback = ({ render, children }, context) => {
  if (render) {
    return render(context)
  } else if (typeof children === 'function') {
    return children(context)
  }

  return children
}

export default renderCallback
