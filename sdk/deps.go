package sdk

import (
	// create a required dependency to pin it otherwise it's transitive in plugins
	_ "golang.org/x/net/context"
	// create a required dependency to pin it otherwise it's transitive in plugins
	_ "golang.org/x/sys/unix"
	// create a required dependency to pin it otherwise it's transitive in plugins
	_ "github.com/mattn/go-colorable"
	// create a required dependency to pin it otherwise it's transitive in plugins
	_ "github.com/mattn/go-isatty"
	// create a required dependency to pin it otherwise it's transitive in plugins
	_ "github.com/spf13/pflag"
)
