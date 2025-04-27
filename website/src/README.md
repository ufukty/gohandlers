> Work in progress. The direction on implementing features and solving bugs is depending on the reach. Leave a mark in issues, PRs or discussions mentioning your use case for gohandlers.

# gohandlers

<img src=".assets/logo@3x.png" width="256px">

**gohandlers** is a command-line tool for Go that automates the generation of boilerplate code for HTTP handlers and their associated types. It’s designed for developers building REST APIs in Go, who want to keep server and client code in sync without manually writing OpenAPI/Swagger specs. By analyzing your Go code (handlers and types), gohandlers generates type-safe code for parsing HTTP requests, building responses, registering routes, and even creating client libraries — all without reflection and with compile-time safety.

## Why Use gohandlers?

- **Single Source of Truth:** The server implementation (your Go handlers and types) becomes the **single source of truth** for your API. No need to maintain separate OpenAPI YAML/JSON files — your Go code is the spec, and gohandlers generates the rest.
- **Eliminate Boilerplate:** It **skips repetitive boilerplate** by generating request parsing and response writing code. This means you write your handler logic, define input/output types, and gohandlers fills in the glue code (like reading query params, writing JSON bodies, etc.).
- **Type Safety, No Reflection:** Generated code uses static types and standard library calls instead of `reflect`. This ensures **compile-time type checking** and better performance.
- **Keep Clients Updated:** gohandlers can generate a Go client (and a mock client for testing) for your API. Whenever your handlers change (e.g., new parameters or endpoints), you can regenerate the client to keep it up-to-date automatically.

In short, gohandlers aims to **fill the missing gap** in framework-less web development by **connecting your handlers, types, and routing in one automated step**. This improves API **consistency and documentation**, since the code itself describes available endpoints, methods, and expected data.

## Installation

Make sure you have a recent version of Go installed (Go 1.18+). Install gohandlers by running:

```bash
go install github.com/ufukty/gohandlers@latest
```

This will put the `gohandlers` binary in your `GOPATH/bin` (assuming Go modules are enabled).

Note that you need to use custom flags as in [`Makefile`](./Makefile) to embed the version number into binary.

## Usage

The general syntax of the tool is:

```bash
gohandlers [command] [flags]
```

**Subcommands:**

- **`bindings`**
- **`client`**
- **`list`**
- **`mock`**
- **`validate`**
- **`yaml`**

Run `gohandlers [command] -help` to see available flags for each subcommand.

## Subcommands, briefly

### `bindings`

Generates a Go file (default `bindings.gh.go`) containing `Build`, `Parse` and `Write` methods for each “binding” struct in your code. **Binding structs** are simply your request and response types for handlers that contains route, query, form or json values as fields.

Continue in [Docs > Commands > `bindings`](docs/commands/bindings.md)

### `client`

Generates a Go file containing a **Client struct** and one method per handler function. These methods construct HTTP requests using your binding types and send them, returning the response (either raw or parsed into a response binding).

Writing and maintaining custom client code for your API (or using generic tools) can lead to mismatches as your API evolves. The gohandlers client ensures **the client library always matches the server**. If you add a new query parameter or change an endpoint, regenerate the client — it will have the updated method signature and logic.

Continue in [Docs > Commands > `client`](docs/commands/client.md)

### `list`

Generates a Go file (default `list.gh.go`) with a function or methods to **list all handlers and their metadata** (like method and path). This is extremely useful to automatically register your handlers with a router (e.g., `http.ServeMux` or any router library) without manually writing each route.

In a growing API, it’s easy to forget to register a handler or to mismatch the route path. With `ListHandlers`, you ensure every handler is accounted for.

Continue in [Docs > Commands > `list`](docs/commands/list.md)

### `mock`

Generates a Go file (default `mock.gh.go`) with mock implementation of `Client` and an interface to use in declaring fields/variables to hold client in consumer/caller side.

Continue in [Docs > Commands > `mock`](docs/commands/mock.md)

### `validate`

Generates a Go file (default `validate.gh.go`) contains validation helpers for request binding types. Those helpers are meant to be called from handlers, after parsing is done. Validation methods call each field's validation method one by one, and collects all errors in a map. Then the handler can use its custom method of serialization based on its return type being JSON or HTML.

Continue in [Docs > Commands > `validate`](docs/commands/validate.md)

### `yaml`

Generates a YAML file (default `gh.yml`) that lists handlers, their HTTP methods, and paths. This is like a lightweight documentation or can be seen as a mini OpenAPI output. It’s useful if other services or tools (not written in Go) need to know about your API.

Continue in [Docs > Commands > `yaml`](docs/commands/yaml.md)

## Getting started

