# 🧶 Bindings and Tags

At the heart of **Gohandlers** is a simple, powerful idea: your Go structs define your API.

Instead of writing verbose parsing logic, Gohandlers inspects your `Request` and `Response` structs—using their names, field types, and tags—to generate the boilerplate that turns raw HTTP requests into typed Go values, and back again.

In this article, you’ll learn how to design effective binding structs, how tags map fields to HTTP inputs/outputs, and how to unlock Gohandlers’ full potential with clean, declarative type definitions.

---

## 📦 What are binding structs?

**Binding structs** are plain Go types that represent the structure of HTTP requests and responses in your application.

There are two main types:

-   `...Request` structs  
    Used to bind data **from** HTTP requests into typed Go values.
-   `...Response` structs  
    Used to serialize Go values **into** HTTP responses.

Gohandlers looks for these by naming convention. If you define a handler named `CreatePet`, it expects:

```go
type CreatePetRequest struct { ... }
type CreatePetResponse struct { ... }
```

These two structs become the source of truth for generating:

-   Parsing code for reading from `*http.Request`
-   Validation helpers
-   Serialization code for writing to `http.ResponseWriter`
-   Client methods and mock interfaces

---

## 🏷 Tags: Declaring data source

Tags tell Gohandlers how to connect struct fields to HTTP components. Each field should include one of the following:

| Tag     | Used In            | Purpose                             |
| ------- | ------------------ | ----------------------------------- |
| `route` | Request structs    | Maps to a path parameter in the URL |
| `query` | Request structs    | Reads from URL query string         |
| `form`  | Request structs    | Reads from form-encoded body        |
| `json`  | Request & Response | Parses/serializes JSON body         |

---

## ⭐️ Typical API request

Here is an example:

```go
type GetPetRequest struct {
  ID basics.String `route:"id"` // /pets/{id}
}

type CreatePetRequest struct {
  Name string `json:"name"` // JSON body
  Tag  string `json:"tag"`
}

type ListPetsRequest struct {
  Limit basics.Int `query:"limit"` // ?limit=10
}
```

---

## 🚦 Route parameters

For data in the URL path like `/users/{userId}`:

```go
type GetUserRequest struct {
  UserID basics.String `route:"userId"` // maps /users/{userId}
}
```

The field name doesn’t have to match the tag—it just needs to match the placeholder in the URL path template.

Gohandlers will replace or extract values from the URL automatically during parsing/building.

---

## 🧭 Query parameters

Query parameters come from the URL’s `?key=value` section:

```go
type ListOrdersRequest struct {
  Page  basics.Int    `query:"page"`
  Sort  basics.String `query:"sort"`
  Limit basics.Int    `query:"limit"`
}
```

Gohandlers gets these values from `r.URL.Query()` and passes to `.FromQuery()` methods of them.

---

## 🧾 Form fields

For APIs that accept `application/x-www-form-urlencoded` or form submissions:

```go
type LoginRequest struct {
  Username basics.String `form:"username"`
  Password basics.String `form:"password"`
}
```

Gohandlers will automatically call `r.FromForm()` and extract the fields accordingly.

---

## 🧱 JSON fields

Use `json` tags for both incoming request bodies and outgoing response payloads:

```go
type CreatePostRequest struct {
  Title string `json:"title"`
  Body  string `json:"body"`
}

type CreatePostResponse struct {
  ID string `json:"id"`
}
```

JSON tags work just like in the standard `encoding/json` package, as Gohandlers only uses them to decide whether to call JSON encoder/decoder inside (de)serialization helpers.

---

## 🎨 Combining tags

Each field should have **one primary source**, but you can mix tags across fields:

```go
type UpdatePetRequest struct {
  PetID basics.String `route:"petId"` // path param
  Age   basics.Int    `query:"age"`   // query param
  Name  string        `json:"name"`   // body param
}
```

In this example:

-   `PetID` is extracted from the URL (`/pets/{petId}`)
-   `Name` comes from the JSON body
-   `Age` is parsed from `?age=3`

Gohandlers will generate code to assemble the entire struct from these parts—no manual parsing required.

---

## 🔍 Domain-specific types

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

Gohandlers recognizes these interfaces:

-   `FromRoute(string) error`
-   `FromQuery(string) error`

and automatically invokes your parsing logic, enabling precise validation.

---

## ✅ Best practices

-   Use **struct names** starting with the handler name and ending in `Request` or `Response`.
-   Tag **every field** with its source (`route`, `query`, `json`, `form`).
-   Keep each field’s source unambiguous (don’t tag one field with multiple input types).
-   Use **custom types** to encapsulate validation logic and parsing.

---

## 🧭 Summary

Binding structs and their tags are the foundation of Gohandlers. They let you describe your API contract clearly, declaratively, and type-safely—without any runtime reflection or handwritten parsing code.

With just a few tagged structs, Gohandlers will generate:

-   Request parsers
-   Response writers
-   Field validators
-   Typed clients
-   Mock implementations

Your API becomes easier to maintain, safer to extend, and faster to develop.

Start with the tags—let Gohandlers handle the rest. 🚀
