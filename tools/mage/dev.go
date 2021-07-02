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

package ttnmage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Dev namespace.
type Dev mg.Namespace

// Misspell fixes common spelling mistakes in files.
func (Dev) Misspell() error {
	if mg.Verbose() {
		fmt.Println("Fixing common spelling mistakes in files")
	}
	return runGoTool("github.com/client9/misspell/cmd/misspell", "-w", "-i", "mosquitto",
		".editorconfig",
		".gitignore",
		".goreleaser.release.yml",
		".goreleaser.snapshot.yml",
		".revive.toml",
		".travis.yml",
		"api",
		"cmd",
		"config",
		"CONTRIBUTING.md",
		"DEVELOPMENT.md",
		"docker-compose.yml",
		"Dockerfile",
		"lorawan-stack.go",
		"Makefile",
		"pkg",
		"README.md",
		"sdk",
		"SECURITY.md",
		"tools",
	)
}

var (
	sqlDatabase           = "cockroach"
	redisDatabase         = "redis"
	devDatabases          = []string{sqlDatabase, redisDatabase}
	devDataDir            = ".env/data"
	devDir                = ".env"
	devDatabaseName       = "ttn_lorawan_dev"
	devDockerComposeFlags = []string{"-p", "lorawan-stack-dev"}
	databaseURI           = fmt.Sprintf("postgresql://root@localhost:26257/%s?sslmode=disable", devDatabaseName)
)

func dockerComposeFlags(args ...string) []string {
	return append(devDockerComposeFlags, args...)
}

func execDockerCompose(args ...string) error {
	_, err := sh.Exec(nil, os.Stdout, os.Stderr, "docker-compose", dockerComposeFlags(args...)...)
	return err
}

// SQLStart starts the SQL database of the development environment.
func (Dev) SQLStart() error {
	if mg.Verbose() {
		fmt.Println("Starting SQL database")
	}
	if err := execDockerCompose(append([]string{"up", "-d"}, sqlDatabase)...); err != nil {
		return err
	}
	return execDockerCompose("ps")
}

// SQLStop stops the SQL database of the development environment.
func (Dev) SQLStop() error {
	if mg.Verbose() {
		fmt.Println("Stopping SQL database")
	}
	return execDockerCompose(append([]string{"stop"}, sqlDatabase)...)
}

// SQLRestoreSnapshot restores the previously taken snapshot, thus restoring all previously
// snapshoted databases.
func (d Dev) SQLRestoreSnapshot() error {
	mg.Deps(Dev.SQLStop)
	if mg.Verbose() {
		fmt.Println("Restoring DB snapshot")
	}
	to := filepath.Join(devDataDir, "cockroach")
	from := filepath.Join(devDataDir, "cockroach-snap")
	if err := os.RemoveAll(to); err != nil {
		return err
	}
	if err := sh.Copy(from, to); err != nil {
		return err
	}
	return d.SQLStart()
}

// SQLDump performs an SQL database dump of the dev database to the .cache folder.
func (Dev) SQLDump() error {
	if mg.Verbose() {
		fmt.Println("Saving sql database dump")
	}
	if err := os.MkdirAll(".cache", 0755); err != nil {
		return err
	}
	output, err := sh.Output("docker-compose", dockerComposeFlags("exec", "-T", "cockroach", "./cockroach", "dump", devDatabaseName, "--insecure")...)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(".cache", "sqldump.sql"), []byte(output), 0644)
}

