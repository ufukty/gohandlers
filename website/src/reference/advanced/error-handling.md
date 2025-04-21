# Error Handling & Customization Patterns in gohandlers

Error handling in HTTP APIs is one of the most nuanced parts of backend development. You want to be descriptive but secure, structured but flexible, and consistent across hundreds of endpoints.

**gohandlers** provides you with just enough structure to handle errors cleanly, while giving you full control over how errors are generated, returned, logged, and encoded.

In this article, we’ll explore the default error handling flow in gohandlers, and how you can customize it to fit your project’s needs—from basic validation to structured error types and application-specific conventions.

---

## 🚧 The Default Error Handling Flow

Out of the box, gohandlers separates your API logic into four clear steps:

1. **Parse** the incoming request:

    ```go
    if err := req.Parse(r); err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
    ```

2. **Validate** the parsed data:

    ```go
    if errs := req.Validate(); len(errs) > 0 {
      w.WriteHeader(http.StatusUnprocessableEntity)
      json.NewEncoder(w).Encode(errs)
      return
    }
    ```

3. **Run your business logic** (possibly returning custom errors)
4. **Write the response or handle application errors**:
    ```go
    if err := resp.Write(w); err != nil {
      http.Error(w, "could not encode response", http.StatusInternalServerError)
    }
    ```

This default structure is clean and minimal. But what if you want more structure? Let’s level up.

---

## 🔄 Centralizing Error Handling

Rather than duplicating error checks in every handler, you can introduce an error-handling layer:

```go
func handleError(w http.ResponseWriter, err error, status int) {
  log.Printf("error: %v", err)

  w.WriteHeader(status)
  json.NewEncoder(w).Encode(map[string]string{
    "error": err.Error(),
  })
}
```

Use it like this:

```go
if err := req.Parse(r); err != nil {
  handleError(w, err, http.StatusBadRequest)
  return
}
```

This keeps your error output consistent and easily stylized (e.g. wrap messages, send error codes, log context).

---

## 🧠 Custom Error Types

You can define structured error types that carry additional metadata:

```go
type APIError struct {
  Code    string `json:"code"`
  Message string `json:"message"`
}

func (e *APIError) Error() string {
  return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
```

In your app logic:

```go
return nil, &APIError{
  Code:    "duplicate_pet",
  Message: "A pet with this name already exists",
}
```

And in your centralized handler:

```go
if apiErr, ok := err.(*APIError); ok {
  w.WriteHeader(http.StatusConflict)
  json.NewEncoder(w).Encode(apiErr)
  return
}
```

Now you have semantic error codes _and_ HTTP status codes.

---

## 🛂 Mapping Errors to Status Codes

Sometimes, it’s helpful to map known errors to appropriate HTTP responses. Here’s a simple mapping layer:

```go
func mapErrorToStatus(err error) int {
  switch {
  case errors.Is(err, sql.ErrNoRows):
    return http.StatusNotFound
  case strings.Contains(err.Error(), "unauthorized"):
    return http.StatusUnauthorized
  default:
    return http.StatusInternalServerError
  }
}
```

Or, with custom error wrapping:

```go
var (
  ErrNotFound = errors.New("not found")
  ErrInvalid  = errors.New("invalid input")
)

func mapAppError(err error) (int, any) {
  switch {
  case errors.Is(err, ErrNotFound):
    return http.StatusNotFound, "resource not found"
  case errors.Is(err, ErrInvalid):
    return http.StatusBadRequest, "invalid data"
  default:
    return http.StatusInternalServerError, "internal server error"
  }
}
```

---

## 🧾 Handling Validation Errors

gohandlers generates a `Validate()` method that returns `map[string]error`. This is a perfect structure for field-level error reporting.

```json
{
    "name": "name is required",
    "email": "invalid email format"
}
```

If you want to customize this further, define your own error wrapper:

```go
type FieldErrors map[string]string

func (f FieldErrors) Error() string {
  return "validation failed"
}
```

In your handler:

```go
if errs := req.Validate(); len(errs) > 0 {
  fieldErrs := FieldErrors{}
  for k, v := range errs {
    fieldErrs[k] = v.Error()
  }
  w.WriteHeader(http.StatusUnprocessableEntity)
  json.NewEncoder(w).Encode(fieldErrs)
  return
}
```

Now you can keep field errors structured and style them however you like.

---

## 🧰 Using Middleware for Global Error Handling

If you want to catch panics or abstract error formatting, you can wrap your handlers with middleware:

```go
func RecoverAndReport(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if v := recover(); v != nil {
        log.Printf("panic: %v", v)
        http.Error(w, "internal server error", http.StatusInternalServerError)
      }
    }()
    next.ServeHTTP(w, r)
  })
}
```

Then apply globally:

```go
mux := http.NewServeMux()
for _, h := range handler.ListHandlers() {
  mux.Handle(h.Path, RecoverAndReport(h.Ref))
}
```

---

## 🧪 Testing Errors in Handlers

Since gohandlers keeps your error flow centralized and clean, you can test each failure path independently.

```go
req := httptest.NewRequest("POST", "/pets", nil)
rec := httptest.NewRecorder()

handler.CreatePet(rec, req)

res := rec.Result()
if res.StatusCode != http.StatusBadRequest {
  t.Fatalf("expected 400, got %d", res.StatusCode)
}
```

You can even test custom error structures if you wrap your error responses consistently.

---

## 🧼 Best Practices

-   Use structured errors (with fields like `code`, `message`, `field`)
-   Map internal errors to HTTP-friendly messages
-   Centralize your error logging and response formatting
-   Always sanitize error messages shown to clients
-   Keep validation errors field-scoped and JSON-encoded
-   Don’t be afraid to panic and recover in dev/test environments

---

## ✅ Summary

Error handling in gohandlers gives you both structure and freedom:

-   Structure through `Parse()` and `Validate()` conventions
-   Freedom through custom types, centralized handlers, and middleware

Whether you're building a prototype or a production-grade API, gohandlers makes it easy to handle errors in a consistent, testable, and user-friendly way—while keeping your core logic clean.

**Say goodbye to duplicated if-else trees and hello to expressive, maintainable error flows.**

Happy handling! 💥🛠️✅
