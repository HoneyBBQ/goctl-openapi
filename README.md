goctl-openapi
===

> **Fork from**：[github.com/jayvynl/goctl-openapi](https://github.com/jayvynl/goctl-openapi)

This project is a plugin for [goctl](https://github.com/zeromicro/go-zero/tree/master/tools/goctl). It's able to generate [openapi specification version 3](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md) file from [go-ctl api](https://go-zero.dev/en/docs/tutorials) file.


### Features

- generate correct schema for any level of embedded structure type.
- generate correct schema for complicated type definition like `map[string][]map[int][]*Author`.
- parse parameter constraints from [validate](https://github.com/go-playground/validator) tag.
- support unified error response type for consistent error handling


### Install

This plugin's version and goctl's version should have the same major and minor version, it's recommended to install the matching version. If versions doesn't match, it may not work properly.

For example, if you use goctl v1.6.3, then you should install this plugin with:

```shell
go install github.com/honeybbq/goctl-openapi@latest
```

### Usage

Help messages.

```bash
Usage goctl-openapi:
  -errorType string
        specify error response type name in api file or json structure for unified error handling.
  -filename string
        openapi file name, default "openapi.json", "-" will output to stdout.
  -format string
        serialization format, "json" or "yaml", default "json".
  -pretty
        pretty print of json.
  -version
        show version and exit.
```

Usage example.

```shell
goctl api plugin -plugin 'goctl-openapi -pretty -format yaml -filename exmaple.yaml' -api example.api -dir docs 
```

Take the api file from [example](https://github.com/honeybbq/goctl-openapi/blob/main/example/example.api), [the generated openapi file](https://github.com/honeybbq/goctl-openapi/blob/main/example/openapi.json) can be visualized by [swagger editor](https://editor.swagger.io/?url=https://raw.githubusercontent.com/jayvynl/goctl-openapi/main/example/openapi.json).

### Error Response Handling

You can use the `-errorType` parameter to set a unified error response format for all APIs. This can be specified in two ways:

1. Using a type name defined in the API file:

```shell
goctl api plugin -p goctl-openapi -api user.api -dir . -errorType ErrorResponse
```

Where `ErrorResponse` is a type defined in the API file:

```go
type ErrorResponse {
    Code int `json:"code"`
    Message string `json:"message"`
}
```

2. Using a JSON string:

```shell
goctl api plugin -p goctl-openapi -api user.api -dir . -errorType '{"code":400,"message":"Error message"}'
```

When the unified error response is set, the generated OpenAPI document will include response definitions for common error status codes, including:

- All standard 4xx client error status codes (400-418, 421-426, 428-429, 431, 451)
- All standard 5xx server error status codes (500-508, 510-511)

This helps API consumers better understand possible error situations and handle them appropriately.

#### Specifying Error Codes for Specific Endpoints

If you want different endpoints to have different error codes, you can specify them in the API file using the `errors` property in the `@doc` annotation:

```go
@doc (
    summary: "Create User"
    description: "Create a new user and return user information"
    errors: "400,409,500"  // Only define these three error codes for this endpoint
)
@handler CreateUser
post /users (CreateUserRequest) returns (CreateUserResponse)

@doc (
    summary: "Get User Information"
    description: "Get user information by user ID"
    errors: "401,404"  // This endpoint only defines two error codes
)
@handler GetUser
get /users/:id returns (UserResponse)
```

This way, you can define the most appropriate set of error codes for each endpoint, making your API documentation more precise.

#### Adding Descriptions for Error Codes

You can also add custom description text for each error code using properties in the format `error{status code}`:

```go
@doc (
    summary: "Create User"
    description: "Create a new user and return user information"
    errors: "400,409,500"
    error400: "Invalid request parameters"
    error409: "Username already taken"
    error500: "Internal server error"
)
@handler CreateUser
post /users (CreateUserRequest) returns (CreateUserResponse)
```

This provides descriptive text for each error code in the generated OpenAPI document, helping API consumers understand possible error situations.

If an endpoint doesn't define a specific `errors` property but the `-errorType` parameter is set, the default set of error codes will be used.
