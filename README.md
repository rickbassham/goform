# goform 

Package goform is meant to make binding http data to structs easy.

    import "github.com/rickbassham/goform"

## Usage

#### func  Unmarshal

```go
func Unmarshal(r *http.Request, v interface{}) error
```
Unmarshal will bind the body and query string values to the given struct. Works
will all primitive types, time.Time, image.Image, and []byte. It first inspects
the Content-Type header of the request. If the Content-Type is json it will use
the json.Unmarshal func and then bind anything from the query string as well.
