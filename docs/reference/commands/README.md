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
