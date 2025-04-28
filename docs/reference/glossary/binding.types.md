# Binding types

Binding types are what you are usually "Unmarshal" your http request or response "into". They store most of (if not all) the information you need from the request/response to access them later inside the handler. As the fields store http parameters are typed, you skip many repetitive steps in your handler and directly start using values.

Defining binding types for your handlers is a great way to organize your files and make them easier to read later. They "list" all the information comes to and goes from an handler at a glance.

The beauty of it, once you have all the parameters of an endpoint in one place, creating helpers like parsing, validating and serializing of parameters become automation friendly.

Here is a handler file which defines a **"C"**RUD endpoint on `Pet` resource that also have request and response bindings.

```go
package pets

type CreatePetRequest struct {
  Name string `json:"name"`
  Tag  string `json:"tag"`
}

type CreatePetResponse struct {
  ID string `json:"id"`
}

func (p *Pets) CreatePet(w http.ResponseWriter, r *http.Request)
```

Notice even though the implementation of `CreatePet` is not here, you can make your guesses on what it internally does based on its name and request, response data.

## Field tags

Gohandlers supports 4 types of field tags:

| Tag    | Description | For |
| ------ | ----------- | --- |
| `form` |
| `json` |
| `json` |
