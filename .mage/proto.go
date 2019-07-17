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
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"golang.org/x/xerrors"
)

const (
	protocName    = "thethingsindustries/protoc"
	protocVersion = "3.1.7"

	protocOut = "/out"
)

// Proto namespace.
type Proto mg.Namespace

// Image pulls the proto image.
func (Proto) Image(context.Context) error {
	out, err := sh.Output("docker", "images", "-q", fmt.Sprintf("%s:%s", protocName, protocVersion))
	if err != nil {
		return xerrors.Errorf("failed to query docker images: %s", err)
	}
	if len(out) > 0 {
		return nil
	}
	return sh.Run("docker", "pull", fmt.Sprintf("%s:%s", protocName, protocVersion))
}

type protocContext struct {
	WorkingDirectory string
	UID, GID         string
}

func makeProtoc() (func(...string) error, *protocContext, error) {
	mg.Deps(Proto.Image)

	wd, err := os.Getwd()
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to get working directory: %w", err)
	}
	usr, err := user.Current()
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to get user: %w", err)
	}
	return sh.RunCmd("docker", "run",
			"--rm",
			"--user", fmt.Sprintf("%s:%s", usr.Uid, usr.Gid),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s/api", filepath.Join(wd, "api"), wd),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s/doc", filepath.Join(wd, "doc"), wd),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s/go.thethings.network/lorawan-stack/pkg/ttnpb", filepath.Join(wd, "pkg", "ttnpb"), protocOut),
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s/sdk/js", filepath.Join(wd, "sdk", "js"), wd),
			"-w", wd,
			fmt.Sprintf("%s:%s", protocName, protocVersion),
			fmt.Sprintf("-I%s", filepath.Dir(wd)),
		), &protocContext{
			WorkingDirectory: wd,
			UID:              usr.Uid,
			GID:              usr.Gid,
		}, nil
}

func withProtoc(f func(pCtx *protocContext, protoc func(...string) error) error) error {
	protoc, pCtx, err := makeProtoc()
	if err != nil {
		return xerrors.New("failed to construct protoc command")
	}
	return f(pCtx, protoc)
}

// Go generates Go protos.
func (p Proto) Go(context.Context) error {
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
			return xerrors.Errorf("failed to generate protos: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := sh.RunV(filepath.Join(".mage", "scripts", "fix-grpc-gateway-names.sh"), "api"); err != nil {
		return xerrors.Errorf("failed to fix gRPC-gateway names: %w", err)
	}

	ttnpb, err := filepath.Abs(filepath.Join("pkg", "ttnpb"))
	if err != nil {
		return xerrors.Errorf("failed to construct absolute path to pkg/ttnpb: %w", err)
	}
	if err := execGo("run", "golang.org/x/tools/cmd/goimports", "-w", ttnpb); err != nil {
		return xerrors.Errorf("failed to run goimports on generated code: %w", err)
	}
	if err := execGo("run", "github.com/mdempsky/unconvert", "-apply", ttnpb); err != nil {
		return xerrors.Errorf("failed to run unconvert on generated code: %w", err)
	}
	return sh.RunV("gofmt", "-w", "-s", ttnpb)
}

// GoClean removes generated Go protos.
func (p Proto) GoClean(context.Context) error {
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

// Swagger generates Swagger protos.
func (p Proto) Swagger(context.Context) error {
	changed, err := target.Glob(filepath.Join("api", "api.swagger.json"), filepath.Join("api", "*.proto"))
	if err != nil {
		return xerrors.Errorf("failed checking modtime: %w", err)
	}
	if !changed {
		return nil
	}
	return withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--swagger_out=allow_merge,merge_file_name=api:%s/api", pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return xerrors.Errorf("failed to generate protos: %w", err)
		}
		return nil
	})
}

// SwaggerClean removes generated Swagger protos.
func (p Proto) SwaggerClean(context.Context) error {
	return sh.Rm(filepath.Join("api", "api.swagger.json"))
}

type HugoPage struct {
	Title       string     `yaml:title`
	Description string     `yaml:description`
	Tags        []string   `yaml:tags`
	SubPage     []HugoPage `yaml:subpage`
	Files       []string   `yaml:files`
	Template    string     `yaml:template`
}

