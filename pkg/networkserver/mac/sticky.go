// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package mac

// The RxParamSetup, RxTimingSetup, TxParamSetup and DlChannel MAC commands
// use the sticky answer mechanism:
// After an end device receives a {Command}Req, it will piggy back the
// associated {Command}Ans on every subsequent uplink until a class A downlink
// reaches the end device. This means that we will observe {Command}Ans
// even in the absence of a request. We should allow this, provided the
// previous uplink messages did contain a {Command}Ans as well.
// A corollary of the sticky answer mechanism is that we should not schedule
// {Command}Req if an {Command}Ans was found in the last uplink received
// from the end device, as we cannot distinguish between a sticky answer and
// a real answer for the new {Command}Req.
