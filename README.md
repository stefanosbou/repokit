# repokit

CLI for bulk-managing local git repositories: registry, status dashboard, parallel pull/fetch, dead-branch reaper.

## Install

```bash
go install github.com/stefanosbou/repokit@latest
```

Or build from source:

```bash
go build -o repokit .
```


## Config

Default location: `~/.config/repokit/config.yaml`

```yaml
version: "1"
settings:
  parallel: 0              # max concurrent ops (0 = NumCPU)
  pull_strategy: ff-only   # ff-only | rebase | merge
  clean:
    stale_after_days: 30
    protected_branches:
      - main
      - master
      - develop
      - staging
repos:
  - name: myrepo
    path: /home/user/code/myrepo
```

Override config path with `--config <path>`.

## Commands

### Registry

```bash
repokit add <path> [--name <name>]   # register a repo
repokit remove <name>                # unregister a repo
repokit list                         # list all registered repos
repokit scan [path] [--depth N]      # discover and register all git repos under a directory (default depth: 5)
```

### Status

```bash
repokit status [--filter <state>]
```

Shows a fleet health dashboard — branch, status, and last commit for every registered repo.

Filter by state: `clean` `dirty` `behind` `unpushed` `stale` `detached`

### Pull

```bash
repokit pull [--strategy ff-only|rebase|merge] [--filter behind] [--force]
```

Pulls all registered repos in parallel with a live per-repo progress view. Falls back to line-by-line output when stdout is not a TTY (CI, pipes).

- `--strategy` overrides the config pull strategy
- `--filter behind` only pulls repos that are behind their upstream
- `--force` pulls even repos with uncommitted changes

### Fetch

```bash
repokit fetch
```

Fetches all registered repos in parallel with a live per-repo progress view.

### Clean

```bash
repokit clean branches [--base <branch>] [--older-than <days>] [--force]
```

Deletes merged local branches across all repos.

- `--base` sets the branch to check merged-into (default: auto-detects `main`/`master`)
- `--older-than N` only deletes branches older than N days
- `--force` uses `git branch -D` instead of `git branch -d`

## Development

```bash
make build    # build to bin/repokit
make test     # run test suite
make lint     # run golangci-lint
make install  # install to $GOPATH/bin
make tidy     # go mod tidy
```

## Global Flags

```
--config string    Override config file path
--parallel int     Max concurrent operations (0 = NumCPU)
--repo string      Target a single repo by name
```
