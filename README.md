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
goctl api plugin -plugin 'goctl-openapi -pretty -format yaml -filename exmaple.yaml  -errorType ErrorResponse' -api example.api -dir docs 
```

Take the api file from [example](https://github.com/honeybbq/goctl-openapi/blob/main/example/example.api), [the generated openapi file](https://github.com/honeybbq/goctl-openapi/blob/main/example/openapi.json) can be visualized by [swagger editor](https://editor.swagger.io/?url=https://raw.githubusercontent.com/jayvynl/goctl-openapi/main/example/openapi.json).

### Error Response Handling

You can use the `-errorType` parameter to set a unified error response format for all APIs. This can be specified in two ways:

1. Using a type name defined in the API file:

```shell
goctl api plugin -p goctl-openapi -api user.api -dir .
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

#### Default Error Status Codes

If an endpoint doesn't define a specific `errors` property but the `-errorType` parameter is set, the following default error status codes will be automatically added:

- Client Errors: 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), 404 (Not Found), 422 (Unprocessable Entity)
- Server Errors: 500 (Internal Server Error), 502 (Bad Gateway), 503 (Service Unavailable), 504 (Gateway Timeout)

This ensures that your API documentation includes the most common error scenarios without requiring you to explicitly define them for each endpoint. You can still override these defaults by specifying your own set of error codes in the `errors` property.

Example:

```go
// This will use the default error status codes (400, 401, 403, 404, 422, 500, 502, 503, 504)
@handler GetUser
get /users/:id returns (UserResponse)

// This will only use the specified error codes (401, 404)
@doc (
    errors: "401,404"
)
@handler GetUserCustom
get /users/custom/:id returns (UserResponse)
```

### Authentication

The plugin supports multiple authentication mechanisms that align with the OpenAPI 3.0 security schemes. Security schemes are only added to the OpenAPI document when explicitly declared in your API definition file.

#### JWT Authentication

To enable JWT Bearer authentication for a group of API endpoints, use the `jwt` property in the `@server` annotation:

```go
@server (
    jwt:    Auth
    prefix: /api
    group:  example
)
service ExampleService {
    // Your API endpoints here...
}
```

This generates an OpenAPI security scheme of type `http` with scheme `bearer` and bearerFormat `JWT`.

#### Cookie-Session Authentication

For web applications using Cookie-Session authentication, use the `cookie` property:

```go
@server (
    cookie: Auth
    prefix: /api
    group:  example
)
service ExampleService {
    // Your API endpoints here...
}
```

This generates an OpenAPI security scheme of type `apiKey` with `in: cookie` and `name: session`.

#### Custom Cookie Name

You can also specify a custom cookie name for cookie-based authentication:

```go
@server (
    cookie: SessionID
    prefix: /api
    group:  example
)
service ExampleService {
    // Your API endpoints here...
}
```

This will use "SessionID" as the cookie name for authentication instead of the default "session".

#### Multiple Authentication Types

You can use different authentication methods for different endpoint groups within the same API:

```go
// Public endpoints with no authentication
@server (
    prefix: /api/public
    group:  public
)
service ExampleService {
    @handler Health
    get /health
}

// JWT authenticated endpoints
@server (
    jwt:    Auth
    prefix: /api/admin
    group:  admin
)
service ExampleService {
    @handler AdminData
    get /data
}

// Cookie authenticated endpoints
@server (
    cookie: Auth
    prefix: /api/user
    group:  user
)
service ExampleService {
    @handler UserProfile
    get /profile
}
```

The generated OpenAPI document will include only the security schemes that are actually used in your API definition, making the specification more accurate and concise.

#### Example Security Schemes in OpenAPI Output

For an API that uses both JWT and Cookie authentication, the generated OpenAPI JSON would include security schemes like this:

```json
{
  "components": {
    "securitySchemes": {
      "jwt": {
        "bearerFormat": "JWT",
        "scheme": "bearer",
        "type": "http"
      },
      "cookieAuth": {
        "description": "API key authentication using HTTP cookie",
        "in": "cookie",
        "name": "session",
        "type": "apiKey"
      }
    }
  }
}
```

#### Using the Generated API Documentation

The security schemes defined in the OpenAPI document are recognized by Swagger UI and other OpenAPI tools. This allows:

1. Frontend developers to understand the authentication requirements for each endpoint
2. API clients to be automatically generated with proper authentication handling
3. API testing tools like Postman to automatically include the required authentication headers/cookies

For testing Cookie authentication in Swagger UI, you'll need to:
1. Authenticate through your application's login endpoint first
2. Allow the browser to store the authentication cookie
3. Swagger UI will then automatically include this cookie in subsequent requests
