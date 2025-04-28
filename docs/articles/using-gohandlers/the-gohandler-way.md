# The Gohandler Way

Writing HTTP handlers in Go is straightforward—but repetitive boilerplate often creeps into your codebase. **gohandlers** simplifies this significantly by providing a consistent, maintainable approach to handler creation and management.

This article guides you through the core philosophy and practical approach of structuring HTTP handlers, "the gohandlers way."

## Embrace a Struct-Driven Approach

In gohandlers, your HTTP endpoints are modeled using clearly defined request and response structs. This creates a direct mapping between your Go types and your API’s interface.

**Example Request and Response:**

```go
type GetUserRequest struct {
  UserID string `route:"userId"`
}

type GetUserResponse struct {
  User dto.User `json:"user"`
}
```

-   Struct names clearly communicate their role (`…Request`, `…Response`).
-   Struct tags (`route`, `query`, `json`) directly express how HTTP data maps to Go fields.

## Consistent Handler Signature

With gohandlers, each handler follows the same predictable flow:

1. **Parsing** – Transform the incoming HTTP request into a structured Go request.
2. **Validation** – Ensure input correctness through clearly defined validation methods.
3. **Business Logic** – Handle domain logic or database interactions.
4. **Responding** – Write structured responses back into HTTP responses.

**Example Handler Implementation:**

```go
func (u *Users) GetUser(w http.ResponseWriter, r *http.Request) {
  req := &GetUserRequest{}
  if err := req.Parse(r); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  if errs := req.Validate(); len(errs) > 0 {
    w.WriteHeader(http.StatusUnprocessableEntity)
    json.NewEncoder(w).Encode(errs)
    return
  }

  user, err := u.db.GetUser(req.UserID)
  if err != nil {
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    return
  }

  resp := &GetUserResponse{User: user}
  if err := resp.Write(w); err != nil {
    http.Error(w, "Response Error", http.StatusInternalServerError)
  }
}
```

Each handler is clear, consistent, and readable—helping teams onboard new developers and easily maintain code.

## Leverage Automatic Code Generation

The power of gohandlers comes from its automated code generation. It transforms your simple request and response definitions into fully-fledged parsing, validation, and serialization logic, eliminating boilerplate.

### Automatic Binding Generation

-   Transforms your request structs into methods that parse HTTP requests (`Parse`) and build HTTP responses (`Write`).

### Validation Generation

-   Automatically produces `Validate()` methods that aggregate errors at the field-level.

### Handler Listing

-   Generates a metadata-driven registry (`ListHandlers`) for seamless endpoint registration and documentation.

### Client & Mock Generation

-   Provides strongly typed clients for interacting with your API.
-   Creates mocks for effortless testing.

## Domain Logic Stays in Custom Types

Your business logic stays in custom types, completely separate from generated code. Define your own domain-specific validations and parsing methods:

```go
type Email string

func (e *Email) Validate() error {
  if !strings.Contains(string(*e), "@") {
    return errors.New("invalid email address")
  }
  return nil
}
```

This allows gohandlers to remain focused purely on generating HTTP-related boilerplate, while your custom types encapsulate domain rules clearly.

## Simplified Handler Registration

Gone are the days of manually registering each handler. gohandlers generates a single method to register all handlers automatically:

**Example Generated ListHandlers:**

```go
func (u *Users) ListHandlers() map[string]HandlerInfo {
  return map[string]HandlerInfo{
    "GetUser": {Method: "GET", Path: "/users/{userId}", Ref: u.GetUser},
    "CreateUser": {Method: "POST", Path: "/users", Ref: u.CreateUser},
  }
}
```

This method can then be easily integrated into your router setup:

```go
mux := http.NewServeMux()
for _, h := range NewUsersHandler(db).ListHandlers() {
  mux.HandleFunc(h.Path, h.Ref)
}
http.ListenAndServe(":8080", mux)
```

This approach guarantees synchronization between your route definitions and actual handler implementations.

## Testing Made Easy

With generated mock clients, testing your handlers and business logic becomes straightforward:

```go
mock := &MockClient{}
mock.GetUserFunc = func(ctx context.Context, id string) (*GetUserResponse, error) {
  return &GetUserResponse{User: UserDTO{ID: id, Name: "Test User"}}, nil
}
```

Inject these mocks into your unit tests for predictable, reliable, and isolated testing scenarios.

## Why Adopt the gohandlers Approach?

-   **Consistency:** Every handler shares the same clear structure.
-   **Reduced Boilerplate:** Automated generation removes tedious, repetitive code.
-   **Clear Separation:** Business logic remains separated from HTTP-handling concerns.
-   **Maintainability:** Struct-driven definitions and generated handlers simplify ongoing maintenance and evolution.

By embracing these principles, gohandlers helps your team build maintainable, scalable, and reliable HTTP APIs—letting you focus more on your application’s core logic, and less on boilerplate HTTP code.
