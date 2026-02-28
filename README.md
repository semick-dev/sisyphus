# sisyphus

Have you ever fought a failing build and wanted to kill yourself? Enter `sisyphus`.

![This could be you](image.png)

For things that are seemingly just fighting with the OS or system, let this simple agent loop push the rock for you.

Start it with the following:

- A ADO PAT for interacting with the build system
- A target build you want to run
- An issue# or starting prompt

Your local pwd will be used as the working directory, so plan accordingly!

## Go rewrite

The agent has been rewritten in Go under:

- `cmd/sisyphus-agent`
- `internal/ado`
- `internal/payload`
- `internal/push`
- `internal/man`

Build the CLI:

```bash
go build -o ./bin/sisyphus-agent ./cmd/sisyphus-agent
```

Install onto your PATH:

```bash
go install ./cmd/sisyphus-agent
```
