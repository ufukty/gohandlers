# ✱ `yaml`

Generates a YAML file (default `gh.yml`) that lists handlers, their HTTP methods, and paths. This is like a lightweight documentation or can be seen as a mini OpenAPI output. It’s useful if other services or tools (not written in Go) need to know about your API.

The YAML might look like:

```yaml
CreatePet:
    method: POST
    path: /create-pet
DeletePet:
    method: DELETE
    path: /pets/{id}
GetPet:
    method: GET
    path: /pets/{id}
ListPets:
    method: GET
    path: /pets
```

_(The exact structure can be observed in the petstore example in the repository.)_

Other tools or documentation generators could use this YAML to produce API docs or stubs in other languages if needed.

---

## What it solves?

While gohandlers removes the need to manually maintain OpenAPI specs, sometimes you still want to share the API structure with others who aren’t reading your Go code. The `yaml` output is a quick way to export the essential info (endpoints and methods) in a language-agnostic format.

---

## Usage

```sh
# gohandlers yaml --help
Usage of yaml:
  -dir string
        the directory contains Go files. one handler and a request binding type is allowed per file
  -out string
        yaml file that will be generated in the 'dir' (default "gh.yml")
  -v    prints additional information
```

---

## Example

You provide the source directory contain your Go handlers and filename that will be created in that directory:

```bash
gohandlers yaml -dir handlers/pets -out gh.yml
```

After this, you will get the yaml file just like the one above.
