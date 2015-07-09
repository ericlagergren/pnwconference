compressedhandler
===================

compressedhandler is Go middleware that will compress an HTTP response
before it's sent back to the client. It first checks which compression
algorithms the client supports, and will compress the content in the
with the first applicable algorithm in this order: Gzip, Deflate, <future algorithms>.

If the client doesn't support any applicable algorithm, it'll simply send
uncompressed content.

## Usage

Simply wrap your current handler inside `CompressedHandler`, and your
handler will automagically return the potentially compressed content.

```go
package main

import (
	"io"
	"net/http"

	ch "github.com/EricLagerg/compressedhandler"
)

func main() {
	uncompressedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "Hello, world!")
	})

	compressedHandler := ch.CompressedHandler(uncompressedHandler)

	http.Handle("/", compressedHandler)
	http.ListenAndServe("0.0.0.0:8000", nil)
}
```
## Documentation

[godoc.org] [docs]

## License

[Apache 2.0] [license].

## Thank You

Thanks to the [New York Times] [nyt] for their Gzip handler. It served as the idea for this little project.
If *all* you need is to Gzip content, go use theirs instead.


[docs]:     https://godoc.org/github.com/EricLagerg/compressedhandler
[license]:  https://github.com/EricLagerg/compressedhandler/blob/master/license.txt
[nyt]:      https://github.com/NYTimes/gziphandler

