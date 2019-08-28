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

package log

import (
	"fmt"
	"strings"
)

func ExampleMiddleware() {
	// build our custom handler (needed for example tests, because it expects us to use fmt.Println)
	handler := HandlerFunc(func(entry Entry) error {
		fmt.Printf("%s: %s\n", strings.ToUpper(entry.Level().String()), entry.Message())
		return nil
	})

	logger, _ := NewLogger(WithHandler(handler), WithLevel(InfoLevel))

	// printer is a middleware that prints the strings be
	printer := MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(entry Entry) error {
			fmt.Println("== Before ==")
			err := next.HandleLog(entry)
			fmt.Println("== After ==")

			return err
		})
	})

	messages := 0

	// counter is a middleware that counts the messages
	counter := MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(entry Entry) error {
			messages++
			return next.HandleLog(entry)
		})
	})

	logger.Use(printer)
	logger.Use(counter)

	logger.Info("Hey!")
	fmt.Println("Number of messages:", messages)

	// Output:
	// == Before ==
	// INFO: Hey!
	// == After ==
	// Number of messages: 1
}
