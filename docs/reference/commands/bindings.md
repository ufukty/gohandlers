# ✱ `bindings`

Generates a Go file (default `bindings.gh.go`) containing `Build`, `Parse` and `Write` methods for each “binding” struct in your code. **Binding structs** are simply your request and response types for handlers that contains route, query, form or json values as fields.

-   **`Parse` Methods:** For each request binding (suffix `Request`) and response binding (suffix `Response`), a method is generated to populate that struct from an `*http.Request` or `*http.Response`. This reads URL path parameters, query parameters, form data, or JSON bodies as needed, and converts them to the correct types:

    ```go
    func (req *XRequest) Parse(r *http.Request) error    // populates XRequest from HTTP request
    func (res *XResponse) Parse(r *http.Response) error  // populates XResponse from HTTP response
    ```

    _Example:_ If `XRequest` has a field tagged `route:"id"`, the generated `Parse` will extract `id` from the URL path. Fields tagged `query:"q"` come from `r.URL.Query()`, and `json:"field"` from the JSON body, etc.. This ensures each part of the request (route, query, form, JSON) is handled appropriately, and errors (like missing or invalid values) are propagated.

-   **`Build` / `Write` Methods:** For request bindings, `Build` creates an `*http.Request` ready to send (with method, URL, query params, JSON body, etc.). For response bindings, `Write` writes the data to an `http.ResponseWriter` (sets headers like Content-Type, writes the body):

    ```go
    func (req XRequest) Build(host string) (*http.Request, error)
    func (res XResponse) Write(w http.ResponseWriter) error
    ```

    When calling `Build`, you provide a `host` (because Go’s `http.NewRequest` needs a URL). The returned request is ready to be sent (with all fields serialized properly). Similarly, `Write` on a response binding will serialize that struct (as JSON, form data, etc.) and write to the HTTP response in a handler context.

## What it solves?

Writing request parsing and response writing logic for each handler can be tedious and error-prone. Tags in your struct (like `route:"id"` or `json:"name"`) guide gohandlers to generate this logic, so you don’t have to write the same code repeatedly. This also ensures that if your types change (e.g., you add a new query param), the parsing/building logic updates on regeneration.

## Usage

```sh
# gohandlers bindings --help
Usage of bindings:
  -dir string
        the source directory contains Go files
  -out string
        the output file that will be created in -dir (default "bindings.gh.go")
  -recv string
        only use request types that is prefixed with handlers defined on this type
  -v    prints additional information
```

## Example

Suppose we have a handler and types in [handlers/pets/create.go](https://github.com/ufukty/gohandlers-petstore/handlers/pets/create.go):

```go
type CreatePetRequest struct {
  Name types.PetName `json:"name"` // from JSON body
  Tag  types.PetTag  `json:"tag"`  // from JSON body
}

type CreatePetResponse struct {
  ID string `json:"id"` // to JSON body of response
}

// CreatePet is an HTTP handler
func (p *Pets) CreatePet(w http.ResponseWriter, r *http.Request) {
  _ = &CreatePetRequest{}
  _ = &CreatePetResponse{}
  // Handler logic...
}
```

Run gohandlers for bindings in the [`handlers/pets`](https://github.com/ufukty/gohandlers-petstore/tree/main/handlers/pets) directory:

```bash
gohandlers bindings -dir handlers/pets -out bindings.gh.go
```

This generates [**`handlers/pets/bindings.gh.go`**](https://github.com/ufukty/gohandlers-petstore/blob/main/handlers/pets/bindings.gh.go) with those:

```go
func (bq GetPetRequest) Build(host string) (*http.Request, error)
func (bq *GetPetRequest) Parse(rq *http.Request) error
func (bs GetPetResponse) Write(w http.ResponseWriter) error
func (bs *GetPetResponse) Parse(rs *http.Response) error
```

Now your handler `CreatePet` can use `bq.Parse(r)` to [parse inputs](https://github.com/ufukty/gohandlers-petstore/blob/280eff72d24d32f5d61b32361653de906cd639bd/handlers/pets/create.go#L21), and `bs.Write(w)` to [write outputs](https://github.com/ufukty/gohandlers-petstore/blob/280eff72d24d32f5d61b32361653de906cd639bd/handlers/pets/create.go#L32), with all the heavy lifting done by gohandlers-generated code. For client-side, a `CreatePetRequest.Build()` will help [create requests](https://github.com/ufukty/gohandlers-petstore/blob/280eff72d24d32f5d61b32361653de906cd639bd/client/client.gh.go#L28) and `CreatePetResponse.Parse()` will help [deserializing responses](https://github.com/ufukty/gohandlers-petstore/blob/280eff72d24d32f5d61b32361653de906cd639bd/client/client.gh.go#L40) before returning to caller:

```go
type CreatePetRequest struct {
  Name types.PetName `json:"name"`
  Tag  types.PetTag  `json:"tag"`
}

type CreatePetResponse struct {
  ID string `json:"id"`
}

func (p *Pets) CreatePet(w http.ResponseWriter, r *http.Request) {
  bq := &CreatePetRequest{}

  if err := bq.Parse(r); err != nil {
    //
  }

  if err := bq.Validate(); err != nil {
    //
  }

  // Handler logic

  bs := &CreatePetResponse{}
  if err := bs.Write(w); err != nil {
    //
  }
}
```
