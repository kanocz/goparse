# goParse
Several functions for simplify go source parsing

## Example
```go
package main

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/kanocz/goparse"
)

type reqProfile struct {
	ID     int64  `rq:"id,omitempty",json:"id"`
	Name   string `rq:"name",json:"-"`
	Passwd string
}

type otherProfile struct {
	ID   int64  `req:"id"`
	Name string `req:"name"`
}

func main() {
	s, _ := goparse.GetFileStructs("example.go", "req", "rq")
	d, _ := json.Marshal(s)
	var out bytes.Buffer
	json.Indent(&out, d, "", "  ")
	out.WriteTo(os.Stdout)
}
```

will produce

```json
[
  {
    "Name": "reqProfile",
    "Field": [
      {
        "Name": "ID",
        "Type": "int64",
        "Tags": [
          "id",
          "omitempty"
        ]
      },
      {
        "Name": "Name",
        "Type": "string",
        "Tags": [
          "name"
        ]
      },
      {
        "Name": "Passwd",
        "Type": "string",
        "Tags": null
      }
    ]
  }
]
```

