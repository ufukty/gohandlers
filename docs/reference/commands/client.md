# ✱ `client`

Generates a Go file (default `client.gh.go`) containing a **Client struct** and one method per handler function. These methods construct HTTP requests using your binding types and send them, returning the response (either raw or parsed into a response binding).

-   **Generated Client Structure:** The client has a simple pool-based design for obtaining host URLs:

    ```go
    type Pool interface {
      Host() (string, error)
    }

    type Client struct {
      p Pool
    }

    func NewClient(p Pool) *Client {
      return &Client{p: p}
    }
    ```

    You supply a `Pool` (which could be as simple as a struct with a `Host()` that returns a constant base URL, or something more sophisticated for load balancing).

-   **Generated Methods:** For each handler, the client has a similarly named method. It takes a pointer to the request struct and returns either a pointer to the response struct (if one is defined) or an `*http.Response` (if no response binding exists). Example:

    ```go
    func (c *Client) CreatePet(req *pets.CreatePetRequest) (*pets.CreatePetResponse, error)
    func (c *Client) DeletePet(req *pets.DeletePetRequest) (*http.Response, error)
    ```

    Under the hood, these methods will:

    1. Call `c.p.Host()` to get the base URL.
    2. Call `req.Build(host)` to get an `*http.Request`.
    3. Use `http.DefaultClient.Do(request)` to perform the HTTP call.
    4. Check for non-200 status codes and return errors accordingly.
    5. If there’s a response struct (`CreatePetResponse` in this case), instantiate it and [call](https://github.com/ufukty/gohandlers-petstore/blob/280eff72d24d32f5d61b32361653de906cd639bd/client/client.gh.go#L40) `.Parse(response)` to populate it, then return it.

---

## What it solves?

Writing and maintaining custom client code for your API (or using generic tools) can lead to mismatches as your API evolves. The Gohandlers client ensures **the client library always matches the server**. If you add a new query parameter or change an endpoint, regenerate the client — it will have the updated method signature and logic.

---

## Usage

```sh
# Gohandlers client --help
Usage of client:
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

Using the `CreatePet` example from previous examples, after running:

```sh
Gohandlers client \
  -dir handlers/pets \
  -out client/client.gh.go \
  -pkg client \
  -import "github.com/yourusername/yourrepo/handlers/pets"
```

This will generate [`client/client.gh.go`](https://github.com/ufukty/gohandlers-petstore/blob/280eff72d24d32f5d61b32361653de906cd639bd/client/client.gh.go#L23-L45) containing a `Client` with a method:

```go
func (c *Client) CreatePet(bq *pets.CreatePetRequest) (*pets.CreatePetResponse, error) {
  h, err := c.p.Host()
  //
  rq, err := bq.Build(h)
  //
  rs, err := http.DefaultClient.Do(rq)
  //
  bs := &pets.CreatePetResponse{}
  err = bs.Parse(rs)
  //
  return bs, nil
}
```

For `DeletePet` which has no response struct, the method would return `(*http.Response, error)` and skip parsing (just return the raw response on success).

You can then use the generated client in your other Go services or tests:

```go
func main() {
  pool := &StaticPool{BaseURL: "http://localhost:8080"}  // your implementation of Pool
  pets := client.NewClient(pool)

  resp, err := pets.CreatePet(&pets.CreatePetRequest{
    Name: "Doge",
    Tag:  "meme",
  })
  if err != nil {
    //
  }
  fmt.Println("New pet ID:", resp.ID)
}
```
