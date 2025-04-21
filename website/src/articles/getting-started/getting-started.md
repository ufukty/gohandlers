# Getting Started with gohandlers: Build HTTP APIs in Go—Effortlessly!

Building HTTP APIs in Go is powerful, but repetitive boilerplate often sneaks into your handlers, validation, and routing setup. Wouldn't it be great if you could automate away that tedious code?

Meet **gohandlers**, your new best friend for creating robust, maintainable, and boilerplate-free Go APIs.

This comprehensive, friendly guide will help you quickly and confidently get started with gohandlers, step-by-step. Let's dive right in!

## ✨ What is gohandlers, Anyway?

**gohandlers** is a code-generation tool designed to remove boilerplate when building HTTP handlers in Go. It automatically generates serialization, validation, routing, clients, and mocks—freeing you to focus purely on your business logic.

Here's the beauty of gohandlers at a glance:

-   **Automated Bindings:** Parses requests and writes responses.
-   **Automatic Validation:** Generates field-level validation effortlessly.
-   **Simplified Routing:** Automatically registers all handlers.
-   **Typed Clients:** Gives you strongly typed clients for your API.
-   **Built-in Mocks:** Makes unit testing a breeze.

Sounds good? Let’s set it up!

## 🚧 Installation

Install the gohandlers CLI with one simple Go command:

```bash
go install github.com/ufukty/gohandlers/cmd/gohandlers@latest
```

This gives you immediate access to the `gohandlers` command-line tool.

Ensure your Go environment (`$GOPATH/bin`) is set in your system’s `PATH`. Check your installation with:

```bash
gohandlers
```

If you see list of available commands, you're all set!

## 🗂️ Set Up Your Project Structure

Let's create a minimal project to see gohandlers in action:

```bash
mkdir petstore && cd petstore
go mod init example.com/petstore
```

Let’s assume you're creating a simple API for a pet store. Your directory might look like this:

```
petstore/
├── handlers
│   └── pets
│     ├── create.go
│     ├── get.go
│     ├── list.go
│     └── delete.go
├── main.go
└── go.mod
```

Each handler file will define request and response structs, along with handler logic.

## 🧩 Define Your Request and Response Structs

Create simple Go structs with clear field tags to describe your HTTP inputs and outputs:

-   Use `json` tags for request/response bodies.
-   Use `route` and `query` tags for URL parameters (we'll see these next!).

```go
// handlers/pets/create.go
package pets

type CreatePetRequest struct {
  Name string `json:"name"`
  Tag  string `json:"tag"`
}

type CreatePetResponse struct {
  ID string `json:"id"`
}

// Pets is your handler receiver (could be DB or service)
type Pets struct{}

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

  // Dummy ID generation
  id := "12345"
  resp := &CreatePetResponse{ID: id}

  if err := resp.Write(w); err != nil {
    http.Error(w, "failed to write response", http.StatusInternalServerError)
  }
}
```

No parsing, validation, or serialization logic is needed—gohandlers handles it automatically.

## 🚀 Generate Your HTTP Bindings

Run the `bindings` command to auto-generate serialization/deserialization code:

```bash
cd handlers/pets
gohandlers bindings --dir . --out bindings.gh.go --recv Pets --v
```

The magic happens automatically, producing methods like:

```go
func (req CreatePetRequest) Parse(r *http.Request) error { /* ... */ }
func (resp CreatePetResponse) Write(w http.ResponseWriter) error { /* ... */ }
```

Your handlers can now directly parse requests and write responses effortlessly!

## 🛡️ Add Effortless Validation

gohandlers also automates validation generation:

```bash
gohandlers validate --dir . --out validate.gh.go --recv Pets --v
```

You'll get field-level validators like:

```go
func (req CreatePetRequest) Validate() map[string]error {
  errs := map[string]error{}
  if req.Name == "" {
    errs["name"] = errors.New("Name cannot be empty")
  }
  return errs
}
```

Use these validators inside your handlers to quickly return precise validation errors.

## 🎯 Write Simple, Consistent Handlers

Here’s your handler, simplified and elegant:

```go
func (p *Pets) CreatePet(w http.ResponseWriter, r *http.Request) {
  req := &CreatePetRequest{}

  if err := req.Parse(r); err != nil {
    //
  }

  if errs := req.Validate(); len(errs) > 0 {
    //
  }

  id := uuid.New().String() // Simulate DB insertion
  resp := &CreatePetResponse{ID: id}

  if err := resp.Write(w); err != nil {
    //
  }
}
```

Your handler is clean, readable, and easy to maintain!

## 🗺️ Automatic Handler Registration

Stop manually registering your routes! Instead, use the auto-generated handler listing:

```bash
gohandlers list --dir . --out list.gh.go --recv Pets --v
```

Your generated `ListHandlers()` method lets you effortlessly wire everything into your HTTP server:

```go
func main() {
  pets := pets.NewPetsHandler()

  mux := http.NewServeMux()
  for _, h := range pets.ListHandlers() {
    mux.HandleFunc(h.Path, h.Ref)
  }

  log.Println("Starting server on :8080...")
  http.ListenAndServe(":8080", mux)
}
```

Voila! All your handlers are registered automatically.

## 🏁 Run Your Server

Start your server:

```bash
go run .
```

You should see:

```
Listening on :8080...
```

Test your handler using `curl` or any REST client:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"Fluffy","tag":"cat"}' \
  http://localhost:8080/create-pet
```

You'll receive a JSON response:

```json
{ "id": "12345" }
```

Congrats! You've successfully set up your first gohandlers-based API endpoint.

## 🎁 Generate Typed API Clients (Bonus!)

Let’s not forget clients. Generate strongly typed Go clients for your API consumers:

```bash
gohandlers client --dir handlers/pets --out client.gh.go --pkg client --v
```

Your consumers will love how easy it is to interact with your API:

```go
client := client.NewClient(client.StaticPool("http://localhost:8080"))
resp, err := client.CreatePet(ctx, dto.Pet{Name: "Buddy", Tag: "dog"})
```

## 🧪 Mock Clients for Simple Testing

Testing your handlers couldn't be easier:

```bash
gohandlers mock --dir handlers/pets --out mock.gh.go --pkg client --v
```

Now you can mock responses effortlessly:

```go
mock := &client.MockClient{}
mock.CreatePetFunc = func(ctx context.Context, dto dto.Pet) (*CreatePetResponse, error) {
  return &CreatePetResponse{ID: "test123"}, nil
}
```

Your unit tests become simple, predictable, and maintainable!

## 🎯 Final Thoughts: Why Use gohandlers?

By adopting gohandlers, you're making your life easier in so many ways:

-   **No more tedious boilerplate:** Save hours of repetitive coding.
-   **Clean separation of concerns:** Focus purely on your business logic.
-   **Consistent patterns:** Predictable and readable handlers.
-   **Strongly-typed APIs:** Avoid runtime errors and mistakes.
-   **Built-in testing support:** Effortlessly test your business logic with mocks.

You're now fully equipped to use gohandlers, saving yourself time, complexity, and headaches—making Go HTTP APIs delightful again.

Happy coding! 🚀
