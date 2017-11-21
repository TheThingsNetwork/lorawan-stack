// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

//+build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

var versionRegexp = regexp.MustCompile("(?:v)?(\\d+)\\.(\\d+)\\.(\\d+)(?:-([a-z]+)(\\d*))?")

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run bump.go 1.2.3-rc4 [major|minor|patch|rc]")
	}
	version := versionRegexp.FindStringSubmatch(os.Args[1])
	if len(version) != 6 {
		log.Fatalf("Invalid version %s", os.Args[1])
	}
	major, _ := strconv.Atoi(version[1])
	minor, _ := strconv.Atoi(version[2])
	patch, _ := strconv.Atoi(version[3])
	preType := version[4]
	preNum, err := strconv.Atoi(version[5])
	if err != nil || preNum == 0 {
		preNum = 1
	}

	switch os.Args[2] {
	case "major":
		major++
		minor = 0
		patch = 0
		preType = ""
	case "minor":
		minor++
		patch = 0
		preType = ""
	case "patch":
		patch++
		preType = ""
	case "rc":
		if preType == "rc" {
			preNum++
		} else {
			major++
			minor = 0
			patch = 0
			preType = "rc"
			preNum = 1
		}
	default:
		log.Fatalf("Invalid bump type %s", os.Args[2])
	}

	bumped := fmt.Sprintf("%d.%d.%d", major, minor, patch)
	if preType != "" {
		bumped += "-" + preType
		if preNum > 1 || preType == "rc" {
			bumped += strconv.Itoa(preNum)
		}
	}

	fmt.Println(bumped)
}
