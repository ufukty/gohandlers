# gohandlers

gohandlers is a CLI tool that creates Go files contain information about handlers and their binding types that are defined in a directory. The produced files make the information previously embedded in the type names and discarded by compiler available to developer's use.

gohandlers allows defining the path and method information in one place and to get type-checked.

## Usage

```
gohandlers [command] <command args>

commands:
    bindings : Produces a file contains Build and Parse methods for binding types
    client   : Produces a file contains Client and its methods
    mock     : Produces a file contains Client mock and an interface
    list     : Produces a file contains ListHandlers function/methods
    yaml     : Produces a yaml file contains handler paths and methods for non-Go clients

Run to get help on specific command, where available:
    gohandlers [command] --help
```

## Commands

gohandlers provides multiple commands each produce a file.

### Listing handlers to register into a router with `list` command

gohandlers can generate a function which returns a `map` of handlers. This allows developer to register handlers to `http.ServeMux` (or a router of dev's choice) by simply iterating over it and call the router's method.

Collecting method and path information of all handlers from the handlers improves consistency over development time, enforcing the single source of truth principle.

The produced file gets named as `list.gh.go` by default. The file will contain `ListHandlers` function for handlers defined on the global scope. It will also contain as many `ListHandler` methods as the types which contains at least one method in the `http.HandlerFunc` signature.

```go
func ListHandlers() map[string]reception.HandlerInfo
func (p *Public) ListHandlers() map[string]reception.HandlerInfo
func (p *Private) ListHandlers() map[string]reception.HandlerInfo
```

### Implementing parser and builder methods on bindings with `bindings` command

gohandlers can produce a file which contains a series of `Build` and `Parse` methods implemented on available binding types.

Parse methods accept `*http.Request` or `*http.Response` values for request and response binding types, respectively. Those methods use the argument to populate fields of binding type.

```go
func (bq XRequest) Parse(rq *http.Request) error
func (bs XResponse) Parse(rs *http.Response) error
```

`Build` methods returns an instance of `*http.Request` or `*http.Response` reflects the information contained in the binding type. Caller can use `*http.Request` value to call `http.DefaultClient.Do` or dev's other choice.

```go
func (bq XRequest) Build(host ) (*http.Request, error)
func (bs XResponse) Build() (*http.Response, error)
```

Build method on requests needs host information as `http.NewRequest` asks for it.

The produced file named as `bindings.gh.go` by default.

### Implementing a client

gohandlers can produce a file contains the Client struct type and its methods. Each method is named as an handler; accepts a request binding type and returns `*http.Response`, or if it is available the response binding type. Client methods use `http.DefaultClient`s `Do` method to send requests and asks the member `Pool` to provide the address of host per request. This enables load balancing to set up at initialization and simplifies the structure of caller.

```go
type Pool interface {
	Host() (string, error)
}

type Client struct {
	p Pool
}

func NewClient(p Pool) *Client {
	return &Client{p: p}
}

// partial bindings
func (c *Client) CreateUser(bq *CreateUserDetails) (*http.Response, error)

// full bindings
func (c *Client) GetProfile(bq *GetProfileRequest) (*GetProfileResponse, error)
```

## Features

**Handler and binding type detection**

gohandlers marks every function or method that have the `http.HandlerFunc` signature as an handler. If there are struct types in the same file which its name starts with the name of handler; they are detected as binding types for request and response:

```go
type XRequest struct {}
type XResponse struct {}
func X(w http.ResponseWriter, r *http.Request)
```

**Binding types**

gohandlers support binding type declarations with field tags `route`, `query` and `json`, for both the request and response binding types. Binding types are used to:

- access field values inside the `Build` method,
- set field values inside the `Parse` method,
- or implicitly detect the HTTP method of handler.

```go
type XRequest struct {
  Field1 string `route:"field-1"`
  Field2 string `query:"field-2"`
  Field3 string `json:"field-3"`
}
```

In case of the request binding contain a field with a tag for `part` or `file`, gohandlers expects the request content type to be `multipart/form-data`. In this case, all fields need to be tagged with those 2 tags.

**Path assignment**

gohandlers create a path for every handler. Paths start with the handler's name. If there is request binding type for the handler, it gets checked for fields sourced from route (aka. path) variables. For example, the handler below will get `/get-profile/{uid}` as path:

```go
type GetProfileRequest struct {
  UserId models.Uid `route:"uid"`
}

func GetProfile(w http.ResponseWriter, r *http.Request)
```

**Method assignment**

gohandlers checks top-of-the handler comment lines to assign a HTTP method to the handler. The method assigned is used in both `ListHandlers` function and request's `Build` method. Available values are: `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`.

```go
// GET
func X(w http.ResponseWriter, r *http.Request)
```

**Method mismatch warning**

gohandlers prints warnings when an handler is declared a method mismatch with the request binding type. Such as one asks a request body and other doesn't.

**Automatic method detection**

gohandlers implicitly assign most fitting HTTP method to every handler that doesn't specify its prefered HTTP method on the top-of-the handler comment block. The handler is assigned as `GET` if there is no `json` tag specified field in its request binding type. Otherwise it marked as a `POST` request. gohandler prints notice to terminal for implicit method assignments per handler.

**User provided serialization and deserialization**

Values in an incoming HTTP requests are always textual even when they are not meant to be processed as a text. A parsing process should perform the type conversions from text to intented type. Also, the process should take failures in conversion into account. gohandlers implement `Parse` methods with such care.

gohandlers expects fields of binding structs to only be in the types which explicitly defines the serialization and deserialization method from route paramaters and query parameters.

There is 2 interfaces for the types that is intended to be used for fields.

- For fields used for route parameters, the type needs to conform `Routier` interface.
- For fields used for query parameters, the type needs to conform `Querier` interface.

```go
type Routier interface {
  FromRoute(string) error
  ToRoute() (string, error)
}

type Querier interface {
  FromQuery(string) error
  ToQuery() (string, bool, error)
}
```

Such as:

```go
type UserId int

func (uid *UserId) FromRoute(s string) error { /* ... */ }
func (uid *UserId) ToRoute() (string, error) { /* ... */ }

type XRequest struct {
  Uid UserId `route:"uid"`
}

func X(w http.ResponseWriter, r *http.Request) {}
```

Since query parameters are optional, the `Querier` interface method `ToQuery` returns a boolean in the middle of output argument list which represents the availability of a value. Different than returning an error, when this value returned `false`, the `Build` method continues to execution without returning or including a section for that query parameter in the URI.

**Methods required on types used for fields**

| Tag key for field with type | Request.Build        | Request.Parse | Response.Write       | Response.Parse |
| --------------------------- | -------------------- | ------------- | -------------------- | -------------- |
| `route`                     | `ToRoute`            | `FromRoute`   |                      |                |
| `query`                     | `ToQuery`            | `FromQuery`   |                      |                |
| `json`                      |                      |               |                      |                |
| `form`                      | `ToForm`             | `FromForm`    | `ToForm`             | `FromForm`     |
| `part`                      | `ToPart`             | `FromPart`    | `ToPart`             | `FromPart`     |
| `file`                      | `ToFile`, `Filename` | `FromFile`    | `ToFile`, `Filename` | `FromFile`     |

Method signatures are below:

```go
type _ interface {
  ToRoute() (string, error)
  FromRoute(string) error
}

type _ interface {
  ToQuery() (string, error)
  FromQuery(string) error
}

type _ interface {
  ToForm() (string, error)
  FromForm(string) error
}

type _ interface {
  ToPart() (string, error)
  FromPart(string) error
}

type _ interface {
  ToFile(dst io.Writer) error
  Filename()
  FromFile(src io.Reader) error
}
```

Additional methods are required for `multipart/form-data` response/requests on types that are meant to hold data for parts. Dependening on the per-part `Content-Type` value; types need to implement:

| Content-Type                        | Request.Build | Request.Parse  | Response.Write | Response.Parse |
| ----------------------------------- | ------------- | -------------- | -------------- | -------------- |
| `application/json`                  | `ToJsonPart`  | `FromJsonPart` | `ToJsonPart`   | `FromJsonPart` |
| `application/x-www-form-urlencoded` | `ToFromPart`  | `FromFormPart` | `ToFormPart`   | `FromFormPart` |

Content type is detected by gohandlers based on field tags contain `json` or `form` in type declaration of the type used for a request/response part. Method signature are below:

```go
type _ interface {
  ToJsonPart(dst io.Writer) error
  FromJsonPart(src io.Reader) error
}

type _ interface {
  ToFormPart(dst io.Writer) error
  FromFormPart(src io.Reader) error
}
```

**Centralized declaration for `HandlerInfo`**

gohandlers can refer to an existing declaration of `HandlerInfo` instead to define another one per-file. Manually implementing the type and referring to it from every file can reduce the complexity of router registration in architectures such as microservices where the handler declarations of the project splitted to packages across repository.

The type is needed to have 3 fields as such:

```go
type HandlerInfo struct {
	Method string
	Path   string
	Ref    http.HandlerFunc
}
```

Use the related parameters in `list` subcommand to direct gohandlers to use custom declaration in return type of `ListHandlers` function and methods and import the package declares the type.

## Miscellaneous

Outline of the geneated methods' usage:

```
┌─────── Client ───────┐   ┌─────── Server ───────┐
│  Request.Build  ─────┼───┼─► Request.Parse      │
│    Response.Parse ◄──┼───┼──── Response.Write   │
└──────────────────────┘   └──────────────────────┘
```
