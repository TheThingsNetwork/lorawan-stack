// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package interop

import (
	"encoding/json"
	"net/http"

	ttnerrors "go.thethings.network/lorawan-stack/v3/pkg/errors"
)

func writeError(w http.ResponseWriter, _ *http.Request, header MessageHeader, err error) {
	answerHeader, headerErr := header.AnswerHeader()
	if headerErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code, desc := ResultOther, err.Error()
	for errDef, protocolCode := range errorResultCodes {
		if ttnerrors.Resemble(errDef, err) {
			if c, ok := protocolCode[header.ProtocolVersion]; ok {
				code = c
				if cause := ttnerrors.Cause(err); cause != nil {
					desc = cause.Error()
				} else {
					desc = ""
				}
			}
			break
		}
	}

	msg := ErrorMessage{
		MessageHeader: answerHeader,
		Result: Result{
			ResultCode:  code,
			Description: desc,
		},
	}
	json.NewEncoder(w).Encode(msg) //nolint:errcheck
}
