# 🤙 Generating Clients

Building an HTTP API is only half the battle. The other half? **Consuming it safely and efficiently.**

Manually crafting client code to call your Go API can lead to duplicated logic, fragile interfaces, and plenty of boilerplate. That’s why **Gohandlers** doesn’t just generate server-side glue—it also gives you **typed, ready-to-use clients** for your endpoints.

In this article, you'll learn how Gohandlers turns your existing request/response types into a complete HTTP client, complete with automatic serialization, request building, and error parsing—no manual code required.

---

## 🧩 Why Typed Clients?

When you consume your own API (or someone else's), you often end up writing code like:

```go
data, _ := json.Marshal(payload)
req, _ := http.NewRequest("POST", baseURL+"/pets", bytes.NewBuffer(data))
req.Header.Set("Content-Type", "application/json")
resp, _ := http.DefaultClient.Do(req)
```

This is:

-   **Verbose** and easy to get wrong
-   **Unstructured**—no compile-time guarantees
-   **Disconnected** from your server’s logic and types

Gohandlers solves this by generating a client that knows how to:

-   Build requests using your `...Request` types
-   Send them with standard or custom HTTP clients
-   Parse responses into your `...Response` types
-   Return typed results with minimal code

---

## ⚙️ Generating the Client Code

After defining your handlers and binding structs, run:

```bash
Gohandlers client \
  --dir handlers/pets \
  --pkg client \
  --out client.gh.go \
  --v
```

This tells Gohandlers to:

-   Inspect all `...Request` and `...Response` types
-   Look up the associated handler metadata (method, path, etc.)
-   Generate a typed `Client` struct with methods for each endpoint

You’ll get a file like:

```go
package client

type Client struct {
  Pool    Pool
  Doer    Doer
  Options Options
}

func (c *Client) CreatePet(ctx context.Context, in *CreatePetRequest) (*CreatePetResponse, error) {
  // generated: build request, send, parse response
}
```

---

## 🏗️ How the Client Works

Each method follows this pattern:

1. **Call `Build()`** on the request struct to create an `*http.Request`
2. **Send** the request using an injected HTTP client
3. **Call `Parse()`** on the response struct to extract the result
4. **Return** the typed response or any errors

This means all serialization logic is already defined by your struct tags—you never touch `json.Marshal` or `http.NewRequest` again.

---

## ✨ Example Usage

Assume you have:

```go
type CreatePetRequest struct {
  Name string `json:"name"`
  Tag  string `json:"tag"`
}

type CreatePetResponse struct {
  ID string `json:"id"`
}
```

Then your generated client lets you call the endpoint like this:

```go
client := client.NewClient(client.StaticPool("http://localhost:8080"))
ctx := context.Background()

resp, err := client.CreatePet(ctx, &CreatePetRequest{
  Name: "Whiskers",
  Tag:  "cat",
})
```

No need to:

-   Manually write HTTP requests
-   Marshal structs to JSON
-   Handle content-type headers
-   Decode responses manually

The generated method does all of that for you.

---

## 🌐 Pooling & Hosts

The generated client uses a **`Pool` interface** to select the base URL for a request. This gives you flexibility:

-   Use `StaticPool("http://localhost:8080")` for single-host clients
-   Implement your own pool (e.g. for sharding, round-robin, or fallback hosts)

You can also override this per-request using client `Options`.

---

## 🔁 Returning Raw Responses

In some cases, you might want more control over the response—headers, status code, etc. The generated client also gives you `...Raw` methods that return `*http.Response` instead of parsed structs:

```go
resp, err := client.CreatePetRaw(ctx, &CreatePetRequest{...})
```

You can then manually inspect headers, body, or status if needed.

---

## 🧪 Testing with Mocks

Gohandlers can also generate a mock implementation of the same client interface:

```bash
Gohandlers mock \
  --dir handlers/pets \
  --pkg client \
  --out mock.gh.go \
  --v
```

Now you can write clean, predictable tests:

```go
mock := &client.MockClient{}
mock.CreatePetFunc = func(ctx context.Context, req *CreatePetRequest) (*CreatePetResponse, error) {
  return &CreatePetResponse{ID: "test123"}, nil
}
```

Inject `mock` wherever the real client would go—no networking required.

---

## 🔐 Bonus: Customizing Transport

The `Client` struct exposes a `Doer` interface, allowing you to:

-   Use `http.DefaultClient`
-   Inject a custom client with timeouts
-   Wrap it with middlewares (e.g. logging, retries)

```go
client := client.NewClient(...)
client.Doer = &http.Client{
  Timeout: 5 * time.Second,
}
```

Or use it with an HTTP tracing package or distributed tracing header injector.

---

## ✅ Summary: Why Typed Clients Rock

| Feature                | Benefit                                               |
| ---------------------- | ----------------------------------------------------- |
| Auto-generated methods | Eliminate manual request/response plumbing            |
| Strongly-typed         | Compile-time guarantees—no `map[string]interface{}`   |
| Consistent with server | Uses the same structs and tags as your handlers       |
| Easy to test           | Built-in mocks for each method                        |
| Configurable transport | Support for pooling, retries, and custom HTTP clients |

---

## 🎯 Conclusion

Typed HTTP clients from Gohandlers bridge the gap between server and consumer. They reduce friction, cut boilerplate, and give you a clean, safe way to consume your own APIs (or provide SDKs to others).

With one command, you gain:

-   An interface to your entire API
-   Full control over transport
-   Easy integration with tests

You’ve already defined the shape of your API—why write a client by hand? Let Gohandlers do it for you.

**Code once. Use everywhere.** 🚀
