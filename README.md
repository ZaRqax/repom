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

Build from source:

```sh
git clone git@github.com:ZaRqax/repom.git
cd repom
go build -o repom .
```

Or run without building:

```sh
go run . /path/to/workspace
```

## Usage

Point `repom` at the directory that contains your repositories. It defaults to
the current directory.

```sh
./repom /path/to/workspace
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

## License

[MIT](LICENSE)
