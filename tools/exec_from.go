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

//+build ignore

package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
}

func main() {
	dir := flag.String("dir", "", "directory to run from")
	flag.Parse()

	if n := len(flag.Args()); n < 1 {
		log.Fatalf("Expected at least 1 argument, got %d", n)
	}

	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
	cmd.Dir = *dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			os.Exit(err.ExitCode())
		default:
			log.Fatalf("Unknown error received: %s", err)
		}
	}
}
