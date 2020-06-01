# TODO

This is an overview of a set of notes and TODOs for bringing this Agent v4 to life.

## Design Stuff

- Web Hooks need to be fully flushed out and implemented
- Need to figure out what's required for onboard still
- Need to figure out the startup sequence for self-managed vs multi-tenant
- Do we need agent pings anymore (maybe for self-managed only?)
- Need a separate project (private) for building the multi-tenant agent with Kustomize
- Need a create a project generator to generate an integration
- How do we simplify work for non-sprint/kanban implementations like GitHub/GitLab?
- Need to really flush out historical vs incremental from a state standpoint
- Need good way to instrument and have good observability of the agents
- Can we simplify / reduce some of the agent messages (payload data)
- Need to think through how we do versioning of plugins/agent since they have to be built together


## Ports

- Jira
- GitHub Issues
- GitLab (Issues + Source)
- BitBucket
- Google GSuite
- O365
- MSFT DevOps
