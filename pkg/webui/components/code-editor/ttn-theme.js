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

/* eslint-disable no-undef*/
ace.define('ace/theme/ttn', ['require', 'exports', 'module', 'ace/lib/dom'], function(
  acequire,
  exports,
  module,
) {
  exports.isDark = false
  exports.cssClass = 'ace-ttn'
  exports.cssText =
    '\
.ace_scroller.ace_scroll-left {\
  box-shadow: none;\
}\
.ace-ttn .ace_gutter {\
background: #e8e8e8;\
color: #AAA;\
}\
.ace-ttn  {\
background: #fff;\
color: #000;\
}\
// begin language\
.ace-ttn .ace_string {\
color: #D14;\
}\
.ace-ttn .ace_constant.ace_numeric {\
color: #099;\
}\
.ace-ttn .ace_constant.ace_buildin {\
color: #0086B3;\
}\
.ace-ttn .ace_entity.ace_name.ace_function,\
.ace-ttn .ace_support.ace_function,\
.ace-ttn .ace_support.ace_class {\
color: #DD4A68;\
}\
.ace-ttn .ace_keyword,\
.ace-ttn .ace_meta,\
.ace-ttn .ace_storage,\
.ace-ttn .ace_storage.ace_type,\
.ace-ttn .ace_support.ace_type{\
color: #07a;\
font-style: italic;\
font-weight: bold;\
}\
.ace-ttn .ace_keyword.ace_operator {\
  color: #a67f59;\
}\
.ace-ttn .ace_constant.ace_character,\
.ace-ttn .ace_constant.ace_language,\
.ace-ttn .ace_keyword.ace_other.ace_unit,\
.ace-ttn .ace_support.ace_constant{\
  color: #905;\
}\
.ace-ttn .ace_comment {\
color: #8E908C;\
}\
.ace-ttn .ace_heading,\
.ace-ttn .ace_markup.ace_heading,\
.ace-ttn .ace_string,\
.ace-ttn .ace_constant.ace_boolean {\
  color: #718C00;\
}\
.ace-ttn .ace_paren {\
  color: #008066;\
}\
.ace-ttn .ace_string.ace_regexp {\
color: #e90;\
font-weight: normal;\
}\
.ace-ttn .ace_variable.ace_instance {\
color: teal;\
}\
.ace-ttn .ace_constant.ace_language {\
font-weight: bold;\
}\
.ace-ttn .ace_cursor {\
color: black;\
}\
.ace-ttn.ace_focus .ace_marker-layer .ace_active-line {\
background: #91d3ff;\
opacity: 0.5;\
}\
.ace-ttn .ace_marker-layer .ace_active-line {\
background: rgb(245, 245, 245);\
}\
.ace-ttn .ace_marker-layer .ace_selection {\
background: rgb(181, 213, 255);\
}\
.ace-ttn.ace_multiselect .ace_selection.ace_start {\
box-shadow: 0 0 3px 0px white;\
}\
.ace-ttn.ace_nobold .ace_line > span {\
font-weight: normal !important;\
}\
.ace-ttn .ace_marker-layer .ace_step {\
background: rgb(252, 255, 0);\
}\
.ace-ttn .ace_marker-layer .ace_stack {\
background: rgb(164, 229, 101);\
}\
.ace-ttn .ace_marker-layer .ace_bracket {\
margin: -1px 0 0 -1px;\
border: 1px solid rgb(192, 192, 192);\
}\
.ace-ttn .ace_gutter-active-line {\
background-color : rgba(0, 0, 0, 0.07);\
}\
.ace-ttn .ace_marker-layer .ace_selected-word {\
border: 1px solid #DFDFDF;\
}\
.ace-ttn .ace_invisible {\
color: #BFBFBF\
}\
.ace-ttn .ace_print-margin {\
width: 1px;\
background: #e8e8e8;\
}\
.ace-ttn .ace_print-margin {\
display: none;\
}\
.ace-ttn .ace_indent-guide {\
border-right: 1px solid rgba(100, 100, 100, .3);\
margin-top: -1px;\
background: none;\
}'

  const dom = acequire('../lib/dom')
  dom.importCssString(exports.cssText, exports.cssClass)
})
