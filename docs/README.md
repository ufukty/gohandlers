# Gohandlers

<img src=".assets/github-social-preview.png" style="width:min(100%, 640px);border-radius:8px">

**Gohandlers** is a CLI tool for Go developers to automate generating 🗿 type safe and 💩 reflectless binding type **helpers**, endpoint **listers** and **clients** for RPC-like use over HTTP without proto.

## Summary

### What you give

A folder with a bunch of Go files each contain binding types and `http.HandlerFunc`s:

```go
type CreateAccountRequest struct {
  Firstname columns.HumanName        `json:"firstname"`
  Lastname  columns.HumanName        `json:"lastname"`
  Birthday  transports.HumanBirthday `json:"birthday"`
  Country   transports.Country       `json:"country"`

  QueryParameter1  `query:"query-param-1"`
  QueryParameter1  `route:"param2"`
}

// POST
func (p *Public) CreateAccount(w http.ResponseWriter, r *http.Request)
```

### What you get

Handler listers:

```go
func ListHandlers() map[string]gohandlers.HandlerInfo
func (pu *Public)  ListHandlers() map[string]gohandlers.HandlerInfo
func (pr *Private) ListHandlers() map[string]gohandlers.HandlerInfo
```

> Listers return the map of handler name and a struct of method, path and pointer to the handler.

Binding helpers:

```go
// For requests:
func (bq CreateAccountRequest) Build(host string) (*http.Request, error)
func (bq *CreateAccountRequest) Parse(rq *http.Request) error
func (bq CreateAccountRequest) Validate() (issues map[string]any)

// For responses:
func (bs CreateEmailGrantResponse) Write(w http.ResponseWriter) error
func (bs *CreateEmailGrantResponse) Parse(rs *http.Response) error
```

Client implementation and a mock with methods like:

```go
type Pool interface {
  Host() (string, error)
}

type Client struct {
  p Pool
}

func (c *Client) CreateAccount(bq *endpoints.CreateAccountRequest) (*http.Response, error)
func (c *Client) CreateEmailGrant(bq *endpoints.CreateEmailGrantRequest) (*endpoints.CreateEmailGrantResponse, error)
func (c *Client) CreatePasswordGrant(bq *endpoints.CreatePasswordGrantRequest) (*endpoints.CreatePasswordGrantResponse, error)
func (c *Client) CreatePhoneGrant(bq *endpoints.CreatePhoneGrantRequest) (*endpoints.CreatePhoneGrantResponse, error)

// for the tests don't need the real deal
type Interface interface {
  CreateAccount(*endpoints.CreateAccountRequest) (*http.Response, error)
  CreateEmailGrant(*endpoints.CreateEmailGrantRequest) (*endpoints.CreateEmailGrantResponse, error)
  CreatePasswordGrant(*endpoints.CreatePasswordGrantRequest) (*endpoints.CreatePasswordGrantResponse, error)
  CreatePhoneGrant(*endpoints.CreatePhoneGrantRequest) (*endpoints.CreatePhoneGrantResponse, error)
}
```

## Why Use Gohandlers?

-   **Single Source of Truth:** The server implementation (your Go handlers and types) becomes the **single source of truth** for your API. No need to maintain separate OpenAPI YAML/JSON files — your Go code is the spec, and Gohandlers generates the rest.
-   **Eliminate Boilerplate:** It **skips repetitive boilerplate** by generating request parsing and response writing code. This means you write your handler logic, define input/output types, and Gohandlers fills in the glue code (like reading query params, writing JSON bodies, etc.).
-   **Type Safety, No Reflection:** Generated code uses static types and standard library calls instead of `reflect`. This ensures **compile-time type checking** and better performance.
-   **Keep Clients Updated:** Gohandlers can generate a Go client (and a mock client for testing) for your API. Whenever your handlers change (e.g., new parameters or endpoints), you can regenerate the client to keep it up-to-date automatically.

In short, Gohandlers aims to **fill the missing gap** in framework-less web development by **connecting your handlers, types, and routing in one automated step**. This improves API **consistency and documentation**, since the code itself describes available endpoints, methods, and expected data.

## Installation

Make sure you have a recent version of Go installed (Go 1.24+). Install Gohandlers by running:

```sh
go install github.com/ufukty/gohandlers/cmd/gohandlers@latest
```

This will put the `gohandlers` binary in your `GOPATH/bin`. To check if your temrinal can find Gohandlers binary:

```sh
which -a gohandlers
```

> **Suggestion**
>
> Use the `make install` command to assign version number correctly.

## Usage

The general syntax of the tool is:

```bash
gohandlers [command] [flags]
```

Run this to see available flags for each subcommand:

```sh
gohandlers [command] -help
```

## Getting started

