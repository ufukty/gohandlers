# gohandlers

gohandlers is a CLI tool that creates Go files contain information about handlers and their binding types that are defined in a directory. The produced files make the information previously embedded in the type names and discarded by compiler available to developer's use.

gohandlers allows defining the path and method information in one place and to get type-checked. gohandlers is designed to let the server implementation of an API to be the single source of truth to let it contain all the information to generate client code and rest. So, gohandlers makes writing API spec definitions redundant.

Contrary to frameworks; using binding types with gohandlers doesn't require reflection (`reflect` package) as it generates the code to map request/response from/to binding type.

With gohandlers, developers can declare a type to list all the parameters for input and output for an endpoint. Writing endpoints and their bindings in the way gohandlers expects is great to make the implementation serve as documentation without maintaining separate files since separate files can get out of date very quickly. As the type definition provides the list of all parameters at a glance using gohandlers contributes to the overall readability of project.

## Index

- [Index](#index)
- [Usage](#usage)
  - [bindings](#bindings)
  - [client](#client)
  - [list](#list)
- [Documentation](#documentation)
  - [Handlers](#handlers)
  - [Bindings](#bindings-1)
    - [Declaring bindings](#declaring-bindings)
    - [User provided serialization and deserialization](#user-provided-serialization-and-deserialization)
      - [Methods expected by developer on field types](#methods-expected-by-developer-on-field-types)
  - [Handler meta data](#handler-meta-data)
    - [Path](#path)
      - [Path generation](#path-generation)
    - [Method](#method)
      - [Assignment](#assignment)
      - [Inference](#inference)
    - [Content type](#content-type)
- [Miscellaneous](#miscellaneous)
- [Internals](#internals)
  - [Additional methods on multipart form data](#additional-methods-on-multipart-form-data)
  - [Centralized declaration for `HandlerInfo`](#centralized-declaration-for-handlerinfo)
- [Considerations](#considerations)

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

### bindings

gohandlers can produce a file which contains a series of `Build` and `Parse` methods implemented on available binding types.

Parse methods accept `*http.Request` or `*http.Response` values for request and response binding types, respectively. Those methods use the argument to populate fields of binding type.

```go
func (bq XRequest) Parse(rq *http.Request) error
func (bs XResponse) Parse(rs *http.Response) error
```

`Build` methods returns an instance of `*http.Request` or `*http.Response` reflects the information contained in the binding type. Caller can use `*http.Request` value to call `http.DefaultClient.Do` or dev's other choice.

```go
func (bq XRequest) Build(host string) (*http.Request, error)
func (bs XResponse) Write(w http.ResponseWriter) error
```

Build method on requests needs host information as `http.NewRequest` asks for it.

The produced file named as `bindings.gh.go` by default.

### client

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

### list

gohandlers can generate a function which returns a `map` of handlers. This allows developer to register handlers to `http.ServeMux` (or a router of dev's choice) by simply iterating over it and call the router's method.

Collecting method and path information of all handlers from the handlers improves consistency over development time, enforcing the single source of truth principle.

The produced file gets named as `list.gh.go` by default. The file will contain `ListHandlers` function for handlers defined on the global scope. It will also contain as many `ListHandler` methods as the types which contains at least one method in the `http.HandlerFunc` signature.

```go
func ListHandlers() map[string]reception.HandlerInfo
func (p *Public) ListHandlers() map[string]reception.HandlerInfo
func (p *Private) ListHandlers() map[string]reception.HandlerInfo
```

## Documentation

### Handlers

gohandlers marks every function or method that have the `http.HandlerFunc` signature as an handler. Such as:

```go
func X(w http.ResponseWriter, r *http.Request)
func (a *A) X(w http.ResponseWriter, r *http.Request)
```

Handlers can have a top-comment to manually assign the HTTP method.

```go
// GET
func X(w http.ResponseWriter, r *http.Request)

// POST
func X(w http.ResponseWriter, r *http.Request)
```

Handler detection is the heart of gohandlers; as the information related to them is the base of each file generated by program.

### Bindings

When there are struct types that are declared in the same file and their name starts with the name of an handler; they are marked as request/response binding types for the handler:

```go
type XRequest struct {}
type XResponse struct {}
func X(w http.ResponseWriter, r *http.Request)
```

Binding types are optional for handlers. Developer can implement one, both or none.

#### Declaring bindings

To implement a request or response binding type, create a struct type declaration. The typename should start with the handler name and continue with `Request` or `Response`. All fields need to be tagged properly. Field tags are used to guide gohandlers to generate `Build`, `Write` and `Parse` methods correctly. The field tags contain one key and its value that represents the point of the request/response. Gohandlers supports variety of source for parametes:

| Field tag       | Targets |
| --------------- | ------- |
| `route`         | Header  |
| `query`         | Header  |
| `json`          | Body    |
| `form`          | Body    |
| `part` / `file` | Body    |

This example shows how the bindings for multipart requests are defined. The body content type will be decided as `multipart/form-data` since the request binding type only contain field tags: `part` and `file`. And the content type for `profile` field will be decided as `json` as the type definition for field only contains field tags with it.

```go
type CreateMemberRequest struct {
  Supervisor     columns.UserId                    `route:"supervisor"`
  Role           columns.Role                      `query:"role"`
  Email          columns.Email                     `part:"email"`
  ProfileDetails CreateMemberRequestProfileDetails `part:"profile"`
  ProfilePhoto   *multipart.FileHeader             `file:"photo"`
}

type CreateMemberRequestProfileDetails struct {
  Name     columns.HumanName `json:"name"`
  Lastname columns.HumanName `json:"lastname"`
  Birthday columns.Date      `json:"birthday"`
}

func CreateMember(w http.ResponseWriter, r *http.Request)
```

Notice how the binding type contains only non-basic types for fields. This is because the serialization to and deserialization from request/response content is performed explicitly in gohandlers; gohandlers expects developer to implement methods on field types to perform those from/to text conversions securely.

#### User provided serialization and deserialization

Values in an incoming HTTP requests are always textual even when they are not meant to be processed as a text. A parsing process should perform the type conversions from text to intented type. Also, the process should take failures in conversion into account. gohandlers implement `Parse` methods with such care.

gohandlers expects fields of binding structs to only be in the types which explicitly defines the serialization and deserialization method from route paramaters and query parameters.

##### Methods expected by developer on field types

| Tag for field type | Request.Build | Request.Parse     | Response.Write | Response.Parse    |
| ------------------ | ------------- | ----------------- | -------------- | ----------------- |
| `route`            | `ToRoute`     | `FromRoute`       |                |                   |
| `query`            | `ToQuery`     | `FromQuery`       |                |                   |
| `json`             |               |                   |                |                   |
| `form`             | `ToForm`      | `FromForm`        | `ToForm`       | `FromForm`        |
| `part`             | `ToPart`      | `FromPart`        | `ToPart`       | `FromPart`        |
| `file`             | `ToFile`      | `FromFileHandler` | `ToFile`       | `FromFileHandler` |

Method signatures are below:

```go
func (t Type) ToRoute() (string, error)
func (t *Type) FromRoute(string) error
```

```go
func (t Type) ToQuery() (string, error)
func (t *Type) FromQuery(string) error
```

```go
func (t Type) ToForm() (string, error)
func (t *Type) FromForm(string) error
```

```go
func (t Type) ToPart() (string, error)
func (t *Type) FromPart(string) error
```

```go
func (t Type) ToFile() (src io.Reader, filename string, contentType string, err error)
func (t *Type) FromFile(src multipart.FileHandler) error
```

### Handler meta data

gohandlers assigns a path, a method and a content type to every handler. It supports explicit assignment through a doc comment, and inference that is based on handler name and binding type content. For handlers that the developer wants to set a specific path and method, assignment through doc comment must be preferred. For rest of the handlers inference will work okay as long as the client code also gets built on the output of gohandlers.

The meta data finds its use in various places:

| Metadata        | Place                                                                                    |
| --------------- | ---------------------------------------------------------------------------------------- |
| path and method | `ListHandlers` function and methods,<br>`Build` methods declared on request binding type |
| content-type    | `Build`, `Write`, `Parse` method on binding types                                        |

#### Path

gohandlers checkes the doc comment of the handler to check if there is a path provided for the handler. If so, the path is assigned to handler. If not, gohandlers generates a path based on the handler name and route parameters in the request binding type.

Provided paths are still checked for conflicts with the

##### Path generation

The path generation creates a path name starts with the handler's name and continues with the route parameters, separated with slashes. The action prefix in the handler name will be trimmed.

See the example:

```go
type GetProfileRequest struct {
  UserId models.Uid `route:"uid"`
}

func GetProfile(w http.ResponseWriter, r *http.Request)
```

the handler will get the path `/profile/{uid}`. Notice the path doesn't contain the action prefix that exist in the handler name.

#### Method

gohandlers can decide on the method for handler either by assignment or inference. In case of there is developer provided method in the doc-comment of handler, gohandlers assigns the method to handler. Otherwise gohandlers run inference to decide on the proper method. Inference leverages the action prefix in handler name and binding type for the handler.

gohandlers prints warnings when an handler is declared a method mismatch with the request binding type. Such as one asks a request body and other doesn't.

gohandler prints notice to terminal for method inference per handler.

##### Assignment

Developers can assign desired method to a handler by typing it to the doc-comment as such:

```go
// GET
func X(w http.ResponseWriter, r *http.Request)
```

Available values are: `GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`.

##### Inference

Developers either assign a method to handler, or let gohandlers to decide. In case of latter, gohandlers levarages handler name and request binding type to infer the method. If the handler name starts with either of the keywords listed in table,

| Action prefix     | Body     | Inferred method |
| ----------------- | -------- | --------------- |
| `Get`, `Visit`    | Absent   | `GET`           |
| `Post`, `Create`  | Expected | `POST`          |
| `Patch`, `Update` | Expected | `PATCH`         |
| `Put`, `Replace`  | Expected | `PUT`           |
| `Delete`          |          | `DELETE`        |

If there is none of the action prefixes present in the handler name; then gohandlers will look to the request binding type to see if there is a body in it. In case of a field with body targeting field tags exist, then the method is inferred as `POST`.

#### Content type

| Field tag(s)   | Inferred content type               |
| -------------- | ----------------------------------- |
| `json`         | `application/json`                  |
| `form`         | `application/x-www-form-urlencoded` |
| `part`, `file` | `multipart/form-data`               |

## Miscellaneous

Outline of the geneated methods' usage:

```
┌─────── Client ───────┐   ┌─────── Server ───────┐
│  Request.Build  ─────┼───┼─► Request.Parse      │
│    Response.Parse ◄──┼───┼──── Response.Write   │
└──────────────────────┘   └──────────────────────┘
```

## Internals

### Additional methods on multipart form data

Additional methods are declared for `multipart/form-data` response/requests on types that are meant to hold data for parts. Dependening on the per-part `Content-Type` value; types implement:

| Content-Type                        | Request.Build | Request.Parse  | Response.Write | Response.Parse |
| ----------------------------------- | ------------- | -------------- | -------------- | -------------- |
| `application/json`                  | `ToJsonPart`  | `FromJsonPart` | `ToJsonPart`   | `FromJsonPart` |
| `application/x-www-form-urlencoded` | `ToFromPart`  | `FromFormPart` | `ToFormPart`   | `FromFormPart` |

Content type is detected by gohandlers based on field tags contain `json` or `form` in type declaration of the type used for a request/response part. Method signature are below:

```go
func (t Type) ToJsonPart(dst io.Writer) error
func (t *Type) FromJsonPart(src io.Reader) error
```

```go
func (t Type) ToFormPart(dst io.Writer) error
func (t *Type) FromFormPart(src io.Reader) error
```

### Centralized declaration for `HandlerInfo`

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

## Considerations

- Streaming is not supported. Adding support is not planned.
