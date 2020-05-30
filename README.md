<div align="center">
	<img width="500" src=".github/logo.svg" alt="pinpt-logo">
</div>

<p align="center" color="#6a737d">
	<strong>This repo contains a working prototype for the next gen agent</strong>
</p>


## Overview

This is a working concept prototype for the next generation of the Agent.  It's meant to experiment with some different design choices and to validate some potential architectural decisions.

## Running

You can run like this:

```
go run . dev github --log-level=debug --config apikey=$PP_GITHUB_TOKEN --config organization=pinpt
```

This will run an export for GitHub and print all the JSON objects to the console.