// Markdown generates Markdown protos.
func (p Proto) Markdown(context.Context) error {
	changed, err := target.Glob(filepath.Join("api", "api.md"), filepath.Join("api", "*.proto"))
	if err != nil {
		return xerrors.Errorf("failed checking modtime: %w", err)
	}
	if !changed {
		fmt.Println("nothing changed")
		return nil
	}
	buff, err := ioutil.ReadFile("./.mage/proto_ref_tree.yaml")
	if err != nil {
		return err
	}
	var root HugoPage
	err = yaml.Unmarshal(buff, &root)
	if err != nil {
		return err
	}
	return markdownGenerateTree(strings.ReplaceAll(strings.ToLower(root.Title), " ", "-"), &root)
}

func markdownGenerateTree(path string, root *HugoPage) error {
	for _, page := range root.SubPage {
		if page.Template == "" || page.Files == nil {
			path = filepath.Join(path, strings.ReplaceAll(strings.ToLower(page.Title), " ", "-"))
			if err := markdownGenerateTree(path, &page); err != nil {
				return err
			}
			continue
		}
		if err := withProtoc(markdownProtoc(path, &page)); err != nil {
			return err
		}
	}
	return nil
}

func markdownProtoc(path string, page *HugoPage) func(pCtx *protocContext, protoc func(...string) error) error {
	return func(pCtx *protocContext, protoc func(...string) error) error {
		err := protoc(
			fmt.Sprintf("--doc_opt=%s/api/%s,%s.md --doc_out=%s/doc/content/reference/%s",
				pCtx.WorkingDirectory,
				page.Template,
				strings.ToLower(page.Title),
				pCtx.WorkingDirectory,
				path),
			fmt.Sprintf("%s/api/*.proto", pCtx.WorkingDirectory),
		)
		if err != nil {
			return xerrors.Errorf("failed to generate protos for %v : %w", page, err)
		}
		return nil
	}
}

// MarkdownClean removes generated Markdown protos.
func (p Proto) MarkdownClean(context.Context) error {
	var root HugoPage
	buff, err := ioutil.ReadFile("./.mage/proto_ref_tree.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(buff, &root)
	if err != nil {
		return err
	}
	return markdownCleanTree(strings.ReplaceAll(strings.ToLower(root.Title), " ", "-"), &root)
}

func markdownCleanTree(rootPath string, root *HugoPage) error {
	for _, page := range root.SubPage {
		path := filepath.Join(rootPath, strings.ReplaceAll(strings.ToLower(page.Title), " ", "-"))
		if page.Template == "" || page.Files == nil {
			if err := markdownCleanTree(path, &page); err != nil {
				return err
			}
			continue
		}
		err := sh.Rm(filepath.Join("doc/content/references/api", path))
		if err != nil {
			return err
		}
	}
	return nil
}

// JsSDK generates javascript SDK protos.
func (p Proto) JsSDK(context.Context) error {
	changed, err := target.Glob(filepath.Join("sdk", "js", "generated", "api.json"), filepath.Join("api", "*.proto"))
	if err != nil {
		return xerrors.Errorf("failed checking modtime: %w", err)
	}
	if !changed {
		return nil
	}
	return withProtoc(func(pCtx *protocContext, protoc func(...string) error) error {
		if err := protoc(
			fmt.Sprintf("--doc_opt=json,api.json --doc_out=%s/sdk/js/generated", pCtx.WorkingDirectory),
			fmt.Sprintf("%s/api/*.proto", pCtx.WorkingDirectory),
		); err != nil {
			return xerrors.Errorf("failed to generate protos: %w", err)
		}
		return nil
	})
}

// JsSDKClean removes generated javascript SDK protos.
func (p Proto) JsSDKClean(context.Context) error {
	return sh.Rm(filepath.Join("sdk", "js", "generated", "api.json"))
}

// All generates protos.
func (p Proto) All(ctx context.Context) {
	mg.CtxDeps(ctx, Proto.Go, Proto.Swagger, Proto.Markdown, Proto.JsSDK)
}

// Clean removes generated protos.
func (p Proto) Clean(ctx context.Context) {
	mg.CtxDeps(ctx, Proto.GoClean, Proto.SwaggerClean, Proto.MarkdownClean, Proto.JsSDKClean)
}