// SQLRestore restores the dev database using a previously generated dump.
func (Dev) SQLRestore(ctx context.Context) error {
	if mg.Verbose() {
		fmt.Println("Restoring database from dump")
	}
	db, err := store.Open(ctx, databaseURI)
	if err != nil {
		return err
	}
	defer db.Close()

	b, err := ioutil.ReadFile(filepath.Join(".cache", "sqldump.sql"))
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %s;
		CREATE DATABASE %s;
		%s`,
		devDatabaseName, devDatabaseName, string(b)),
	).Error
}

// RedisFlush deletes all keys from redis.
func (Dev) RedisFlush() error {
	if mg.Verbose() {
		fmt.Println("Deleting all keys from redis")
	}

	keys, err := sh.Output("docker-compose", dockerComposeFlags("exec", "-T", "redis", "redis-cli", "keys", "ttn:v3:*")...)
	if err != nil {
		return err
	}
	ks := strings.Split(keys, "\n")
	if len(ks) == 0 {
		return nil
	}
	flags := dockerComposeFlags(append([]string{"exec", "-T", "redis", "redis-cli", "del"}, ks...)...)
	_, err = sh.Exec(nil, nil, os.Stderr, "docker-compose", flags...)
	return err
}

// DBStart starts the databases of the development environment.
func (Dev) DBStart() error {
	if mg.Verbose() {
		fmt.Println("Starting dev databases")
	}
	if err := execDockerCompose(append([]string{"up", "-d"}, devDatabases...)...); err != nil {
		return err
	}
	return execDockerCompose("ps")
}

// DBStop stops the databases of the development environment.
func (Dev) DBStop() error {
	if mg.Verbose() {
		fmt.Println("Stopping dev databases")
	}
	return execDockerCompose(append([]string{"stop"}, devDatabases...)...)
}

// DBErase erases the databases of the development environment.
func (Dev) DBErase() error {
	mg.Deps(Dev.DBStop)
	if mg.Verbose() {
		fmt.Println("Erasing dev databases")
	}
	return os.RemoveAll(devDataDir)
}

// DBSQL starts an SQL shell.
func (Dev) DBSQL() error {
	mg.Deps(Dev.DBStart)
	if mg.Verbose() {
		fmt.Println("Starting SQL shell")
	}
	return execDockerCompose("exec", "cockroach", "./cockroach", "sql",
		"--insecure",
		"-d", devDatabaseName,
	)
}

// DBRedisCli starts a Redis-CLI shell.
func (Dev) DBRedisCli() error {
	mg.Deps(Dev.DBStart)
	if mg.Verbose() {
		fmt.Println("Starting Redis-CLI shell")
	}
	return execDockerCompose("exec", "redis", "redis-cli")
}

// InitDeviceRepo initializes the device repository.
func (Dev) InitDeviceRepo() error {
	return runGo("./cmd/ttn-lw-stack", "dr-db", "init")
}

// InitStack initializes the Stack.
func (Dev) InitStack() error {
	if mg.Verbose() {
		fmt.Println("Initializing the Stack")
	}
	if err := runGo("./cmd/ttn-lw-stack", "is-db", "init"); err != nil {
		return err
	}
	if err := runGo("./cmd/ttn-lw-stack", "is-db", "create-admin-user",
		"--id", "admin",
		"--email", "admin@example.com",
		"--password", "admin",
	); err != nil {
		return err
	}
	if err := runGo("./cmd/ttn-lw-stack", "is-db", "create-oauth-client",
		"--id", "cli",
		"--name", "Command Line Interface",
		"--owner", "admin",
		"--no-secret",
		"--redirect-uri", "local-callback",
		"--redirect-uri", "code",
	); err != nil {
		return err
	}
	if err := runGo("./cmd/ttn-lw-stack", "is-db", "create-oauth-client",
		"--id", "console",
		"--name", "Console",
		"--owner", "admin",
		"--secret", "console",
		"--redirect-uri", "https://localhost:8885/console/oauth/callback",
		"--redirect-uri", "http://localhost:1885/console/oauth/callback",
		"--redirect-uri", "/console/oauth/callback",
		"--logout-redirect-uri", "https://localhost:8885/console",
		"--logout-redirect-uri", "http://localhost:1885/console",
		"--logout-redirect-uri", "/console",
	); err != nil {
		return err
	}
	var key ttnpb.APIKey
	var jsonVal []byte
	var err error
	if jsonVal, err = outputJSONGo("run", "./cmd/ttn-lw-stack", "is-db", "create-user-api-key",
		"--user-id", "admin",
		"--name", "Admin User API Key",
	); err != nil {
		return err
	}
	if err := json.Unmarshal(jsonVal, &key); err != nil {
		return err
	}
	if err := writeToFile(filepath.Join(devDir, "admin_api_key.txt"), []byte(key.Key)); err != nil {
		return err
	}
	return nil
}

// StartDevStack starts TTS in end-to-end test configuration.
func (Dev) StartDevStack() error {
	os.Setenv("TTN_LW_IS_DATABASE_URI", databaseURI)
	os.Setenv("TTN_LW_IS_ADMIN_RIGHTS_ALL", "true")
	if mg.Verbose() {
		fmt.Println("Starting the Stack")
	}
	if err := os.MkdirAll(".cache", 0755); err != nil {
		return err
	}
	logFile, err := os.OpenFile(filepath.Join(".cache", "devStack.log"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	return execGo(logFile, logFile, "run", "./cmd/ttn-lw-stack", "start", "--log.format=json")
}

func init() {
	initDeps = append(initDeps, Dev.Certificates, Dev.InitDeviceRepo)
}
