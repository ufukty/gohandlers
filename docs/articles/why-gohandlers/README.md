# 🙋 Why Gohandlers?

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
