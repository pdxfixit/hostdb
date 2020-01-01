# HostDB

This repository contains golang code shared among hostdb-server and related projects.

The package requires a few environment variables to be set, in order to be able to send data to HostDB. Those variables are:

* `HOSTDB_URL` (optional &ndash; defaults to `https://hostdb.pdxfixit.com/v0`)
* `HOSTDB_USER` (optional &ndash; defaults to `writer`)
* `HOSTDB_PASS`

They can be omitted, which will prevent transmission.

## Import for use

```go
package main

import (
	"fmt"
	
	"github.com/pdxfixit/hostdb"
)

func main() {
	fmt.Println(fmt.Sprintf("%v", hostdb.Record{}))
}
```

## Tests

Run `make test`.

## See Also

* https://github.com/pdxfixit/hostdb-server
* https://github.com/pdxfixit/hostdb-collector-aws
* https://github.com/pdxfixit/hostdb-collector-oneview
* https://github.com/pdxfixit/hostdb-collector-vrops
