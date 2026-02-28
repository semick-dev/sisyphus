# sisyphus

Have you ever fought a failing build and wanted to kill yourself? Enter `sisyphus`.

![This could be you](image.png)

For things that are seemingly just fighting with the OS or system, let this simple agent loop push the rock for you.

Start it with the following:

- A ADO PAT for interacting with the build system
- A target build URL

Your local pwd will be used as the working directory, so plan accordingly!

## Go rewrite

The agent has been rewritten in Go under:

Build the CLI:

```bash
go build -o ./bin/sisyphus ./cmd/sisyphus
```

Install onto your PATH:

```bash
go install ./cmd/sisyphus
```

## Invocation Patterns

`--build` supports two URL forms:

- Build definition URL (`?definitionId=...`): queues a new build for your current branch. You will be prompted for an optional initial prompt; if it produces git changes, they are committed/pushed before the normal loop starts.
- Build results URL (`?buildId=...`): treats that starting build as a failure context, attempts a fix, commits/pushes, then enters the normal queue/wait loop.

Examples:

```bash
sisyphus \
  --build "https://dev.azure.com/myorg/myproject/_build?definitionId=42" \
  --pat "$ADO_PAT"
```

```bash
sisyphus \
  --build "https://dev.azure.com/myorg/myproject/_build/results?buildId=447&view=results" \
  --pat "$ADO_PAT"
```

`--cli` controls the executor and defaults to `codex`.