1. **Define your handlers and types:** Write your handlers following the expected patterns:

    - Signature `func(w http.ResponseWriter, r *http.Request)` (functions or methods).
    - Optional request/response structs with proper naming and tags. Make sure handler body contains request and response body types (even if it's in underscore assignment).
    - (Optionally, add doc comments to specify HTTP method or path if you want to override inference.)

2. **Run Gohandlers:** Use the subcommands as needed:

    - Always start with `bindings` to generate the core Parse/Build/Write code.
    - Use `list` to prepare route registration helpers.
    - Use `client` (and `mock`) if you need client-side code for your API.
    - Use `yaml` if you want an external spec overview.

3. **Integrate into build/test:** Projects are meant to add Gohandlers to a Makefile or a code generation script, so it runs whenever types or handlers change. It’s also common to check in the generated `.gh.go` files so that others can use the client library without generating it themselves. Just remember to regenerate when you make changes.

4. **Use generated code:** In your server, call `bq.Parse(r)` at the top of handlers to get a filled request struct; use `bs.Write(w)` to output responses. In other services (or even the same codebase), use the `Client` to make requests in a type-safe way. In tests, use `Mock` to simulate server behavior.

## How Gohandlers Works?

To better utilize Gohandlers, it helps to know how it identifies handlers and types:

-   **Handler Identification:** Any function with signature `func(http.ResponseWriter, *http.Request)` is considered a handler. If it’s a method (with a receiver), that method is a handler too. You can **optionally add a comment** above the function to explicitly declare the HTTP method (e.g., `// GET` or `// POST`). If no comment is present, Gohandlers will **infer the method** from the name (`GetX` -> GET, `CreateX` -> POST, `UpdateX` -> PATCH, `DeleteX` -> DELETE, etc.) or the presence of a request body. For example, a handler named `GetProfile` with no body will default to GET, whereas `CreateProfile` will default to POST.

-   **Binding Types:** If you define struct types that **share the handler’s name as prefix** and end in `Request` or `Response`, Gohandlers links them as the input/output for that handler. Handlers can have:

    -   Request type only
    -   Response type only
    -   Both request and response types
    -   Neither (in which case parse/build aren’t used)

-   **Tags in Binding Structs:** The fields of your request/response types should use struct tags to indicate where the data comes from or goes to:

    -   `` `route:"var"` `` – corresponds to a path parameter `{var}` in the URL.
    -   `` `query:"var"` `` – corresponds to URL query parameter `?var=`.
    -   `` `json:"var"` `` – part of the JSON body (also implies `Content-Type: application/json`).
    -   `` `form:"var"` `` – part of form data (implies `Content-Type: application/x-www-form-urlencoded`).

    Gohandlers uses these tags to generate correct code. It will skip binding a type to a handler if, for example, the names don’t match or the tags suggest one is for body but the method infers no body, etc., and will warn you of mismatches.

-   **Custom (De)serialization:** Importantly, Gohandlers expects that you provide `FromRoute()`, `FromQuery()`, `FromForm()`, `ToRoute()`, `ToQuery()` and `ToForm()` methods to serialize/deserialize fields. For instance, if you have a type `UserID` used in a `route` tag, your type should implement:

    ```go
    func (id *UserID) FromRoute(value string) error   // parse from path string
    func (id UserID) ToRoute() (string, error)        // convert to string for path
    ```

    Similarly, for `query`, `form` tags (as `FromQuery/ToQuery`, `FromForm/ToForm`). This design **pushes type conversions and validation to your types**, giving you full control (for example, converting a string to an `int64` ID safely, or parsing a date string into a `time.Time`). Basic types like `string`, `int`, etc., are handled by default, but your custom types need these methods so that Gohandlers knows how to handle them in `Parse` and `Build`.

-   **Path Generation:** If you don’t supply a path via comment, Gohandlers will generate one from the handler name and any `route` tags. It strips common verbs (`Get, Create, Update, Delete`) and uses the remainder as the resource. For example, `GetProfile` becomes `/profile`. If `CreateMember` has a route param `supervisor`, the path might become `/member/{supervisor}`. This is an inference to follow REST conventions (but you can always override by providing a full path in a top comment).

-   **Conflict Detection:** Gohandlers will alert you if it detects conflicting paths or methods (like one handler's body/verb inferred method conflicts with the specifed doc method). This helps maintain clarity as your API grows.

## Full example: Petstore

See petstore repository to see a full example which shows how everything works in concert.

Go to [github.com/ufukty/gohandlers-petstore](https://github.com/ufukty/gohandlers-petstore)

## Summary

Gohandlers streamlines Go API development by **eliminating the disconnect between server code and API specifications**. It **solves pain points** like writing request parsing logic, keeping documentation up-to-date, and writing client libraries by hand. With a few commands, you get:

-   Consistent, up-to-date **binding code** that ensures every handler’s inputs and outputs are properly handled.
-   A **client library** that’s always in sync with the server (plus an easy way to test via mocks).
-   An **auto-generated router map** for easy registration, avoiding the error-prone manual setup of routes.
-   A **YAML summary** of your API, should you need to share it outside the Go ecosystem.

The result is a more maintainable codebase: your code is easier to read (handlers show all their inputs/outputs via types), easier to maintain (change a type once and regenerate), and safer for refactoring. It’s ideal for “framework-less” Go development, where you prefer stdlib `net/http` or lightweight routers but still want some of the conveniences that full frameworks or OpenAPI codegen would provide — without their complexity.

Happy coding with **Gohandlers**!
