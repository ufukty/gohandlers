# ✱ `validate`

Generates a Go file (default `validate.gh.go`) contains validation helpers for request binding types. Those helpers are meant to be called from handlers, after parsing is done. Validation methods call each field's validation method one by one, and collects all errors in a map. Then the handler can use its custom method of serialization based on its return type being JSON or HTML.

---

## Args

```sh
# gohandlers validate -help
Usage of validate:
  -dir string
        the directory contains Go files
  -out string
        the output file (default "validate.gh.go")
  -recv string
        ignore handlers defined on other receivers
  -v    prints additional information
```

---

## Example

You provide the source directory contain your Go handlers and filename that will be created in that directory:

```sh
gohandlers validate -dir handlers/pets -out validate.gh.go
```

This would create [handlers/pets/validate.gh.go](https://github.com/ufukty/gohandlers-petstore/blob/main/handlers/pets/validate.gh.go):

```go
func (bq CreatePetRequest) Validate() (errs map[string]error) {
  errs = map[string]error{}
  if err := bq.Name.Validate(); err != nil {
    errs["name"] = err
  }
  if err := bq.Tag.Validate(); err != nil {
    errs["tag"] = err
  }
  return
}

func (bq DeletePetRequest) Validate() (errs map[string]error) {
  //
}

func (bq GetPetRequest) Validate() (errs map[string]error) {
  //
}

func (bq ListPetsRequest) Validate() (errs map[string]error) {
  //
}
```

Notice `CreatePetRequest.Validate` collects all errors returned by field validators to return to caller.
