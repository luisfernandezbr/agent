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

## Creating an integration

You will need to cd to your $GOPATH/src/{org_name} and call the generate command with a project name (make sure you run the previous step first).

For example:

```
cd $GOPATH/src/github.com/pinpt
agent.next generate project_name
```

This command will fail if called from anywhere else. Once you run it, it will ask for some information, such as integration name, publisher info, etc.
The project will be created. Go to that directory and open it in VSCode to start working on it. Happy coding.

## Running for local dev

Clone the repo for the GitHub integration:

```
git clone git@github.com:pinpt/agent.next.github
```

Then run:

```
go run . dev ../agent.next.github --log-level=debug --config api_key=$PP_GITHUB_TOKEN
```

This will print each exported model to the console.

You can run and have exports go to a directory such as:

```
go run . dev ../agent.next.github --log-level=debug --config api_key=$PP_GITHUB_TOKEN --dir exports
```

The `--dir` takes a folder to place the exported models (all data per model goes into one file JSON new line delimited).

## Running for server

The server mode can either run in standalone or multi agent mode.  Standalone mode is currently how agent's work today.  Each customer has one instance of an agent.  Multi agent mode is where the agent can act in a multi-tenant fashion and can process requests for multiple customers and these agents can be horizontally scaled.

### Requirements

You must first build integrations:

```
go run . build ../agent.next.github
```

This will be placed in your `dist` folder as a file named `github`.

### Standalone

Once you build the integration, you can just run it:

```
github --log-level debug --config agent.json
```

Currently, the agent.config format matches the current (legacy) agent config.

### Multi

```
github --log-level debug
```

By default in this mode, will only talk with the local dev event-api. You can set `--channel` and `--secret` to point at another environment.

To place an event, use the event-api `produce` command such as:

```
go run . produce --log-level debug --channel dev agent.ExportRequest --input '{"customer_id":"1234","integrations":[{"name":"github","authorization":{"api_key":"XYZ"}}]}' --secret 'fa0s8f09a8sd09f8iasdlkfjalsfm,.m,xf' --header integration=github
```

Make sure you update the `api_key` with the value of your `PP_GITHUB_TOKEN`.  Also, make sure you're running event-api server locally such as:

```
PP_CUSTOMER_ID=1234 PP_INTERNAL=1 make local
```
