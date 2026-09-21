# repom — Repo Manager

A terminal UI for managing many git repositories that live side by side in a
single workspace directory.

`repom` scans a root directory for git repositories and lets you update them or
create feature branches across many repos at once — all from one screen, with
the operations running in parallel.

## Features

- Scans a workspace directory and lists every sub-directory that contains a
  `.git` directory.
- Shows the current branch and a dirty-marker (`●`) for uncommitted changes.
- **Update** selected repos: stash uncommitted work (including untracked
  files), check out the default branch (`develop`, falling back to `main`) and
  pull, reporting commit/file stats.
- **Create branch** in selected repos in parallel: temp-commits uncommitted
  work if needed, checks out and pulls the base branch, then creates the new
  branch.
- Multi-select with checkboxes and run everything concurrently.

## Requirements

- Go 1.24+ (to build from source)
- `git` available in `PATH`

## Install

### Install script (Linux & macOS)

Downloads the prebuilt binary for your platform and installs it to
`/usr/local/bin` (or `~/.local/bin` when that is not writable):

```sh
curl -fsSL https://raw.githubusercontent.com/ZaRqax/repom/main/install.sh | sh
```

Install a specific version with `REPOM_VERSION`:

```sh
curl -fsSL https://raw.githubusercontent.com/ZaRqax/repom/main/install.sh | REPOM_VERSION=1.0.0 sh
```

### Go

Requires Go 1.24+:

```sh
go install github.com/ZaRqax/repom@latest
```

### Prebuilt binaries

Grab the archive for your OS/arch from the
[releases page](https://github.com/ZaRqax/repom/releases), extract it and put
`repom` on your `PATH`.

### Build from source

```sh
git clone https://github.com/ZaRqax/repom.git
cd repom
go build -o repom .
```

## Usage

Point `repom` at the directory that contains your repositories. It defaults to
the current directory.

```sh
./repom /path/to/workspace
```

Show the installed version:

```sh
repom --version
```

### Key bindings

| Key | Action |
|---|---|
| `↑` / `k`, `↓` / `j` | Move the cursor |
| `space` | Select / deselect the repo under the cursor |
| `a` | Select or deselect all repos |
| `u` | Update selected repos (stash → checkout default branch → pull) |
| `b` | Create a branch in selected repos |
| `r` | Refresh the repo list |
| `q` / `ctrl+c` | Quit |

## How updates work

For each selected repository:

1. If the working tree is dirty, changes are stashed (untracked files
   included).
2. The default branch is checked out — `develop` when it exists locally or on
   `origin`, otherwise `main`.
3. `git pull` runs and the number of new commits and changed files is reported.

## How branch creation works

For each selected repository:

1. If the working tree is dirty, a temporary `WIP` commit is created.
2. The default branch is checked out and pulled.
3. The new branch is created with `git checkout -b`.

## Project layout

The code follows a clean-architecture layering, `domain → repository →
service → transport`, with `main.go` acting as the composition root:

```
main.go                                   wiring and CLI entry point
internal/
├── domain/                               entities (Repo, PullStats, OpResult)
├── repository/                           interfaces (Git, RepoFinder)
│   ├── git/                              git CLI implementation
│   └── finder/                           filesystem repo discovery
├── service/repomanager/                  update / branch business logic
└── transport/tui/                        bubbletea model and views
```

The transport layer depends only on the service interface, and the service
depends only on repository interfaces — the concrete git and finder
implementations are injected in `main.go`.

## License

[MIT](LICENSE)
