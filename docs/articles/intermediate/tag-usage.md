> **PREVIEW**
>
> Some of the functionality in this post are in consideration and is not currently under development.

# 🏷️ Tag Usage

**gohandlers** greatly simplifies building Go APIs by generating boilerplate code based on simple struct tags. For basic endpoints, the default tags (`route`, `query`, `json`) cover most cases effortlessly. But as your API grows more sophisticated, you'll need to manage complex scenarios like deeply nested JSON, optional query parameters, conditional validation, or dynamic routes.

This article guides you through powerful techniques and advanced tag usage to elegantly handle these more intricate API requirements.

## 👟 Quick Refresher: Standard Tags

Before we dive into advanced usage, let's briefly revisit the basic tags:

-   **`route`**: Pulls data from URL path parameters.
-   **`query`**: Extracts values from URL query strings.
-   **`json`**: Serializes/deserializes JSON payloads.
-   **`form`**: Gets data from form fields.

```go
type UpdateUserRequest struct {
  ID    string `route:"userId"` // /user/{userId}
  Name  string `json:"name"`
  Email string `json:"email"`
}
```

These tags handle the majority of common use cases clearly and efficiently. Now, let's explore advanced usage scenarios.

## 🚦 Handling Optional Query Parameters

In more complex APIs, you might have optional query parameters that require custom handling.

```go
type ListPetsRequest struct {
  Limit  int    `query:"limit"`
  SortBy string `query:"sortBy"`
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

As gohandlers only calls the JSON encoder/decoder inside helpers when it sees a field with `json` tag, you can keep defining your binding types as you do it with JSON encoder.

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
