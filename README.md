gomed
=====

A go library for working with HL7

Depends on xmlx: https://github.com/jteeuwen/go-pkg-xmlx

A library for converting HL7 into XML, working with that XML, and then
converting it back again.

example usage:

```go
import (
  "gomed"
)

func main() {
  xmldoc := gomed.FromFile("hl7.txt")
  edited := gomed.EditNode("PID", "PID.25", xmldoc, "(111)222-3333")
  println(gomed.DocumentToHl7(edited))
}
```
