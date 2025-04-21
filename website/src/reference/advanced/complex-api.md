> **PREVIEW**
>
> Some of the functionality in this post are in consideration and is not currently under development.

# Managing Complex APIs: Advanced Tag Usage in gohandlers

**gohandlers** greatly simplifies building Go APIs by generating boilerplate code based on simple struct tags. For basic endpoints, the default tags (`route`, `query`, `json`) cover most cases effortlessly. But as your API grows more sophisticated, you'll need to manage complex scenarios like deeply nested JSON, optional query parameters, conditional validation, or dynamic routes.

This article guides you through powerful techniques and advanced tag usage to elegantly handle these more intricate API requirements.

## 🏷 Quick Refresher: Standard Tags

Before we dive into advanced usage, let's briefly revisit the basic tags:

-   **`route`**: Pulls data from URL path parameters.
-   **`query`**: Extracts values from URL query strings.
-   **`json`**: Serializes/deserializes JSON payloads.
-   **`form`**: Handles form-encoded data (`application/x-www-form-urlencoded`).

```go
type UpdateUserRequest struct {
  ID    string `route:"userId"`    // /users/{userId}
  Name  string `json:"name"`
  Email string `json:"email"`
}
```

These tags handle the majority of common use cases clearly and efficiently. Now, let's explore advanced usage scenarios.

## 🚦 Handling Optional Query Parameters

In more complex APIs, you might have optional query parameters that require custom handling.

**Solution:** Use pointers or custom types to indicate optional parameters.

```go
type ListPetsRequest struct {
  Limit  *int    `query:"limit"`  // optional parameter
  SortBy *string `query:"sortBy"`
}
```

-   If the query param is omitted, the field remains `nil`.
-   You can provide default values in your handler logic or via custom validation methods.

### Recommended pattern:

```go
func (req *ListPetsRequest) Validate() map[string]error {
  errs := map[string]error{}

  if req.Limit == nil {
    defaultLimit := 10
    req.Limit = &defaultLimit
  } else if *req.Limit <= 0 {
    errs["limit"] = errors.New("limit must be greater than zero")
  }

  return errs
}
```

## 🌳 Deeply Nested JSON Structures

Complex APIs frequently use nested structures. gohandlers supports nested JSON structs seamlessly:

```go
type CreateOrderRequest struct {
  CustomerInfo struct {
    Name  string `json:"name"`
    Email string `json:"email"`
  } `json:"customerInfo"`

  Items []struct {
    ProductID string `json:"productId"`
    Quantity  int    `json:"quantity"`
  } `json:"items"`
}
```

gohandlers automatically generates correct parsing and serialization logic for deeply nested structures without any extra effort from you.

## 🌀 Arrays & Collections in Query Parameters

Sometimes, APIs use arrays or multiple query params:

```
GET /pets?tag=cat&tag=dog&tag=hamster
```

**Solution:** Use slices in your struct.

```go
type ListPetsByTagsRequest struct {
  Tags []string `query:"tag"`
}
```

gohandlers automatically captures multiple values and populates your slice correctly.

## 🔍 Custom Tag Parsing with Domain-Specific Types

For custom parsing logic, define your own domain-specific types with custom parsing methods.

```go
type PetID string

func (p *PetID) FromRoute(s string) error {
  if len(s) != 8 {
    return errors.New("PetID must be exactly 8 characters")
  }
  *p = PetID(s)
  return nil
}

type GetPetRequest struct {
  ID PetID `route:"id"`
}
```

gohandlers recognizes these interfaces:

-   `FromRoute(string) error`
-   `FromQuery(string) error`

and automatically invokes your parsing logic, enabling precise validation.

## 🚧 Conditional & Cross-Field Validation

Advanced APIs often require validation rules that span multiple fields.

**Best Practice:** Define a custom `Validate()` method for your request struct.

```go
type CreateEventRequest struct {
  StartTime string `json:"startTime"`
  EndTime   string `json:"endTime"`
}

func (req *CreateEventRequest) Validate() map[string]error {
  errs := map[string]error{}

  start, err1 := time.Parse(time.RFC3339, req.StartTime)
  end, err2 := time.Parse(time.RFC3339, req.EndTime)

  if err1 != nil {
    errs["startTime"] = errors.New("invalid startTime format")
  }
  if err2 != nil {
    errs["endTime"] = errors.New("invalid endTime format")
  }

  if err1 == nil && err2 == nil && !start.Before(end) {
    errs["endTime"] = errors.New("endTime must be after startTime")
  }

  return errs
}
```

This ensures robust, clear, and maintainable validation logic even for complex rules.

## 🗺️ Dynamic & Complex Route Patterns

For dynamic routes, you may have multiple route parameters. gohandlers handles these gracefully:

```go
// Endpoint: /users/{userId}/pets/{petId}
type UpdateUserPetRequest struct {
  UserID string `route:"userId"`
  PetID  string `route:"petId"`
  Name   string `json:"name"`
}
```

All parameters specified in the path automatically map directly into your struct fields.

## 🚩 Handling Form-Encoded Data

For APIs accepting form data (`application/x-www-form-urlencoded`):

```go
type LoginRequest struct {
  Username string `form:"username"`
  Password string `form:"password"`
}
```

gohandlers auto-generates correct logic to parse and handle form data efficiently.

## 📌 Mixing Data Sources Gracefully

You might combine multiple data sources (route, query, json, form):

```go
type ComplexSearchRequest struct {
  UserID string   `route:"userId"`
  Tags   []string `query:"tag"`
  Filters struct {
    Status string `json:"status"`
    Limit  int    `json:"limit"`
  } `json:"filters"`
}
```

gohandlers seamlessly handles multiple simultaneous sources, keeping your code clean and declarative.

## 🌐 Tagging Fields for Explicit Clarity

Sometimes, making your tags explicit improves readability and clarity:

```go
type ExplicitExample struct {
  ID        string `json:"id,omitempty"`
  IsVisible bool   `json:"isVisible" query:"visible"`
}
```

Here, `IsVisible` can be populated either from a JSON body or from a query param, depending on context. Explicit tagging clearly documents your intentions.

## 🚨 Advanced Troubleshooting: Tag Conflicts & Priorities

In rare cases, you may encounter tag conflicts (e.g., both `json` and `form` tags on the same field). gohandlers resolves these based on context:

-   **Request Parsing:** Prioritizes `route` > `query` > `form` > `json`.
-   **Response Serialization:** Primarily uses `json`.

Avoid ambiguous tagging by clearly defining fields according to their primary source or target.

## 🎯 Summary: Best Practices for Complex Tagging

-   Use **pointer fields** to represent optional parameters.
-   Embrace **custom types** for precise validation logic.
-   Prefer clear, simple tag definitions—avoid ambiguity.
-   Use custom `Validate()` methods to handle advanced validation logic.
-   Leverage gohandlers' automatic support for nested structures and arrays.

By mastering advanced tag usage, gohandlers lets you manage complexity gracefully and keeps your API maintainable, clear, and robust as your project evolves.

Happy API-building! 🌟
