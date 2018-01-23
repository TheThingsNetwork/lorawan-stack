// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
