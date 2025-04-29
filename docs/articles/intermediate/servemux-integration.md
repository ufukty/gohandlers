DONE:

# 🧇 `ServeMux` Integration

Routing in Go web applications can quickly become a tangled web of `s.HandleFunc(...)` calls, especially as your API grows. Keeping route paths, handler names, and HTTP methods in sync across documentation and implementation can become a serious source of bugs and drift.

**Gohandlers** eliminates this problem by embracing a **metadata-driven approach** to routing—automatically generating handler registrations based on your code’s structure and field tags. In this article, we’ll walk through how it works, how to use it, and how it keeps your server and docs perfectly in sync.

---

## 🎯 The Problem with Manual Routing

Here’s what typical manual routing might look like:

```go
s := http.NewServeMux()
s.HandleFunc("GET /pets", listPetsHandler)
s.HandleFunc("GET /pets/{id}", getPetHandler)
s.HandleFunc("POST /pets", getPetHandler) // ⚠️ wrong handler
```

Issues this introduces:

-   Easy to duplicate or misconfigure paths and methods
-   No link between the URL and the handler’s logic or data model
-   Hard to discover or document routes
-   No enforcement of consistency between handlers and their metadata

---

## ✅ The Gohandlers Solution

Gohandlers generates routing metadata automatically by inspecting:

-   Handler names  
    (e.g., `GetPet`, `ListPets`)

-   Associated request/response structs  
    (`GetPetRequest`, `ListPetsResponse`, etc.)

-   Field tags of associated request/response structs,  
    (eg. `route`, `query`, `form` and `json`)

-   Doc comments.  
    (eg. `// GET /pets/{id}`)

From this, it generates: a file with functions and methods that returns all your handlers as maps. Thus, everything stays in sync with your real code.

---

## 🦺 Automatic Routing Benefits

-   **No manual route wiring:** Paths and handlers stay in sync with your code.
-   **No route duplication bugs:** Each handler is defined only once.
-   **Cleaner `main.go`:** No clutter of `s.HandleFunc(...)` calls.
-   **Perfect doc generation:** Routes and methods are exported to a YAML file (`gh.yml`) for use in documentation or tooling.

---

## 🗂️ The `ListHandlers()` Function

For each group of handlers (e.g., all methods on a struct like `*Pets`), Gohandlers generates a `ListHandlers()` method:

```go
func (p *Pets) ListHandlers() map[string]HandlerInfo {
  return map[string]HandlerInfo{
  "CreatePet": {
    Method: "POST",
    Path:   "/pets",
    Ref:  p.CreatePet,
  },
  "GetPet": {
    Method: "GET",
    Path:   "/pets/{id}",
    Ref:  p.GetPet,
  },
  }
}
```

Each `HandlerInfo` struct includes:

-   `Method`: The HTTP method (e.g. `GET`, `POST`)
-   `Path`: The route pattern, with `{param}` placeholders
-   `Ref`: A reference to the actual handler function (`http.HandlerFunc`)

You can now wire up all routes in one consistent loop:

```go
s := http.NewServeMux()
for _, h := range myHandler.ListHandlers() {
  s.HandleFunc(h.Path, h.Ref)
}
```

---

## 🔍 How Metadata is Collected

Gohandlers parses your code and builds a model for each handler, including:

-   Handler name (e.g. `CreatePet`)
-   HTTP method (e.g. `POST`)
-   URL path (e.g. `/pets`)
-   Reference from the actual Go handler function

This metadata comes from:

| Source                  | What it provides                            |
| ----------------------- | ------------------------------------------- |
| Handler name            | Base name of the handler                    |
| Handler doc comment     | HTTP method and path if explicitly declared |
| Binding type field tags | URL path, URL query and body parameters     |

If no method or path is declared, Gohandlers uses naming conventions and your binding struct’s `route:` tags to derive them automatically.

---

## ✍️ Doc Comments

If you want more control, Gohandlers supports in-code comments above your handlers:

```go
// POST /pets
// gh:list
func (p *Pets) CreatePet(...) { ... }
```

This explicitly tells Gohandlers:

-   Use `POST` as the HTTP method
-   Use `/pets` as the path
-   Include this handler in `ListHandlers()`

This is useful when the default inference from naming isn’t sufficient or if you want to override it.

---

## 🧼 Best Practices

-   Use meaningful handler names (`GetPet`, `ListOrders`, etc.)
-   Tag route values with `route:"..."` to clarify URL parameters
-   Use structured comments to document custom routes and methods

---

## 🚀 Summary

Gohandlers’ metadata-driven routing and automatic registration eliminates the need to manually wire up routes. It:

-   Keeps your handlers and mux in sync
-   Reduces the potential for errors or duplication
-   Provides a clean, declarative routing setup
-   Enables easy documentation export via YAML

Whether your API has 5 endpoints or 500, Gohandlers ensures every route is consistent, discoverable, and connected to real code.

**One source of truth. Zero boilerplate. Fully synchronized.**

Happy routing! 🗺️
