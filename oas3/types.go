// Copyright 2023 The goctl-openapi Authors
// Fork from https://github.com/jayvynl/goctl-openapi
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

package oas3

import "github.com/getkin/kin-openapi/openapi3"

// TypePointer 辅助函数，将类型转换为 *openapi3.Types
func TypePointer(t string) *openapi3.Types {
	types := openapi3.Types([]string{t})
	return &types
}

// Common type constants for convenience
var (
	TypeObject  = TypePointer(openapi3.TypeObject)
	TypeArray   = TypePointer(openapi3.TypeArray)
	TypeString  = TypePointer(openapi3.TypeString)
	TypeNumber  = TypePointer(openapi3.TypeNumber)
	TypeInteger = TypePointer(openapi3.TypeInteger)
	TypeBoolean = TypePointer(openapi3.TypeBoolean)
)
