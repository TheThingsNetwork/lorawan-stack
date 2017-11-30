// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

//+build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang/dep"
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

	sw, err := dep.NewSafeWriter(p.Manifest, nil, p.Lock, dep.VendorNever)
	if err != nil {
		logger.Fatalf("Failed to initalize dep writer: %s", err)
	}
	if err := sw.Write(p.AbsRoot, nil, false, logger); err != nil {
		logger.Fatalf("Failed to write files: %s", err)
	}
	os.Exit(0)
}
