// +build exclude

package main

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// this is a simple file used by go:generate to create a file with the current git tag
func main() {
	cwd, _ := os.Getwd()
	tag, err := exec.Command("git", "tag").Output()
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	tags := strings.Split(string(tag), "\n")
	if len(tags) == 0 {
		fmt.Println("ERROR", "no tags found, this is so weird")
		os.Exit(1)
	}
	currenttag := tags[len(tags)-2] // last item is an empty space
	fn := filepath.Join(cwd, "generator", "gittag.go")
	os.Remove(fn)

	content := fmt.Sprintf(`
		package generator
		const gitTag = "%s"`, currenttag,
	)

	formatted, err := format.Source([]byte(content))
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(fn, formatted, 0644); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}
