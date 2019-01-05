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
