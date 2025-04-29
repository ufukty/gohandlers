# ✱ `mock`

Generates a Go file (default `mock.gh.go`) with two things:

-   An **interface** that your `Client` implements (all the methods generated in `client`).
-   A **mock client struct** that also implements this interface, but allows you to customize behavior for testing.

---

## Why `mock`?

When writing unit tests for your application, you might not want to make real HTTP calls. The mock client has stubbed methods (you can replace them) to simulate responses. This way, your service code can depend on an interface (abstract client) and in tests you inject the mock.

---

## Usage

```sh
# gohandlers mock --help
Usage of mock:
  -dir string
        input directory
  -import string
        the import path of package declares binding types
  -out string
        output directory
  -pkg string
        package name for the generated file
  -v    prints additional information
```

---

## Example

Running `gohandlers mock` similarly to `client`:

```bash
gohandlers mock \
  -dir handlers/pets \
  -out client/mock.gh.go \
  -pkg client \
  -import "github.com/yourusername/yourrepo/handlers/pets"
```

Generates a [`client/mock.gh.go`](https://github.com/ufukty/gohandlers-petstore/blob/main/client/mock.gh.go) with content like:

```go
type Interface interface {
  CreatePet(*pets.CreatePetRequest) (*pets.CreatePetResponse, error)
  DeletePet(*pets.DeletePetRequest) (*http.Response, error)
  GetPet(*pets.GetPetRequest) (*pets.GetPetResponse, error)
  ListPets(*pets.ListPetsRequest) (*pets.ListPetsResponse, error)
}

type Mock struct {}

func (m *Mock) CreatePet(*pets.CreatePetRequest) (*pets.CreatePetResponse, error) {
  return nil, nil
}

func (m *Mock) DeletePet(*pets.DeletePetRequest) (*http.Response, error) {
  return nil, nil
}

func (m *Mock) GetPet(*pets.GetPetRequest) (*pets.GetPetResponse, error) {
  return nil, nil
}

func (m *Mock) ListPets(*pets.ListPetsRequest) (*pets.ListPetsResponse, error) {
  return nil, nil
}

```