1. **Define your handlers and types:** Write your handlers following the expected patterns:

   - Signature `func(w http.ResponseWriter, r *http.Request)` (functions or methods).
   - Optional request/response structs with proper naming and tags. Make sure handler body contains request and response body types (even if it's in underscore assignment).
   - (Optionally, add doc comments to specify HTTP method or path if you want to override inference.)

2. **Run gohandlers:** Use the subcommands as needed:

   - Always start with `bindings` to generate the core Parse/Build/Write code.
   - Use `list` to prepare route registration helpers.
   - Use `client` (and `mock`) if you need client-side code for your API.
   - Use `yaml` if you want an external spec overview.

3. **Integrate into build/test:** Projects are meant to add gohandlers to a Makefile or a code generation script, so it runs whenever types or handlers change. It’s also common to check in the generated `.gh.go` files so that others can use the client library without generating it themselves. Just remember to regenerate when you make changes.

4. **Use generated code:** In your server, call `bq.Parse(r)` at the top of handlers to get a filled request struct; use `bs.Write(w)` to output responses. In other services (or even the same codebase), use the `Client` to make requests in a type-safe way. In tests, use `Mock` to simulate server behavior.

## How gohandlers Works (Briefly)

To better utilize gohandlers, it helps to know how it identifies handlers and types:

- **Handler Identification:** Any function with signature `func(http.ResponseWriter, *http.Request)` is considered a handler. If it’s a method (with a receiver), that method is a handler too. You can **optionally add a comment** above the function to explicitly declare the HTTP method (e.g., `// GET` or `// POST`). If no comment is present, gohandlers will **infer the method** from the name (`GetX` -> GET, `CreateX` -> POST, `UpdateX` -> PATCH, `DeleteX` -> DELETE, etc.) or the presence of a request body. For example, a handler named `GetProfile` with no body will default to GET, whereas `CreateProfile` will default to POST.

- **Binding Types:** If you define struct types that **share the handler’s name as prefix** and end in `Request` or `Response`, gohandlers links them as the input/output for that handler. In [this example](docs/commands/bindings.md#example), `CreatePetRequest` and `CreatePetResponse` are automatically tied to the `CreatePet` handler. Handlers can have:

  - Request type only
  - Response type only
  - Both request and response types
  - Neither (in which case parse/build aren’t used)

- **Tags in Binding Structs:** The fields of your request/response types should use struct tags to indicate where the data comes from or goes to:

  - `` `route:"var"` `` – corresponds to a path parameter `{var}` in the URL.
  - `` `query:"var"` `` – corresponds to URL query parameter `?var=`.
  - `` `json:"var"` `` – part of the JSON body (also implies `Content-Type: application/json`).
  - `` `form:"var"` `` – part of form data (implies `Content-Type: application/x-www-form-urlencoded`).

  gohandlers uses these tags to generate correct code. It will skip binding a type to a handler if, for example, the names don’t match or the tags suggest one is for body but the method infers no body, etc., and will warn you of mismatches.

- **Custom (De)serialization:** Importantly, gohandlers expects that for non-basic types in your bindings, you provide methods to serialize/deserialize them. For instance, if you have a type `UserID` used in a `route` tag, your type should implement:

  ```go
  func (id *UserID) FromRoute(value string) error   // parse from path string
  func (id UserID) ToRoute() (string, error)        // convert to string for path
  ```

  Similarly, for `query`, `form` tags (as `FromQuery/ToQuery`, `FromForm/ToForm`). This design **pushes type conversions and validation to your types**, giving you full control (for example, converting a string to an `int64` ID safely, or parsing a date string into a `time.Time`). Basic types like `string`, `int`, etc., are handled by default, but your custom types need these methods so that gohandlers knows how to handle them in `Parse` and `Build`.

- **Path Generation:** If you don’t supply a path via comment, gohandlers will generate one from the handler name and any `route` tags. It strips common verbs (`Get, Create, Update, Delete`) and uses the remainder as the resource. For example, `GetProfile` becomes `/profile`. If `CreateMember` has a route param `supervisor`, the path might become `/member/{supervisor}`. This is an inference to follow REST conventions (but you can always override by providing a full path in a top comment).

- **Conflict Detection:** gohandlers will alert you if it detects conflicting paths or methods (like one handler's body/verb inferred method conflicts with the specifed doc method). This helps maintain clarity as your API grows.

## Full example: Petstore

See petstore repository to see a full example which shows how everything works in concert.

Go to [github.com/ufukty/gohandlers-petstore](https://github.com/ufukty/gohandlers-petstore)

## Summary

gohandlers streamlines Go API development by **eliminating the disconnect between server code and API specifications**. It **solves pain points** like writing request parsing logic, keeping documentation up-to-date, and writing client libraries by hand. With a few commands, you get:

- Consistent, up-to-date **binding code** that ensures every handler’s inputs and outputs are properly handled.
- A **client library** that’s always in sync with the server (plus an easy way to test via mocks).
- An **auto-generated router map** for easy registration, avoiding the error-prone manual setup of routes.
- A **YAML summary** of your API, should you need to share it outside the Go ecosystem.

The result is a more maintainable codebase: your code is easier to read (handlers show all their inputs/outputs via types), easier to maintain (change a type once and regenerate), and safer for refactoring. It’s ideal for “framework-less” Go development, where you prefer stdlib `net/http` or lightweight routers but still want some of the conveniences that full frameworks or OpenAPI codegen would provide — without their complexity.

Happy coding with **gohandlers**!

## License

MIT
