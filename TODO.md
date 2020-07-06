# Agent 4.0

# Current Status

| Integration         | Export | WebHook | Mutation  | WebApp  | Notes                   |
|---------------------|:------:|:-------:|:---------:|:-------:| ------------------------|
| Jira                |   ✅   |   🛑    |   ✅      |   🛑    |                          |
| BitBucket           |   ✅   |   ✅    |   🛑      |   🛑    |                          |
| GitHub              |   ✅   |   ✅    |   ✅      |   ✅    |                          |
| GitLab              |   ✅   |   🛑    |   🛑      |   🛑    |                          |
| Azure               |   ✅   |   🛑    |   🛑      |   🛑    |                          |
| GSuite              |   ✅   |   🛑    |   🛑      |   🛑    |                          |
| Office365           |   ✅   |   🛑    |   🛑      |   🛑    |                          |


# TODO

- Store error, error_message on project, repo
- Store issue types, resolutions on project
- Store transition on issue
- Model: Ability to send active for repo, comment, etc when deleted
- WebHook: need a model to track webhooks (add to repo, project, instance)
- Add integration_instance_id to data-manager models (ad update other issues from above)
- Finish UI SDK changes to webapp
- Finish UI webapp for all integrations (Need design help from Congrove)
- Validate Issue from GitHub on pipeline
- Github user mapping validation
- UI: Add on-premise UI
- UI: Ability to re-auth
