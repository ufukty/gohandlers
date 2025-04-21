# Understanding Binding Structs and Tags in gohandlers

At the heart of **gohandlers** is a simple, powerful idea: your Go structs define your API.

Instead of writing verbose parsing logic, gohandlers inspects your `Request` and `Response` structs—using their names, field types, and tags—to generate the boilerplate that turns raw HTTP requests into typed Go values, and back again.

In this article, you’ll learn how to design effective binding structs, how tags map fields to HTTP inputs/outputs, and how to unlock gohandlers’ full potential with clean, declarative type definitions.

## 📦 What Are Binding Structs?

**Binding structs** are plain Go types that represent the structure of HTTP requests and responses in your application.

There are two main types:

-   `...Request` structs  
    Used to bind data **from** HTTP requests into typed Go values.
-   `...Response` structs  
    Used to serialize Go values **into** HTTP responses.

gohandlers looks for these by naming convention. If you define a handler named `CreatePet`, it expects:

```go
type CreatePetRequest struct { ... }
type CreatePetResponse struct { ... }
```

These two structs become the source of truth for generating:

-   Parsing code for reading from `*http.Request`
-   Validation helpers
-   Serialization code for writing to `http.ResponseWriter`
-   Client methods and mock interfaces

## 🏷 Field Tags: Declaring Where Data Comes From

Tags tell gohandlers how to connect struct fields to HTTP components. Each field should include one of the following:

| Tag     | Used In            | Purpose                             |
| ------- | ------------------ | ----------------------------------- |
| `route` | Request structs    | Maps to a path parameter in the URL |
| `query` | Request structs    | Reads from URL query string         |
| `form`  | Request structs    | Reads from form-encoded body        |
| `json`  | Request & Response | Parses/serializes JSON body         |

### Example: Typical API Request

```go
type GetPetRequest struct {
  ID string `route:"id"`           // /pets/{id}
}

type CreatePetRequest struct {
  Name string `json:"name"`        // JSON body
  Tag  string `json:"tag"`
}

type ListPetsRequest struct {
  Limit int `query:"limit"`        // ?limit=10
}
```

## 🚦 Route Parameters (`route:"..."`)

For data in the URL path like `/users/{userId}`:

```go
type GetUserRequest struct {
  UserID string `route:"userId"` // maps /users/{userId}
}
```

The field name doesn’t have to match the tag—it just needs to match the placeholder in the URL path template.

gohandlers will replace or extract values from the URL automatically during parsing/building.

## 🧭 Query Parameters (`query:"..."`)

Query parameters come from the URL’s `?key=value` section:

```go
type ListOrdersRequest struct {
  Page  int    `query:"page"`
  Sort  string `query:"sort"`
  Limit int    `query:"limit"`
}
```

gohandlers parses these values from `r.URL.Query()`, converting them to the correct types.

Use `[]string` or `[]int` to capture multiple values like `?tag=foo&tag=bar`.

## 🧾 Form Fields (`form:"..."`)

For APIs that accept `application/x-www-form-urlencoded` or form submissions:

```go
type LoginRequest struct {
  Username string `form:"username"`
  Password string `form:"password"`
}
```

gohandlers will automatically call `r.ParseForm()` and extract the fields accordingly.

## 🧱 JSON Fields (`json:"..."`)

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

JSON tags work just like in the standard `encoding/json` package—gohandlers uses those conventions directly for marshaling and unmarshaling.

## 🎨 Combining Tags

Each field should have **one primary source**, but you can mix tags across fields:

```go
type UpdatePetRequest struct {
  PetID string `route:"petId"`   // path param
  Name  string `json:"name"`     // body param
  Age   int    `query:"age"`     // query param
}
```

In this example:

-   `PetID` is extracted from the URL (`/pets/{petId}`)
-   `Name` comes from the JSON body
-   `Age` is parsed from `?age=3`

gohandlers will generate code to assemble the entire struct from these parts—no manual parsing required.

## 🧠 Custom Types & Interfaces

Want to validate or parse your fields with custom logic? Use your own types that implement:

-   `FromRoute(string) error`
-   `FromQuery(string) error`
-   `Validate() error`

Example:

```go
type Email string

func (e *Email) FromQuery(s string) error {
  if !strings.Contains(s, "@") {
    return errors.New("invalid email")
  }
  *e = Email(s)
  return nil
}

func (e Email) Validate() error {
  if len(e) < 5 {
    return errors.New("email too short")
  }
  return nil
}

type InviteRequest struct {
  Email Email `query:"email"`
}
```

gohandlers will automatically detect these interfaces and call them in the generated `Parse` and `Validate` methods.

## 🧪 Response Structs & Output Tags

Response structs work just like request structs, usually with JSON output:

```go
type GetPetResponse struct {
  Pet PetDTO `json:"pet"`
}
```

When you call `resp.Write(w)`, gohandlers serializes the response into JSON and sets the correct headers.

Form-encoded output is also supported (with `form` tags), but most APIs use JSON.

## ✅ Best Practices

-   Use **meaningful struct names** ending in `Request` or `Response`.
-   Tag **every field** with its source (`route`, `query`, `json`, `form`).
-   Keep each field’s source unambiguous (don’t tag one field with multiple input types).
-   Use **custom types** to encapsulate validation logic and parsing.
-   Prefer **pointers** for optional values like `*int`, `*string`.

## 🧭 Summary

Binding structs and their tags are the foundation of gohandlers. They let you describe your API contract clearly, declaratively, and type-safely—without any runtime reflection or handwritten parsing code.

With just a few tagged structs, gohandlers will generate:

-   Request parsers
-   Response writers
-   Field validators
-   Typed clients
-   Mock implementations

Your API becomes easier to maintain, safer to extend, and faster to develop.

Start with the tags—let gohandlers handle the rest. 🚀
