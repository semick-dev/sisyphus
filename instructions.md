`sisypnus` is intended to be an _agent loop_ whose goal is to utilize minimal llm resources to queue a job, interpret the result, make a change, queue a job...

`codex`, I want you to help me build a plan utilizing the structure that I give you.

I am writing a project that will be an `agent loop`. by that, it will run in a users terminal and utilize their logged in `codex`, `claude`, or `copilot` to accept a github issue, a build definition, and a local path to start a branch in our name.

- it'll build up an `instructions.md` based in the input issue and build definition and local path
- then `codex -p <path to instructions.md>` will be invoked in `autopilot` mode, permissions should be asked for in the very beginning before invoking copilot to ensure that dotnet/python3/bash commands can be run within the target directory only.
- if codex exited with 0, add the commit, run the input build definition, store the build id, sleep for <some input duration>
- wake, check for pending/success/failure from previously stored build id
- if failure, grab the failed job output, add to a new instructions.md, call selected llm cli -p <our new instructions.md>
  - we need to add some limiting where we only give partial bits of the failed log, as these can get up till 200mb
- on success, exit


