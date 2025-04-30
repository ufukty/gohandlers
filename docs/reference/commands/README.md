# 🛠️ Commands

Gohandlers provide multiple commands to support stepped adoption. Here is a quick overview:

| Command      | Purpose                                                                                 | Generates        |
| ------------ | --------------------------------------------------------------------------------------- | ---------------- |
| **bindings** | Creates `Build()`, `Parse()`, and `Write()` methods for request/response serialization. | `bindings.gh.go` |
| **client**   | Generates strongly typed API clients wrapping HTTP interactions.                        | `client.gh.go`   |
| **list**     | Provides a `ListHandlers()` registry.                                                   | `list.gh.go`     |
| **mock**     | Generates mock implementations of API clients for testing.                              | `mock.gh.go`     |
| **validate** | Adds `Validate() map[string]error` methods for request validation.                      | `validate.gh.go` |
| **yaml**     | Writes the handler metadata to a YAML file.                                             | `gh.yml`         |

## Commands, briefly

### ✱ `bindings`

Generates a Go file (default `bindings.gh.go`) containing `Build`, `Parse` and `Write` methods for each “binding” struct in your code. **Binding structs** are simply your request and response types for handlers that contains route, query, form or json values as fields.

---

### ✱ `client`

Generates a Go file containing a **Client struct** and one method per handler function. These methods construct HTTP requests using your binding types and send them, returning the response (either raw or parsed into a response binding).

Writing and maintaining custom client code for your API (or using generic tools) can lead to mismatches as your API evolves. The Gohandlers client ensures **the client library always matches the server**. If you add a new query parameter or change an endpoint, regenerate the client — it will have the updated method signature and logic.

---

### ✱ `list`

Generates a Go file (default `list.gh.go`) with a function or methods to **list all handlers and their metadata** (like method and path). This is extremely useful to automatically register your handlers with a router (e.g., `http.ServeMux` or any router library) without manually writing each route.

In a growing API, it’s easy to forget to register a handler or to mismatch the route path. With `ListHandlers`, you ensure every handler is accounted for.

---

### ✱ `mock`

Generates a Go file (default `mock.gh.go`) with mock implementation of `Client` and an interface to use in declaring fields/variables to hold client in consumer/caller side.

---

### ✱ `validate`

Generates a Go file (default `validate.gh.go`) contains validation helpers for request binding types. Those helpers are meant to be called from handlers, after parsing is done. Validation methods call each field's validation method one by one, and collects all errors in a map. Then the handler can use its custom method of serialization based on its return type being JSON or HTML.

---

### ✱ `yaml`

Generates a YAML file (default `gh.yml`) that lists handlers, their HTTP methods, and paths. This is like a lightweight documentation or can be seen as a mini OpenAPI output. It’s useful if other services or tools (not written in Go) need to know about your API.
