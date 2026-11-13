# CLAUDE.md

Open Exercise: an open standard for describing exercises and the public catalogue in that format
(v0 draft; formats and IDs may change until v1). Lifterclub snapshots tagged releases of the
catalogue (`make catalogue-snapshot` there).

version_file: none (releases are git tags `vX.Y.Z`; `release.yml` publishes the tag's changelog
section as the GitHub release notes)

## Rules

- Records are never deleted and IDs never reused; a rename adds the old ID to `previous_ids`.
- Read `CONTRIBUTING.md`, `GOVERNANCE.md` and `GUIDE.md` before changing the format or records.

## Changelog

`CHANGELOG.md` is the running record of this repository. Every change someone would notice adds an
entry under `## Unreleased` in the same pull request; work left undone goes under `## Still to do`
and comes off when it ships. On release, rename `Unreleased` to the release heading and start a new
empty one. `CONTRIBUTING.md` has the contributor version of this rule.

## Commits

Conventional commits (`type(scope): description`), on a branch with a pull request; never commit
to `main` directly. No attribution trailers.
