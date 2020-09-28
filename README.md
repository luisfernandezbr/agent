<div align="center">
	<img width="500" src=".github/logo.svg" alt="pinpt-logo">
</div>

<p align="center" color="#6a737d">
	<strong>Agent is the software that collects and deliver performance details to the Pinpoint Cloud</strong>
</p>


## Building

You can build like this:

```
go install -tags dev
```
**NOTE:** Include the `dev` build tag do have access to the developer tools.

This will build the agent binary and install it in your `$GOPATH/bin` folder. Make sure this folder is in your `$PATH`.

Make sure and setup your `$GOPRIVATE` env to read any pinpoint modules from our repo:

```
go env -w GOPRIVATE=github.com/pinpt
```

## Creating an integration

You will need to call the `generate` command and answer a few questions, for example:

```
> agent generate
    ____  _                   _       __ 
   / __ \(_)___  ____  ____  (_)___  / /_
  / /_/ / / __ \/ __ \/ __ \/ / __ \/ __/
 / ____/ / / / / /_/ / /_/ / / / / / /_  
/_/   /_/_/ /_/ .___/\____/_/_/ /_/\__/  
             /_/                         
Welcome to the Pinpoint Integration generator!

? Go Package Name: pinpt/github
? Name of the integration: Github
? Your company's name: Pinpoint Software Inc.
? Your company's url: http://pinpoint.com
? Your company's short, unique identifier: pinpt
? Your company's avatar url: https://avatars0.githubusercontent.com/u/24400526?s=200&v=4
? Choose integration capabilities: Issue Tracking, Source Code

🎉 project created! open ~/Documents/go/src/pinpt/github in your editor and start coding!

```

The project will be created. Go to that directory and open it in VSCode to start working on it. Happy coding.

## Running for local dev

Clone the repo for the GitHub integration:

```
git clone git@github.com:pinpt/agent
```

### Export

This will print each exported model to the console.

```
cd pinpt/agent
agent dev . --log-level debug --console-out --set apikey_auth='{"apikey": "$GITHUB_TOKEN" }'
```

You can run and have exports go to a directory if `console-out` is omitted. The default directory is `./dev-dir` but you can pass in a `--dir` arg instead:

```
cd pinpt/agent
agent dev . --log-level debug --set apikey_auth='{"apikey": "$GITHUB_TOKEN" }' --dir exports
```

The `--dir` takes a folder to place the exported models (all data per model goes into one file JSON new line delimited).

If you are going to subscribe to webhooks, pass in the `--webhook` arg to register the url with out server, otherwise the register call will fail

```
cd pinpt/agent
agent dev . --log-level debug --set apikey_auth='{"apikey": "$GITHUB_TOKEN" }' --dir exports --webhook
```

### Webhooks

```
cd pinpt/agent
agent dev webhook . --log-level debug --set apikey_auth='{"apikey": "$GITHUB_TOKEN" }' --input webhook_payload.json
```

The `webhook_payload.json` contains a sample webhook payload 


## Running for server

The server mode can either run in standalone or multi agent mode.  Standalone mode is currently how agent's work today.  Each customer has one instance of an agent.  Multi agent mode is where the agent can act in a multi-tenant fashion and can process requests for multiple customers and these agents can be horizontally scaled.

### Requirements

You must first build integrations:

```
cd pinpt/agent
agent build .
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
### Record & Playback

You can enable record or playback of all HTTP interactions by passing in the `--record` or `--replay` arguments to dev. For record, this should be a directory to place the recording file.  For replay, this will be the directory where the recording file was saved.


## Contributions

We ♥️ open source and would love to see your contributions (documentation, questions, pull requests, isssue, etc). Please open an Issue or PullRequest!  If you have any questions or issues, please do not hesitate to let us know.

## License

This code is open source and licensed under the terms of the MIT License. Copyright &copy; 2020 by Pinpoint Software, Inc.