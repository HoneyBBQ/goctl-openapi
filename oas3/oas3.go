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

package oas3

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/plugin"
)

var DefaultResponseDesc = "Successful response."
var DefaultErrorDesc = "Error response."

// Define a default error response structure
type DefaultErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func GetDoc(p *plugin.Plugin, errorType string) (*openapi3.T, error) {
	doc := &openapi3.T{
		OpenAPI:      "3.0.3",
		Components:   newComponents(),
		Info:         getInfo(p.Api.Info.Properties),
		Paths:        openapi3.NewPaths(),
		Servers:      getServers(p.Api.Info.Properties),
		ExternalDocs: getExternalDocs(p.Api.Info.Properties),
	}

	// 所有安全方案都在扫描API定义时动态添加

	types := make(map[string]spec.DefineStruct) // all defined types from api spec
	for _, typ := range p.Api.Types {
		if ds, ok := typ.(spec.DefineStruct); ok {
			types[ds.Name()] = ds
		}
	}

	// Process unified error response body
	errorSchema := createErrorResponseSchema(errorType, types, doc.Components.Schemas)
	if errorSchema != nil {
		// Add error response to components
		doc.Components.Responses["Error"] = &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: &DefaultErrorDesc,
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{
						Schema: errorSchema,
					},
				},
			},
		}
	}

	fillPaths(p, doc, types, doc.Components.RequestBodies, doc.Components.Responses, doc.Components.Schemas, errorSchema)
	return doc, nil
}

// Create error response schema
func createErrorResponseSchema(errorType string, types map[string]spec.DefineStruct, schemas openapi3.Schemas) *openapi3.SchemaRef {
	if errorType == "" {
		// If no error type is provided, return default error structure
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: TypeObject,
				Properties: openapi3.Schemas{
					"code": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    TypeInteger,
							Example: 400,
						},
					},
					"message": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:    TypeString,
							Example: "Error message",
						},
					},
				},
				Required: []string{"code", "message"},
			},
		}
	}

	// Try to parse error type as JSON
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal([]byte(errorType), &jsonSchema); err == nil {
		// Successfully parsed JSON, create corresponding schema
		return createSchemaFromJSON(jsonSchema)
	}

	// Not JSON, try to find the type in API definition
	if ds, ok := types[errorType]; ok {
		return getStructSchema(ds, types, schemas)
	}

	// If specified type is not found, return default error structure
	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: TypeObject,
			Properties: openapi3.Schemas{
				"code": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    TypeInteger,
						Example: 400,
					},
				},
				"message": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    TypeString,
						Example: "Error message",
					},
				},
			},
			Required: []string{"code", "message"},
		},
	}
}

// Create schema from JSON
func createSchemaFromJSON(jsonObj map[string]interface{}) *openapi3.SchemaRef {
	schema := &openapi3.Schema{
		Type:       TypeObject,
		Properties: make(openapi3.Schemas),
	}

	for key, value := range jsonObj {
		var propSchema *openapi3.SchemaRef

		switch v := value.(type) {
		case string:
			propSchema = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:    TypeString,
					Example: v,
				},
			}
		case float64:
			// In JSON, all numbers are float64
			if v == float64(int(v)) {
				propSchema = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    TypeInteger,
						Example: int(v),
					},
				}
			} else {
				propSchema = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    TypeNumber,
						Example: v,
					},
				}
			}
		case bool:
			propSchema = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:    TypeBoolean,
					Example: v,
				},
			}
		case map[string]interface{}:
			propSchema = createSchemaFromJSON(v)
		case []interface{}:
			if len(v) > 0 {
				// Get array element type
				itemSchema := createItemSchemaFromJSON(v[0])
				propSchema = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:  TypeArray,
						Items: itemSchema,
					},
				}
			} else {
				propSchema = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:  TypeArray,
						Items: nil,
					},
				}
			}
		default:
			propSchema = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: TypeObject,
				},
			}
		}

		schema.Properties[key] = propSchema
		schema.Required = append(schema.Required, key)
	}

	return &openapi3.SchemaRef{Value: schema}
}

// 从JSON数组元素创建Schema
func createItemSchemaFromJSON(item interface{}) *openapi3.SchemaRef {
	switch v := item.(type) {
	case string:
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:    TypeString,
				Example: v,
			},
		}
	case float64:
		if v == float64(int(v)) {
			return &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:    TypeInteger,
					Example: int(v),
				},
			}
		} else {
			return &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:    TypeNumber,
					Example: v,
				},
			}
		}
	case bool:
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:    TypeBoolean,
				Example: v,
			},
		}
	case map[string]interface{}:
		return createSchemaFromJSON(v)
	case []interface{}:
		if len(v) > 0 {
			itemSchema := createItemSchemaFromJSON(v[0])
			return &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:  TypeArray,
					Items: itemSchema,
				},
			}
		} else {
			return &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: TypeArray,
				},
			}
		}
	default:
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: TypeObject,
			},
		}
	}
}

