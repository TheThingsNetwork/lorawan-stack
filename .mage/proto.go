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
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

const protocVersion = "3.1.3"

const protocOut = "/out"

// Proto namespace.
type Proto mg.Namespace

type protocContext struct {
	WorkingDirectory string
	Uid, Gid         string
}

func makeProtoc() (func(...string) error, *protocContext, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, nil, errors.New("failed to get working directory").WithCause(err)
	}
	usr, err := user.Current()
	if err != nil {
		return nil, nil, errors.New("failed to get user").WithCause(err)
	}
	return sh.RunCmd("docker", "run",
			"--rm",
			"--user", fmt.Sprintf("%s:%s", usr.Uid, usr.Gid),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s/api", filepath.Join(wd, "api"), wd),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s/go.thethings.network/lorawan-stack/pkg/ttnpb", filepath.Join(wd, "pkg", "ttnpb"), protocOut),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s/sdk/js", filepath.Join(wd, "sdk", "js"), wd),
			"-w", wd,
			fmt.Sprintf("thethingsindustries/protoc:%s", protocVersion),
			fmt.Sprintf("-I%s", filepath.Dir(wd)),
		), &protocContext{
			WorkingDirectory: wd,
			Uid:              usr.Uid,
			Gid:              usr.Gid,
		}, nil
}

func withProtoc(f func(pCtx *protocContext, protoc func(...string) error) error) error {
	protoc, pCtx, err := makeProtoc()
	if err != nil {
		return errors.New("failed to construct protoc command")
	}
	return f(pCtx, protoc)
}

func (p Proto) Go(ctx context.Context) error {
	if err := withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		var convs []string
		for _, t := range []string{"any", "duration", "empty", "field_mask", "struct", "timestamp", "wrappers"} {
			convs = append(convs, fmt.Sprintf("Mgoogle/protobuf/%s.proto=github.com/gogo/protobuf/types", t))
		}
		convStr := strings.Join(convs, ",")

		if err := protoc(
			fmt.Sprintf("--fieldmask_out=lang=gogo,%s:%s", convStr, protocOut),
			fmt.Sprintf("--gogottn_out=plugins=grpc,%s:%s", convStr, protocOut),
			fmt.Sprintf("--grpc-gateway_out=%s:%s", convStr, protocOut),
			fmt.Sprintf("%s/api/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return errors.New("failed to generate protos").WithCause(err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := sh.RunV(filepath.Join(".mage", "scripts", "fix-grpc-gateway-names.sh"), "api"); err != nil {
		return errors.New("failed to fix gRPC-gateway names").WithCause(err)
	}

	ttnpb, err := filepath.Abs(filepath.Join("pkg", "ttnpb"))
	if err != nil {
		return errors.New("failed to construct absolute path to pkg/ttnpb").WithCause(err)
	}
	if err := execGo("run", "golang.org/x/tools/cmd/goimports", "-w", ttnpb); err != nil {
		return errors.New("failed to run goimports on generated code").WithCause(err)
	}
	if err := execGo("run", "github.com/mdempsky/unconvert", "-apply", ttnpb); err != nil {
		return errors.New("failed to run unconvert on generated code").WithCause(err)
	}
	return sh.RunV("gofmt", "-w", "-s", ttnpb)
}

func (p Proto) GoClean(ctx context.Context) error {
	return filepath.Walk(filepath.Join("pkg", "ttnpb"), func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, ext := range []string{".pb.go", ".pb.gw.go", ".pb.fm.go", ".pb.util.go"} {
			if strings.HasSuffix(path, ext) {
				if err := sh.Rm(path); err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	})
}

func (p Proto) Swagger(ctx context.Context) error {
	return withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--swagger_out=allow_merge,merge_file_name=api:%s/api", pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return errors.New("failed to generate protos").WithCause(err)
		}
		return nil
	})
}

func (p Proto) SwaggerClean(ctx context.Context) error {
	return sh.Rm(filepath.Join("api", "api.swagger.json"))
}

func (p Proto) Markdown(ctx context.Context) error {
	return withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--doc_opt=%s/api/api.md.tmpl,api.md --doc_out=%s/api", pCtx.WorkingDirectory, pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return errors.New("failed to generate protos").WithCause(err)
		}
		return nil
	})
}

func (p Proto) MarkdownClean(ctx context.Context) error {
	return sh.Rm(filepath.Join("api", "api.md"))
}

func (p Proto) JsSDK(ctx context.Context) error {
	if err := withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--doc_opt=json,api.json --doc_out=%s/sdk/js/generated", pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return errors.New("failed to generate protos").WithCause(err)
		}
		return nil
	}); err != nil {
		return err
	}
	yarn, err := yarn()
	if err != nil {
		return errors.New("failed to construct yarn command").WithCause(err)
	}
	if err := yarn("--cwd=./sdk/js", "run", "definitions"); err != nil {
		return errors.New("failed to generate definitions").WithCause(err)
	}
	return nil
}

func (p Proto) JsSDKClean(ctx context.Context) error {
	for _, f := range []string{"api.json", "api-definition.json"} {
		if err := sh.Rm(filepath.Join("sdk", "js", "generated", f)); err != nil {
			return err
		}
	}
	return nil
}

func (p Proto) All(ctx context.Context) {
	mg.CtxDeps(ctx, p.GoClean, p.SwaggerClean, p.MarkdownClean, p.JsSDKClean)
	mg.CtxDeps(ctx, p.Go, p.Swagger, p.Markdown, p.JsSDK)
}
