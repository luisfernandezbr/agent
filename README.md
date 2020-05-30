<div align="center">
	<img width="500" src=".github/logo.svg" alt="pinpt-logo">
</div>

<p align="center" color="#6a737d">
	<strong>This repo contains a working prototype for the next gen agent</strong>
</p>


## Overview

This is a working concept prototype for the next generation of the Agent.  It's meant to experiment with some different design choices and to validate some potential architectural decisions.

## Building

You can build like this:

```
go install
```

This will build the agent.next binary and install it in your `$GOPATH/bin` folder. Make this this folder is on your `$PATH`.

Make sure and setup your `$GOPRIVATE` env to read any pinpoint modules from our repo:

```
go env -w GOPRIVATE=github.com/pinpt
```

## Running

Clone the GitHub repo integration:

```
git clone git@github.com:pinpt/agent.next.github
```

Then run:

```
go run . dev ../agent.next.github --log-level=debug --config apikey=$PP_GITHUB_TOKEN
```

This will print each exported model to the console.

You can run and have exports go to a directory such as:

```
go run . dev ../agent.next.github --log-level=debug --config apikey=$PP_GITHUB_TOKEN --dir exports
```

The `--dir` takes a folder to place the exported models (all data per model goes into one file JSON new line delimited).
