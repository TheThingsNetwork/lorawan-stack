// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

//+build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	keyFile        = get("GITHUB_APP_KEY_FILE", ".make/github.pem")
	issuer         = get("GITHUB_ISSUER", "4949")
	repo           = get("TRAVIS_REPO_SLUG", "")
	installationID = get("GITHUB_INSTALLATION_ID", "")
	pr             = get("TRAVIS_PULL_REQUEST", "")
	jobNumber      = get("TRAVIS_JOB_NUMBER", "")
)

func get(key, def string) string {
	val := os.Getenv(key)
	if val == "" && def == "" {
		log.Fatalf("Env key %s not set", key)
	}

	if val == "" {
		return def
	}

	return val
}

func getJWT() string {
	iss, err := strconv.ParseUint(issuer, 10, 64)
	if err != nil {
		log.WithError(err).Fatal("Could not parse github issuer")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": iss,
		"iat": time.Now().Add(-10 * time.Second).Unix(),
		"exp": time.Now().Add(7 * time.Minute).Unix(),
	})

	pem, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.WithError(err).Fatal("Could not find private key file")
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		log.WithError(err).Fatal("Could not parse private key")
	}

	str, err := token.SignedString(key)
	if err != nil {
		log.WithError(err).Fatal("Could sign token")
	}

	return str
}

func getToken() string {
	jwt := getJWT()
	url := fmt.Sprintf("https://api.github.com/installations/%s/access_tokens", installationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.WithError(err).Fatal("Failed to create HTTP request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Fatal("Failed to perform request")
	}

	if resp.StatusCode != 201 {
		log.WithField("code", resp.StatusCode).Fatal("Unexpected response code")
	}

	defer resp.Body.Close()
	res := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse response")
	}

	token, ok := res["token"].(string)
	if !ok {
		log.Fatal("Failed to get token from request")
	}

	return token
}

func postComment(comment string) {
	token := getToken()
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", repo, pr)
	b, err := json.Marshal(map[string]interface{}{
		"body": comment,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to marshal body")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		log.WithError(err).Fatal("Failed to create HTTP request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("toket n %s", token))
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Fatal("Failed to perform request")
	}

	if resp.StatusCode != 201 {
		log.WithField("code", resp.StatusCode).Fatal("Unexpected response code")
	}
}

func main() {
	comment := strings.Join(os.Args[1:], "\n")
	fmt.Println(comment)

	// only comment if job number is the only or the first job
	if nums := strings.Split(jobNumber, "."); len(nums) <= 1 || nums[1] == "1" {
		postComment(comment)
	}
}
