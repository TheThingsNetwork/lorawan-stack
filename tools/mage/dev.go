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
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
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
	sqlDatabase           = "postgres"
	redisDatabase         = "redis"
	devDatabases          = []string{sqlDatabase, redisDatabase}
	devDataDir            = ".env/data"
	devDir                = ".env"
	devDatabaseName       = "ttn_lorawan_dev"
	devDockerComposeFlags = []string{"compose", "-p", "lorawan-stack-dev"}
	databaseURI           = fmt.Sprintf("postgresql://root:root@localhost:5432/%s?sslmode=disable", devDatabaseName)
	testDatabaseNames     = []string{"ttn_lorawan_is_test", "ttn_lorawan_is_store_test"}
)

func dockerComposeFlags(args ...string) []string {
	return append(devDockerComposeFlags, args...)
}

func execDockerCompose(args ...string) error {
	_, err := sh.Exec(nil, os.Stdout, os.Stderr, "docker", dockerComposeFlags(args...)...)
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

// SQLDump performs an SQL database dump of the dev database to the .cache folder.
func (Dev) SQLDump() error {
	if mg.Verbose() {
		fmt.Println("Saving sql database dump")
	}
	if err := os.MkdirAll(path.Join(".env", "cache"), 0o755); err != nil {
		return err
	}
	return execDockerCompose("exec", "-T", "postgres",
		"pg_dump", "-Fc", "-f", "/var/lib/ttn-lorawan/cache/database.pgdump", devDatabaseName,
	)
}

// SQLRestore restores the dev database using a previously generated dump.
func (Dev) SQLRestore() error {
	if mg.Verbose() {
		fmt.Println("Restoring database from dump")
	}
	d := filepath.Join(".env", "cache", "database.pgdump")
	if _, err := os.Stat(d); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Dumpfile does not exist: %w", d)
	}
	return execDockerCompose("exec", "-T", "postgres", "/bin/bash", "-c",
		strings.Join([]string{
			fmt.Sprintf("dropdb --if-exists --force %s", devDatabaseName),
			fmt.Sprintf("createdb %s", devDatabaseName),
			fmt.Sprintf("pg_restore -d %s -Fc /var/lib/ttn-lorawan/cache/database.pgdump", devDatabaseName),
		}, " && "),
	)
}

// RedisFlush deletes all keys from redis.
func (Dev) RedisFlush() error {
	if mg.Verbose() {
		fmt.Println("Deleting all keys from redis")
	}

	keys, err := sh.Output("docker", dockerComposeFlags("exec", "-T", "redis", "redis-cli", "keys", "ttn:v3:*")...)
	if err != nil {
		return err
	}
	ks := strings.Split(keys, "\n")
	if len(ks) == 0 {
		return nil
	}
	flags := dockerComposeFlags(append([]string{"exec", "-T", "redis", "redis-cli", "del"}, ks...)...)
	_, err = sh.Exec(nil, nil, os.Stderr, "docker", flags...)
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

	// When TimescaleDB starts for the first time, it restarts Postgres after initialization.
	// Therefore, pg_isready may return a successful exit code during initialization, while the database shuts down
	// shortly after. Therefore, the ready check goes in two cycles and only returns after 10 successive ready
	// indications.
	if mg.Verbose() {
		fmt.Println("Waiting for Postgres to be ready")
	}
	flags := dockerComposeFlags("exec", "-T", "postgres", "pg_isready")
nextCycle:
	for i := 0; i < 2; i++ {
		var (
			wasReady        bool
			successiveReady int
		)
		for j := 0; j < 30; j++ {
			time.Sleep(time.Second)
			_, err := sh.Exec(nil, nil, nil, "docker", flags...)
			isReady := err == nil
			switch {
			case wasReady && !isReady:
				if mg.Verbose() {
					fmt.Println("Postgres is not ready anymore")
				}
				continue nextCycle
			case !wasReady && isReady:
				if mg.Verbose() {
					fmt.Println("Postgres is ready, checking if it stays ready")
				}
				successiveReady = 1
			case wasReady && isReady:
				successiveReady++
				if successiveReady == 10 {
					if mg.Verbose() {
						fmt.Println("Postgres ready state seems stable")
					}
					return nil
				}
			}
			wasReady = isReady
		}
		return errors.New("No ready indication within 30 checks")
	}
	return errors.New("Postgres is not ready")
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
	return execDockerCompose("exec", "postgres", "psql", devDatabaseName)
}

// DBCreate creates the SQL databases used for unit tests.
func (Dev) DBCreate() error {
	mg.Deps(Dev.DBStart)
	if mg.Verbose() {
		fmt.Println("Creating dev databases")
	}
	for _, db := range testDatabaseNames {
		if err := execDockerCompose("exec", "postgres", "psql", devDatabaseName, "-c", fmt.Sprintf("CREATE DATABASE %s;", db)); err != nil {
			return err
		}
	}
	return nil
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
	if err := runGo("./cmd/ttn-lw-stack", "is-db", "migrate"); err != nil {
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
	if err := jsonpb.TTN().Unmarshal(jsonVal, &key); err != nil {
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
	if err := os.MkdirAll(".cache", 0o755); err != nil {
		return err
	}
	logFile, err := os.OpenFile(filepath.Join(".cache", "devStack.log"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	if os.Getenv("CI") == "true" {
		return execFrom(
			"",
			map[string]string{},
			logFile,
			logFile,
			"./ttn-lw-stack",
			"start",
			"--log.format=json",
			"--config=config/stack/ttn-lw-stack-tls.yml",
		)
	}
	return execGo(logFile, logFile, "run", "./cmd/ttn-lw-stack", "start", "--log.format=json")
}

func (Dev) ServeDevWebui() error {
	if mg.Verbose() {
		fmt.Println("Starting the webui with interactive configs")
	}
	return sh.RunV("node", "tools/js/serve-dev-webui.js")
}

func init() {
	initDeps = append(initDeps, Dev.Certificates, Dev.InitDeviceRepo)
}

func execFrom(
	dir string, env map[string]string, stdout, stderr io.Writer, cmd string, args ...string,
) error {
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
	}
	c.Dir = dir
	c.Stderr = stderr
	c.Stdout = stdout
	c.Stdin = os.Stdin
	if mg.Verbose() {
		log.Println("exec:", cmd, strings.Join(args, " "))
	}
	err := c.Run()
	if err == nil {
		return nil
	}
	ran, code := sh.CmdRan(err), sh.ExitStatus(err)
	if ran {
		return mg.Fatalf(code, `running "%s %s" failed with exit code %d`, cmd, strings.Join(args, " "), code)
	}
	return fmt.Errorf(`failed to run "%s %s: %w"`, cmd, strings.Join(args, " "), err)
}
