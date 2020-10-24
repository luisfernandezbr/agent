// +build exclude

package main

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	parts := strings.Split(currenttag, ".")
	// need to add 1 to the tag since there will be a release after generating these files
	last, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	parts[len(parts)-1] = fmt.Sprint(last + 1)
	fn := filepath.Join(cwd, "generator", "gittag.go")
	os.Remove(fn)

	content := fmt.Sprintf(`
		package generator
		const gitTag = "%s"`, strings.Join(parts, "."),
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
