# 🧶 Core Concepts

Understand the foundational principles behind **gohandlers**, including how it covers request/response bindings, validation, handler listing, client generation, and mocking—so you can design your APIs in alignment with its core abstractions.

---

## Tag‑Driven Binding Types

Your HTTP contract lives in Go structs, annotated with specific tags:

```go
type GetPetRequest struct {
  ID      PetID   `route:"id"`      // path: /pets/{id}
  Verbose bool    `query:"verbose"` // query: ?verbose=true
}

type CreatePetResponse struct {
  Pet  PetDTO   `json:"pet"`        // JSON response body
}
```

-   **`route:"…"`** ➔ binds URL path parameters
-   **`query:"…"`** ➔ binds URL query parameters
-   **`json:"…"`** ➔ serializes/deserializes JSON payloads

These tags directly map HTTP request and response fields, allowing your Go types to clearly define your API's shape.

---

## Automatic Glue Code Generation

gohandlers reads your tagged structs and generates all necessary boilerplate code:

| Stage        | What It Does                                                                           | Generated File(s)      |
| ------------ | -------------------------------------------------------------------------------------- | ---------------------- |
| **bindings** | Creates `Build()`, `Parse()`, and `Write()` methods for request/response serialization | `bindings.gh.go`       |
| **validate** | Adds `Validate() map[string]error` methods for request validation                      | `validate.gh.go`       |
| **list**     | Provides a `ListHandlers()` registry and YAML metadata                                 | `list.gh.go`, `gh.yml` |
| **client**   | Generates strongly typed API clients wrapping HTTP interactions                        | `client.gh.go`         |
| **mock**     | Generates mock implementations of API clients for testing                              | `mock.gh.go`           |

This automated pipeline ensures consistency and removes repetitive coding tasks.

---

## Custom Types & Domain Logic

Domain-specific logic lives in your own custom types, not the generated code:

```go
type PetID string

func (p *PetID) FromRoute(s string) error {
  if s == "" {
    return errors.New("PetID cannot be empty")
  }
  *p = PetID(s)
  return nil
}

func (p PetID) Validate() error {
  if len(p) < 3 {
    return fmt.Errorf("PetID '%s' is too short", p)
  }
  return nil
}
```

-   Implement `FromRoute(string)` and `FromQuery(string)` to define parsing logic.
-   Use `Validate()` methods to enforce data constraints and business rules.

Generated code simply invokes these methods, maintaining separation between your domain logic and HTTP handling code.

---

## Consistent Handler Flow

Handlers built with gohandlers typically follow a consistent four-step pattern:

1. **Parse**: Convert incoming HTTP requests into structured request objects.
2. **Validate**: Run field-level validation methods to ensure data correctness.
3. **Execute**: Handle business logic (database calls, computations, etc.).
4. **Respond**: Serialize structured response objects back into HTTP responses.

Here's an example handler following this structure:

```go
func (p *Pets) CreatePet(w http.ResponseWriter, r *http.Request) {
  req := &CreatePetRequest{}
  if err := req.Parse(r); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  if errs := req.Validate(); len(errs) > 0 {
    w.WriteHeader(http.StatusUnprocessableEntity)
    json.NewEncoder(w).Encode(errs)
    return
  }

  pet, err := p.db.Insert(req.Body)
  if err != nil {
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }

  resp := &CreatePetResponse{Pet: pet}
  if err := resp.Write(w); err != nil {
    http.Error(w, "error writing response", http.StatusInternalServerError)
  }
}
```

This uniform approach simplifies readability, debugging, and testing.

---

## Metadata‑Driven Handler Registration

Rather than manually registering each handler, gohandlers generates a registry:

```go
map[string]HandlerInfo{
  "CreatePet": {Method:"POST", Path:"/pets", Ref: p.CreatePet},
  "GetPet":    {Method:"GET",  Path:"/pets/{id}", Ref: p.GetPet},
}
```

This registry enables automatic routing setup and keeps documentation and server code synchronized effortlessly.

---

## Typed Client Generation & Mocking

---

### Typed API Clients

gohandlers generates typed Go client implementations that mirror your server's API endpoints:

```go
client := NewClient(pool)
resp, err := client.CreatePet(ctx, petDTO)
```

These clients abstract away repetitive HTTP request logic, giving you strongly typed interactions with your APIs.

---

### Mock Clients for Testing

Generated mock implementations help unit-test your business logic without real HTTP calls:

```go
mock := &MockClient{}
mock.CreatePetFunc = func(_ context.Context, dto PetDTO) (*CreatePetResponse, error) {
  return &CreatePetResponse{Pet: Pet{ID: "test123"}}, nil
}
```

These mocks streamline testing, enabling quick and reliable test setups.

---

## Benefits and Philosophy

-   **Declarative API**: Define API endpoints using struct tags, keeping definitions clear and intuitive.
-   **Strong Typing**: Generated code respects your Go types, minimizing runtime errors.
-   **Reduced Boilerplate**: Automate tedious code, maintaining focus on business logic.
-   **Easy Testing**: Consistent handlers and built-in mocking simplify testing practices.
-   **Maintainable Design**: Centralized definitions and generated code keep your project easy to scale and maintain.

By embracing these core concepts, gohandlers helps you build maintainable, robust, and efficient HTTP APIs in Go.
