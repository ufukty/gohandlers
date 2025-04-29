# Validation & Custom Types

One of the most powerful features of **gohandlers** is its ability to generate consistent and thorough validation logic—without adding extra boilerplate to your handlers. This is made possible by a combination of **validation-aware request structs** and support for **custom types** that encapsulate their own parsing and checking rules.

In this article, we’ll explore how gohandlers manages validation, how to design custom types, and how to apply both in real-world scenarios.

---

## ✅ Why Validation Matters

Validation ensures your application doesn't receive malformed, missing, or inconsistent data. In traditional Go APIs, you often write repetitive if-else blocks like:

```go
if req.Email == "" {
  http.Error(w, "email is required", http.StatusBadRequest)
  return
}
```

These checks become cumbersome to maintain, especially across many endpoints.

With **gohandlers**, validation is cleanly defined within the struct and handled via generated `Validate()` methods—without cluttering your handler logic.

---

## 🧩 Auto-Generated Field-Level Validation

gohandlers analyzes your `...Request` structs and produces a `Validate()` method for each, which returns a `map[string]error`.

Example:

```go
type CreatePetRequest struct {
  Name string `json:"name"`
  Tag  string `json:"tag"`
}
```

Generated:

```go
func (req CreatePetRequest) Validate() map[string]error {
  errs := map[string]error{}
  if req.Name == "" {
    errs["name"] = errors.New("name is required")
  }
  if req.Tag == "" {
    errs["tag"] = errors.New("tag is required")
  }
  return errs
}
```

Inside your handler, you simply call:

```go
if errs := req.Validate(); len(errs) > 0 {
  w.WriteHeader(http.StatusUnprocessableEntity)
  json.NewEncoder(w).Encode(errs)
  return
}
```

No need to repeat yourself.

---

## 🧠 Custom Types for Reusable Validation

Validation gets even more powerful when you move logic into **custom types**. You can define domain-specific types that:

-   Know how to parse themselves from strings (e.g., from URL/query params)
-   Know how to validate themselves

gohandlers detects and uses these interfaces automatically:

-   `FromRoute(string) error`
-   `FromQuery(string) error`
-   `Validate() error`

---

### Example: Email Type

```go
type Email string

func (e *Email) FromQuery(s string) error {
  *e = Email(s)
  return nil
}

func (e Email) Validate() error {
  if !strings.Contains(string(e), "@") {
    return errors.New("invalid email format")
  }
  return nil
}
```

Used in a request struct:

```go
type InviteUserRequest struct {
  Email Email `query:"email"`
}
```

gohandlers will automatically:

-   Call `FromQuery()` during parsing.
-   Call `Validate()` during validation.
-   Include `"email": "invalid email format"` in the error map if needed.

---

## 🔄 Reusable Validators with `pkg/types/basics`

gohandlers provides a standard set of reusable wrappers in the [`types/basics`](https://github.com/ufukty/gohandlers/tree/main/pkg/types/basics) package:

-   `types.String`: For parsing strings with length/regex rules
-   `types.Int`: For parsing integers with min/max
-   `types.Boolean`: For boolean values (`true`, `false`, `1`, `0`)
-   `types.Time`: For parsing RFC3339 timestamps

---

### Example: Bounded Integer

```go
type ListPetsRequest struct {
  Limit types.Int `query:"limit"`
}

func init() {
  types.IntRules["limit"] = types.IntRule{
    Min: 1,
    Max: 100,
  }
}
```

gohandlers will enforce that `?limit=0` returns:

```json
{ "limit": "value must be at least 1" }
```

These types make it trivial to define reusable validation across many endpoints.

---

## 🔗 Cross-Field Validation

While field-level checks cover most use cases, you can also define cross-field validation manually in your `Validate()` method.

```go
func (req CreateBookingRequest) Validate() map[string]error {
  errs := map[string]error{}
  if req.StartTime.After(req.EndTime) {
    errs["endTime"] = errors.New("endTime must be after startTime")
  }
  return errs
}
```

You can combine generated validation with your own logic. If you use custom types on each field, gohandlers will still generate the wrapper that aggregates them.

---

## 🛑 Handling Optional Fields

If you use a pointer field (`*int`, `*string`, etc.), you can determine if a value was supplied and validate accordingly.

```go
type SearchRequest struct {
  Query *string `query:"q"`
  Page  *int    `query:"page"`
}
```

In your `Validate()` method:

```go
if req.Page != nil && *req.Page < 1 {
  errs["page"] = errors.New("page must be >= 1")
}
```

This gives you full control over optional vs. required semantics.

---

## 🧪 Testing Validation

Because all validation lives in `Validate()` methods, it’s easy to unit test:

```go
func TestInviteUserValidation(t *testing.T) {
  req := InviteUserRequest{Email: "invalid-email"}
  errs := req.Validate()
  if errs["email"] == nil {
    t.Fatal("expected email validation error")
  }
}
```

This separation also allows you to reuse request types across HTTP and non-HTTP contexts, like CLI tools or background jobs.

---

## ✅ Best Practices

-   Use custom types to encapsulate parsing + validation logic.
-   Validate required fields explicitly using `Validate()`.
-   Prefer `map[string]error` for clear, field-level error feedback.
-   Use pointers for optional fields.
-   Adopt `types/basics` for quick rules like min/max, pattern matching, etc.
-   Keep validation logic _inside_ request structs—not your handler.

---

## 🔚 Summary

With gohandlers, validation is:

-   **Automatic:** Generated for each field based on type and presence.
-   **Extendable:** Easy to override or enhance with custom logic.
-   **Encapsulated:** Lives alongside your request structs or custom types.
-   **Testable:** Clean, consistent, and independently testable.

By combining generated validators with domain-specific custom types, gohandlers helps you write APIs that are not only easy to build—but also hard to break.

Validation becomes something you **declare**, not something you **repeat**.

✨ Happy validating!
