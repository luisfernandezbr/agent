//go:generate go run generator/gittag/main.go
//go:generate go-bindata -pkg generator -prefix generator/ -o generator/gen.go generator/template/...
package main

import (
	"github.com/pinpt/agent/v4/cmd"
)

// these values go from the go build, do not change them
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, commit, date)
}
