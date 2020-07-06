# Agent 4.0

# Current Status

| Integration         | Export | WebHook | Mutation  | WebApp  | Notes                     |
|---------------------|:------:|:-------:|:---------:|:--------:| -------------------------|
| Jira                |   ✅   |   🛑    |   ✅      |   🛑    |                          |
| BitBucket           |   ✅   |   ✅    |   🛑      |   🛑    |                          |
| GitHub              |   ✅   |   ✅    |   ✅      |   ✅    |                          |
| GitLab              |   ✅   |   🛑    |   🛑      |   🛑    |                          |
| Azure               |   ✅   |   🛑    |   🛑      |   🛑    |                          |
| GSuite              |   ✅   |   🛑    |   🛑      |   🛑    |                          |
| Office365           |   ✅   |   🛑    |   🛑      |   🛑    |                          |


# TODO

- Finish UI SDK changes to webapp
- Finish UI webapp for all integrations (Need design help from Congrove)
- Validate Issue from GitHub on pipeline
- Github user mapping validation
- UI: Add on-premise UI
- UI: Ability to re-auth
- WebHook: need a model to track webhooks
- Robin: do we still use PullRequestBranch?
- Model: Ability to send active for repo, comment, etc when deleted
- Export: need ability to deliver repo issues and show in the UI like we do today (entity_errors)
