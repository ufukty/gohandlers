# gohandlers

gohandlers is a CLI tool that creates Go files contain information about handlers and their binding types that are defined in a directory. The produced files make the information previously embedded in the type names and discarded by compiler available to developer's use.

gohandlers allows defining the path and method information in one place and to get type-checked.

## Usage

```
gohandlers [command] <command args>

commands:
    bindings : Produces the file contains Build and Parse methods for binding types
    list     : Produces the file contains ListHandlers function/methods
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

## Features

**Handler and binding type detection**

gohandlers marks every function or method that have the `http.HandlerFunc` signature as an handler. If there are struct types in the same file which its name starts with the name of handler; they are detected as binding types for request and response:

```go
type XRequest struct {}
type XResponse struct {}
func X(w http.ResponseWriter, r *http.Request)
```

**Binding types**

gohandlers support binding type declarations with field tags `route`, `query`, `json` and `cookie`, for both the request and response binding types. Binding types are used to:

- access field values inside the `Build` method,
- set field values inside the `Parse` method,
- or implicitly detect the HTTP method of handler.

```go
type XRequest struct {
  Field1 string `route:"field-1"`
  Field2 string `query:"field-2"`
  Field3 string `json:"field-3"`
  Field4 string `cookie:"field-4"`
}
```

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

**Automatic method detection**

gohandlers implicitly assign most fitting HTTP method to every handler that doesn't specify its prefered HTTP method on the top-of-the handler comment block. The handler is assigned as `GET` if there is no `json` tag specified field in its request binding type. Otherwise it marked as a `POST` request. gohandler prints notice to terminal for implicit method assignments per handler.

**Type aware assignments**

Assigning values from text based http request to type aware binding type fields should take the failures into account. gohandlers implement `Parse` methods with care.

Code generation involves checking the field types to see if they are also text based or not. Text based types such as `string` and `[]byte` results with explicit `=` assignment statements appear in the body of `Parse` methods. The custom types which embeds those types will result the right hand side of the assignment get wrapped with explicit conversion.

For non-textual types; gohandlers checks if they implement a `Set` method which accepts a string value and returns an error if the assignment fails. The method should set the variable's value based on the argument.

```go
func (x *X) Set(v string) error
```

Such as:

```go
type UserId int

func (uid *UserId) Set(s string) error {
  i, err := strconv.Atoi(s)
  if err != nil {
    return fmt.Errorf("converting text to number: %w", err)
  }
  uid = UserId(i)
  return nil
}
```

Now the `UserId` can be used as a field's type in a binding type. Value assignment in the `Parse` method for related field of this type will involve the call of `Set` method instead of using `=` operator. Following line in the `Parse` method will check if `Set` returned an error. If so, the `Parse` method will also return early after wrap the error. gohandlers will follow the import path for examining the type's declaration and methods.

Any request field's type that is neither textual nor implements `Set` method will lead the generation to exit with failure. The types will be listed on terminal to allow dev to fix the issue.
