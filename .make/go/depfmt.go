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

//+build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang/dep"
	"github.com/golang/dep/gps"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "From project root:\n")
		fmt.Fprintf(os.Stderr, "go run depfmt.go\n")
	}
	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Failed to get working directory: %s", err)
	}

	p, err := (&dep.Ctx{
		WorkingDir: wd,
		GOPATHs:    strings.Split(os.Getenv("GOPATH"), ":"),
		Out:        logger,
		Err:        log.New(os.Stderr, "", 0),
	}).LoadProject()
	if err != nil {
		logger.Fatalf("Failed to load project: %s", err)
	}

	sw, err := dep.NewSafeWriter(p.Manifest, nil, p.Lock, dep.VendorNever, gps.CascadingPruneOptions{
		DefaultOptions: gps.PruneNestedVendorDirs | gps.PruneUnusedPackages | gps.PruneNonGoFiles | gps.PruneGoTestFiles,
	}, nil)
	if err != nil {
		logger.Fatalf("Failed to initalize dep writer: %s", err)
	}
	if err := sw.Write(p.AbsRoot, nil, false, logger); err != nil {
		logger.Fatalf("Failed to write files: %s", err)
	}
	os.Exit(0)
}