func newComponents() *openapi3.Components {
	return &openapi3.Components{
		Schemas:         make(openapi3.Schemas),
		Parameters:      make(openapi3.ParametersMap),
		Headers:         make(openapi3.Headers),
		RequestBodies:   make(openapi3.RequestBodies),
		Responses:       make(openapi3.ResponseBodies),
		SecuritySchemes: make(openapi3.SecuritySchemes),
		Examples:        make(openapi3.Examples),
		Links:           make(openapi3.Links),
		Callbacks:       make(openapi3.Callbacks),
	}
}

func fillPaths(
	p *plugin.Plugin,
	doc *openapi3.T,
	types map[string]spec.DefineStruct, // all defined types from api spec
	requests openapi3.RequestBodies, // request body references
	responses openapi3.ResponseBodies, // response body references
	schemas openapi3.Schemas, // schema references, json field of struct type will read and write this map
	errorSchema *openapi3.SchemaRef, // 错误返回体Schema
) {
	rp := newRequestParser()

	// 记录已经添加的安全方案
	addedSecuritySchemes := make(map[string]bool)

	service := p.Api.Service.JoinPrefix()
	for _, group := range service.Groups {
		// 检查组中是否使用了JWT认证
		if group.Annotation.Properties["jwt"] != "" {
			// 如果还没添加JWT安全方案，添加它
			if !addedSecuritySchemes["jwt"] {
				doc.Components.SecuritySchemes["jwt"] = &openapi3.SecuritySchemeRef{
					Value: openapi3.NewJWTSecurityScheme(),
				}
				addedSecuritySchemes["jwt"] = true
			}
		}

		// 检查组中是否使用了Cookie认证
		if cookieAuth := group.Annotation.Properties["cookie"]; cookieAuth != "" {
			if cookieAuth == "Auth" {
				// 使用默认cookie名称"session"
				if !addedSecuritySchemes["cookieAuth"] {
					doc.Components.SecuritySchemes["cookieAuth"] = createCookieSecurityScheme("session")
					addedSecuritySchemes["cookieAuth"] = true
				}
			} else {
				// 使用自定义cookie名称
				cookieAuthName := "cookieAuth_" + cookieAuth
				if !addedSecuritySchemes[cookieAuthName] {
					doc.Components.SecuritySchemes[cookieAuthName] = createCookieSecurityScheme(cookieAuth)
					addedSecuritySchemes[cookieAuthName] = true
				}
			}
		}

		for _, route := range group.Routes {
			method := strings.ToUpper(route.Method)
			hasBody := method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch

			tags := getTags(route.AtDoc.Properties)
			if len(tags) == 0 {
				if gn := group.GetAnnotation("group"); len(gn) > 0 {
					tags = []string{gn}
				} else {
					tags = []string{service.Name}
				}
			}

			summary := GetProperty(route.AtDoc.Properties, "summary")
			if summary == "" {
				summary = route.AtDoc.Text
			}
			desc := GetProperty(route.AtDoc.Properties, "description")
			if desc == "" {
				desc = strings.Join(route.Docs, " ")
			}

			var (
				params   openapi3.Parameters
				request  *openapi3.RequestBodyRef
				response *openapi3.ResponseRef
			)
			if typ, ok := route.RequestType.(spec.DefineStruct); ok {
				params, request = rp.Parse(typ, types, requests, schemas)
				if !hasBody {
					request = nil
				}
			}

			var responseTypeName string
			if route.ResponseType != nil {
				responseTypeName = route.ResponseType.Name()
			}

			if responseTypeName == "" {
				response = &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: &DefaultResponseDesc,
					},
				}
			} else {
				response = parseResponse(responseTypeName, types, responses, schemas)
			}

			var security *openapi3.SecurityRequirements
			if jwtAuth := group.Annotation.Properties["jwt"]; jwtAuth != "" {
				security = &openapi3.SecurityRequirements{{"jwt": []string{}}}
			}

			// Check for cookie authentication (cookie: Auth 格式)
			if cookieAuth := group.Annotation.Properties["cookie"]; cookieAuth != "" {
				if cookieAuth == "Auth" {
					// 使用默认cookie名称"session"
					security = &openapi3.SecurityRequirements{{"cookieAuth": []string{}}}
				} else {
					// 使用自定义cookie名称(格式: cookie: 自定义名称)
					cookieAuthName := "cookieAuth_" + cookieAuth
					security = &openapi3.SecurityRequirements{{cookieAuthName: []string{}}}
				}
			}

			var servers *openapi3.Servers
			if ss := getServers(route.AtDoc.Properties); len(ss) > 0 {
				servers = &ss
			}

			// Collect all response options
			responseOptions := []openapi3.NewResponsesOption{
				openapi3.WithStatus(http.StatusOK, response),
			}

			// Check if specific error codes are defined in the API documentation
			errorsProperty := GetProperty(route.AtDoc.Properties, "errors")

			// Store already added error codes to avoid duplicates
			processedErrorCodes := make(map[int]bool)

			// Add error responses
			if errorSchema != nil {
				errorResponse := &openapi3.ResponseRef{
					Ref: "#/components/responses/Error",
				}

				// If specific error codes are defined in @doc annotation, use only these error codes
				if errorsProperty != "" {
					// Parse error codes list, format like: "400,401,404,500"
					errorCodes := strings.Split(errorsProperty, ",")

					// First add all specified error codes
					for _, code := range errorCodes {
						code = strings.TrimSpace(code)
						statusCode, err := strconv.Atoi(code)
						if err == nil && statusCode >= 100 && statusCode < 600 {
							if !processedErrorCodes[statusCode] {
								responseOptions = append(responseOptions,
									openapi3.WithStatus(statusCode, errorResponse))
								processedErrorCodes[statusCode] = true
							}
						}
					}

					// Then process descriptions for specific error codes
					for _, code := range errorCodes {
						code = strings.TrimSpace(code)
						errorDescKey := "error" + code
						errorDesc := GetProperty(route.AtDoc.Properties, errorDescKey)

						if errorDesc != "" {
							statusCode, err := strconv.Atoi(code)
							if err == nil && statusCode >= 100 && statusCode < 600 {
								// Create error response with custom description
								customDesc := errorDesc
								customErrorResponse := &openapi3.ResponseRef{
									Value: &openapi3.Response{
										Description: &customDesc,
										Content: openapi3.Content{
											"application/json": &openapi3.MediaType{
												Schema: errorSchema,
											},
										},
									},
								}

								// Override with custom description for the same status code
								responseOptions = append(responseOptions,
									openapi3.WithStatus(statusCode, customErrorResponse))
								processedErrorCodes[statusCode] = true
							}
						}
					}
				} else {
					// 如果没有定义特定的错误码，则添加常见的错误状态码
					// 常见的客户端错误
					clientErrors := []int{400, 401, 403, 404, 422}
					for _, code := range clientErrors {
						if !processedErrorCodes[code] {
							responseOptions = append(responseOptions,
								openapi3.WithStatus(code, errorResponse))
							processedErrorCodes[code] = true
						}
					}

					// 常见的服务器错误
					serverErrors := []int{500, 502, 503, 504}
					for _, code := range serverErrors {
						if !processedErrorCodes[code] {
							responseOptions = append(responseOptions,
								openapi3.WithStatus(code, errorResponse))
							processedErrorCodes[code] = true
						}
					}
				}
			}

			// Create response object
			responses := openapi3.NewResponses(responseOptions...)

			doc.AddOperation(
				ConvertPath(route.Path),
				method,
				&openapi3.Operation{
					Tags:         tags,
					Summary:      summary,
					Description:  desc,
					OperationID:  route.Handler,
					Parameters:   params,
					RequestBody:  request,
					Responses:    responses,
					Security:     security,
					Servers:      servers,
					ExternalDocs: getExternalDocs(route.AtDoc.Properties),
				},
			)
		}
	}
}

// createCookieSecurityScheme 创建基于Cookie的API密钥认证方案
// 根据OpenAPI 3.0规范，Cookie认证是apiKey类型的一种特殊形式
// 参数:
//   - cookieName: Cookie的名称，当客户端请求API时需要在Cookie中包含该名称的值
//
// 返回:
//   - 一个OpenAPI安全方案引用，类型为apiKey，位置为cookie
func createCookieSecurityScheme(cookieName string) *openapi3.SecuritySchemeRef {
	return &openapi3.SecuritySchemeRef{
		Value: &openapi3.SecurityScheme{
			Type:        "apiKey",
			In:          "cookie",
			Name:        cookieName,
			Description: "API key authentication using HTTP cookie",
		},
	}
}
