# Gohandlers

<img src=".assets/github-social-preview.png" style="width:min(100%, 640px);border-radius:8px">

**Gohandlers** is a CLI tool for Go developers to automate generating type safe and reflectless binding type **helpers**, endpoint **listers** and **clients** for RPC-like use over HTTP without proto.

In short, Gohandlers aims to **fill the missing gap** in framework-less web development by **connecting your handlers, types, and routing in one automated step**. This improves API **consistency and documentation**, since the code itself describes available endpoints, methods, and expected data.

## Features

-   On binding types:
    -   Build, parse and validate requests
    -   Write and parse responses
-   Get listers:
    -   Easy to register all endpoints
    -   No handlers left unregistered
-   Client code
    -   Real and mock implementations
    -   Interface for testing
-   Route method & paths:
    -   Inference via handler prefix and binding body
    -   Override via doc-comments
    -   Conflict checks
-   Enforces on field types:
    -   Validation methods
    -   De/serialization methods for query, route and form contexts
-   Safety:
    -   No fields are skipped
    -   Compile time checks on missing field validators
    -   Single source of truth for API definition
-   Convenience
    -   Integrate into build system
    -   Always up-to-date
    -   Leverage code completion
    -   No boilerplate
-   Keep using the good old:
    -   net/http for requests/responses
    -   json encoder for bodies

## What you give and get

### What you give

A folder with a bunch of Go files each contain binding types and `http.HandlerFunc`s:

```go
type CreateAccountRequest struct {
  Firstname columns.HumanName        `json:"firstname"`
  Lastname  columns.HumanName        `json:"lastname"`
  Birthday  transports.HumanBirthday `json:"birthday"`
  Country   transports.Country       `json:"country"`

  RouteParameter1 basics.String `route:"param2"`
  QueryParameter1 basics.Int    `query:"query-param-1"`
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

## Go to Petstore

See [Petstore](https://github.com/ufukty/gohandlers-petstore) example in Github to understand how backend development works with Gohandlers:
