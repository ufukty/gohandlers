# 🧪 Testing Clients

When building an HTTP API, it’s crucial to test the layers _around_ your HTTP handlers: the business logic, the services, the integrations. But if your code depends on a real HTTP client to call your own API, testing can get messy and slow—requiring servers, requests, responses, and a lot of setup.

That’s where **Gohandlers** shines again. It doesn’t just generate server-side code and typed HTTP clients—it also creates **mock clients**, so you can test your services quickly and deterministically, with no need to spin up HTTP servers or stub request objects.

In this article, you’ll learn how to use Gohandlers’ generated mocks to write fast, focused, and maintainable tests for any code that uses your API clients.

---

## 🧪 Why Use a Mock Client?

Imagine you have a service that relies on an API client to call another module or service in your system:

```go
type PetService struct {
  Client client.Interface
}

func (s *PetService) RegisterPet(ctx context.Context, name, tag string) (string, error) {
  resp, err := s.Client.CreatePet(ctx, &client.CreatePetRequest{
    Name: name,
    Tag:  tag,
  })
  if err != nil {
    return "", err
  }
  return resp.ID, nil
}
```

How do you test this without calling the real API?

You use a **mock client**—one that pretends to be the real thing, but behaves how you want in your test.

---

## ⚙️ Generating Mocks

To generate mocks for your typed API client, run:

```bash
gohandlers mock \
  --dir handlers/pets \
  --pkg client \
  --out mock.gh.go \
  --v
```

This will generate:

-   A `MockClient` type
-   An `Interface` that both the real and mock clients implement
-   Methods like `CreatePetFunc` that you can stub in your tests

---

## 🧰 Example: Unit Testing with a Mock

Here’s how you’d test the `PetService` example from earlier:

```go
func TestRegisterPet(t *testing.T) {
  mock := &client.MockClient{}
  mock.CreatePetFunc = func(ctx context.Context, req *client.CreatePetRequest) (*client.CreatePetResponse, error) {
    if req.Name == "" {
      return nil, errors.New("name is required")
    }
    return &client.CreatePetResponse{ID: "mock123"}, nil
  }

  service := &PetService{Client: mock}

  id, err := service.RegisterPet(context.Background(), "Whiskers", "cat")
  if err != nil {
    t.Fatalf("unexpected error: %v", err)
  }
  if id != "mock123" {
    t.Errorf("unexpected ID: got %s, want mock123", id)
  }
}
```

✅ No network.  
✅ No JSON marshalling.  
✅ No HTTP server.  
✅ 100% pure logic test.

---

## 📐 Mock Signature

Each method on the mock client has this pattern:

```go
type MockClient struct {
  CreatePetFunc func(ctx context.Context, req *CreatePetRequest) (*CreatePetResponse, error)
}
```

If you don’t assign the function, calling the method will panic with a helpful message like:

```
panic: mock CreatePetFunc not implemented
```

This ensures you never silently skip behavior in tests.

---

## 🎛 Mocking Edge Cases

You can fully control the behavior of the mock to simulate:

-   Network errors
-   Unexpected response payloads
-   Validation failures
-   Delays or timeouts

Example:

```go
mock.CreatePetFunc = func(ctx context.Context, req *CreatePetRequest) (*CreatePetResponse, error) {
  return nil, fmt.Errorf("server temporarily unavailable")
}
```

This lets you test retries, fallbacks, or error handling logic—without hitting a real endpoint.

---

## 🔁 Table-Driven Testing with Mocks

Want to test multiple scenarios cleanly? Use table-driven tests:

```go
tests := []struct {
  name string
  setup func(*client.MockClient)
  expectErr bool
}{
  {
    name: "success",
    setup: func(mock *client.MockClient) {
      mock.CreatePetFunc = func(ctx context.Context, req *CreatePetRequest) (*CreatePetResponse, error) {
        return &CreatePetResponse{ID: "abc"}, nil
      }
    },
    expectErr: false,
  },
  {
    name: "empty name",
    setup: func(mock *client.MockClient) {
      mock.CreatePetFunc = func(ctx context.Context, req *CreatePetRequest) (*CreatePetResponse, error) {
        return nil, errors.New("name is required")
      }
    },
    expectErr: true,
  },
}

for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
    mock := &client.MockClient{}
    tt.setup(mock)

    service := &PetService{Client: mock}
    _, err := service.RegisterPet(context.Background(), "", "dog")

    if tt.expectErr && err == nil {
      t.Errorf("expected error but got nil")
    }
  })
}
```

This pattern keeps your test logic tidy and expressive.

---

## 🧠 Why Mocks Are Better Than Test Servers

| With MockClient               | With Test HTTP Server                 |
| ----------------------------- | ------------------------------------- |
| No actual HTTP traffic        | Requires `httptest.Server`            |
| Fast and deterministic        | Can be slower and flaky               |
| Full control over behavior    | Must parse requests, encode responses |
| Statically typed interactions | Often uses `interface{}` or raw JSON  |
| Zero setup or teardown        | Needs cleanup and coordination        |

Mocks are ideal when testing **how your code behaves**, not the behavior of the HTTP infrastructure.

---

## 🧼 Best Practices

-   Use mock clients to test **business logic** that calls APIs
-   Prefer static stubs over spinning up real HTTP servers
-   Fail tests if mock methods aren’t configured (`nil` func = panic)
-   Test edge cases like retries, timeouts, and malformed responses

---

## ✅ Summary

Gohandlers-generated mock clients help you write tests that are:

-   Fast ⚡
-   Predictable 📋
-   Focused 🔍
-   Typed-safe 🧬

You don’t need to write mocks by hand, and you don’t need to simulate full HTTP interactions. With Gohandlers, testing your API-consuming code is just as clean as writing it.

**Code with confidence. Test with ease.** 🧪✅🚀
