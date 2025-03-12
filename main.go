// Copyright 2023 The goctl-openapi Authors
// Fork from https://github.com/honeybbq/goctl-openapi
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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/honeybbq/goctl-openapi/oas3"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
	"gopkg.in/yaml.v2"
)

const Version = "v1.6.0"

var (
	version   = flag.Bool("version", false, `show version and exit.`)
	output    = flag.String("filename", "", `openapi file name, default "openapi.json", "-" will output to stdout.`)
	format    = flag.String("format", "", `serialization format, "json" or "yaml", default "json".`)
	pretty    = flag.Bool("pretty", false, `pretty print of json.`)
	errorType = flag.String("errorType", "", `specify error response type name or json structure for unified error handling.`)
)

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("goctl-openapi %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
		return
	}

	p, err := plugin.NewPlugin()
	if err != nil {
		fmt.Printf("goctl-openapi: %s\n", err)
		return
	}

	var (
		o = "openapi"
		f = "json"
	)
	if *output != "" {
		o = *output
	}
	if strings.HasSuffix(o, ".json") {
		f = "json"
	} else if strings.HasSuffix(o, ".yml") || strings.HasSuffix(o, ".yaml") {
		f = "yaml"
	} else {
		if *format != "" {
			switch *format {
			case "json":
				f = "json"
			case "yaml", "yml":
				f = "yaml"
			default:
				fmt.Println("goctl-openapi: format must be json or yaml")
				return
			}
		}
		if o != "-" {
			o = fmt.Sprintf("%s.%s", o, f)
		}
	}

	var w *os.File
	if o == "-" {
		w = os.Stdout
	} else {
		w, err = os.Create(path.Join(p.Dir, o))
		if err != nil {
			fmt.Printf("goctl-openapi: %s\n", err)
			return
		}
		defer w.Close()
	}

	doc, err := oas3.GetDoc(p, *errorType)
	if err != nil {
		fmt.Printf("goctl-openapi: %s\n", err)
		return
	}

	if f == "json" {
		encoder := json.NewEncoder(w)
		if *pretty {
			encoder.SetIndent("", "  ")
		}
		err = encoder.Encode(doc)
		if err != nil {
			fmt.Printf("goctl-openapi: %s\n", err)
			return
		}
	} else {
		encoder := yaml.NewEncoder(w)
		defer encoder.Close()
		err = encoder.Encode(doc)
		if err != nil {
			fmt.Printf("goctl-openapi: %s\n", err)
			return
		}
	}
}
