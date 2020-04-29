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

package commands

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/discover"
	"gopkg.in/yaml.v2"
)

const configFileName = ".ttn-lw-cli.yml"

var (
	errNoHost     = errors.DefineInvalidArgument("no_host", "no host set")
	errFileExists = errors.DefineAlreadyExists("file_exists", "`{file}` exists")
	errFailWrite  = errors.DefinePermissionDenied("fail_write", "failed to write `{file}`")
	useCommand    = &cobra.Command{
		Use:               "use [host]",
		Short:             "Use",
		PersistentPreRunE: preRun(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errNoHost
			}
			insecure, _ := cmd.Flags().GetBool("insecure")
			fetchCA, _ := cmd.Flags().GetBool("fetch-ca")
			user, _ := cmd.Flags().GetBool("user")
			overwrite, _ := cmd.Flags().GetBool("overwrite")

			host := args[0]
			grpcPort, _ := cmd.Flags().GetInt("grpc-port")
			if grpcPort == 0 {
				grpcPort = discover.DefaultPorts[!insecure]
			}
			grpcServerAddress, err := discover.DefaultPort(host, grpcPort)
			if err != nil {
				return err
			}
			oauthServerAddress, _ := cmd.Flags().GetString("oauth-server-address")
			if oauthServerAddress == "" {
				oauthServerBaseAddress, err := discover.DefaultURL(host, discover.DefaultHTTPPorts[!insecure], !insecure)
				if err != nil {
					return err
				}
				oauthServerAddress = oauthServerBaseAddress + "/oauth"
			}
			conf := MakeDefaultConfig(grpcServerAddress, oauthServerAddress, insecure)
			conf.CredentialsID = host

			destPath := func(base string, user bool, overwrite bool) (string, error) {
				fileName := base
				if user {
					dir, err := os.UserConfigDir()
					if err != nil {
						return "", err
					}
					if err = os.MkdirAll(dir, 0755); err != nil {
						return "", err
					}
					fileName = filepath.Join(dir, base)
				}
				_, err := os.Stat(fileName)
				if !os.IsNotExist(err) && !overwrite {
					logger.Warnf("%s already exists. Use --overwrite", fileName)
					return "", errFileExists.WithAttributes("file", fileName)
				}
				return fileName, nil
			}

			// Get CA certificate from server
			if !insecure && fetchCA {
				h := md5.New()
				io.WriteString(h, host)
				caFileName := fmt.Sprintf("ca.%s.pem", hex.EncodeToString(h.Sum(nil))[:6])
				caFile, err := destPath(caFileName, user, overwrite)
				if err != nil {
					return err
				}

				logger.Infof("Will retrieve certificate from %s", conf.NetworkServerGRPCAddress)
				conn, err := tls.Dial("tcp", conf.NetworkServerGRPCAddress, &tls.Config{InsecureSkipVerify: true})
				if err != nil {
					return err
				}
				defer conn.Close()

				buf := &bytes.Buffer{}
				for _, cert := range conn.ConnectionState().PeerCertificates {
					if !cert.IsCA {
						continue
					}
					if err = pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
						logger.Warnf("Could not retrieve certificate: %s", err)
					}
				}
				if err = ioutil.WriteFile(caFile, buf.Bytes(), 0644); err != nil {
					return errFailWrite.WithCause(err).WithAttributes("file", caFile)
				}
				logger.Infof("CA file for %s written in %s", host, caFile)
				abs, err := filepath.Abs(caFile)
				if err != nil {
					return err
				}
				conf.CA = abs
			}

			b, err := yaml.Marshal(conf)
			if err != nil {
				return err
			}

			configFile, err := destPath(configFileName, user, overwrite)
			if err != nil {
				return err
			}

			if err = ioutil.WriteFile(configFile, b, 0644); err != nil {
				return errFailWrite.WithCause(err).WithAttributes("file", configFile)
			}
			logger.Infof("Config file for %s written in %s", host, configFile)
			return nil
		},
	}
)

func init() {
	useCommand.Flags().Bool("insecure", defaultInsecure, "Connect without TLS")
	useCommand.Flags().String("host", defaultClusterHost, "Server host name")
	useCommand.Flags().String("oauth-server-address", "", "OAuth Server address")
	useCommand.Flags().Bool("fetch-ca", false, "Connect to server and retrieve CA")
	useCommand.Flags().Bool("user", false, "Write config file in user config directory")
	useCommand.Flags().Bool("overwrite", false, "Overwrite existing config files")
	useCommand.Flags().Int("grpc-port", 0, "")
	Root.AddCommand(useCommand)
}
