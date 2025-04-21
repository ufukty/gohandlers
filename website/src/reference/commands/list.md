# `list`

Generates a Go file (default `list.gh.go`) with a function or methods to **list all handlers and their metadata** (like method and path). This is extremely useful to automatically register your handlers with a router (e.g., `http.ServeMux` or any router library) without manually writing each route.

What it generates:

- **Global `ListHandlers` function:** If you have any free functions (not methods) that serve as handlers, it generates:

  ```go
  func ListHandlers() map[string]HandlerInfo { ... }
  ```

  where `HandlerInfo` is a struct containing at least `Method`, `Path`, and `Ref` (the handler function). The map key is usually a unique name for the handler (like `"CreatePet"`).

- **`ListHandlers` methods on receiver types:** If you have handlers defined as methods on structs (e.g., `func (p *Pets) CreatePet(...)`), it will generate a `ListHandlers()` method for each such receiver type:

  ```go
  func (p *Pets) ListHandlers() map[string]HandlerInfo { ... }
  ```

  This map will include entries for all handlers that have receiver `*Pets`.

- These functions gather **method and path** from the handlers (using gohandlers’ knowledge of HTTP method and path assignments, see below) and the handler function itself.

- **Custom HandlerInfo:** By default, gohandlers defines its own `HandlerInfo` in each generated file. However, if you want to use a shared type (perhaps your project defines a global route struct), you can provide `-hi-import "myapp/router"` and `-hi-type "HandlerInfo"` flags. Then gohandlers will import your package and use `myapp/router.HandlerInfo` instead in the return type. This can simplify integrating with your router setup.

## What it solves?

In a growing API, it’s easy to forget to register a handler or to mismatch the route path. With `ListHandlers`, you ensure every handler is accounted for. For example:

## Usage

```sh
# gohandlers list --help
Usage of list:
  -dir string
        the directory contains Go files. one handler and a request binding type is allowed per file
  -hi-import string
        the package to import for custom implementation of HandlerInfo
  -hi-type string
        the string to be substituted with mentions of HandlerInfo
  -out string
        output file that will be generated in the 'dir' (default "list.gh.go")
  -v    prints additional information
```

## Example

If [`handlers/pets`](https://github.com/ufukty/gohandlers-petstore/tree/main/handlers/pets) directory has handlers on `*Pets` receiver, after running:

```bash
gohandlers list -dir handlers/pets -out list.gh.go
```

You [get](https://github.com/ufukty/gohandlers-petstore/blob/main/handlers/pets/list.gh.go) something like:

```go
type HandlerInfo struct {
  Method string
  Path   string
  Ref    http.HandlerFunc
}

func (pe *Pets) ListHandlers() map[string]HandlerInfo {
  return map[string]HandlerInfo{
    "CreatePet": {Method: "POST", Path: "/create-pet", Ref: pe.CreatePet},
    "DeletePet": {Method: "DELETE", Path: "/pets/{id}", Ref: pe.DeletePet},
    "GetPet":    {Method: "GET", Path: "/pets/{id}", Ref: pe.GetPet},
    "ListPets":  {Method: "GET", Path: "/pets", Ref: pe.ListPets},
  }
}
```

_(Here `HandlerInfo` is a type provided by gohandlers by default. It has `Method, Path, Ref` fields as shown.)_

To register these with a router, you could do:

```go
func main() {
  pets := pets.New()
  s := http.NewServeMux()
  for name, handler := range pets.ListHandlers() {
    pattern := fmt.Sprintf("%s %s", handler.Method, handler.Path)
    fmt.Println("registering", name, "as", pattern)
    s.HandleFunc(pattern, handler.Ref)
  }
}
```

This loops through the map and registers each handler with its path and method. If you’re using the standard `http.ServeMux`, it doesn’t support methods directly, but you can still use the path and attach the handler (method filtering would be inside handlers or by using another router library).
