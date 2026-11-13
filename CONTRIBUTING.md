# Contributing

Open an issue or pull request. Run `go run ./cmd/oes validate` and `go test ./...` first.

## Reviewing a record

Reviewing is how records go from `machine-generated` to trusted. [`catalogue/REVIEW.md`](catalogue/REVIEW.md)
lists what to look at first.

1. Check the values against your knowledge and a source you can cite.
2. Correct what's wrong.
3. Remove each pointer you checked from `provenance.generated`, or check them all.
4. Mark it: `go run ./cmd/oes review --by <your name or handle> oec:<slug>`. It sets
   `provenance.review` to `community-reviewed` with `by` and today's date, and refuses while
   `provenance.generated` lists values; `--checked` says you checked every one. A qualified coach
   or clinician uses `--level expert-reviewed`.
5. Open a pull request with the records you reviewed. `oes validate` refuses a reviewed record that
   still lists generated values or lacks `by` and `at`.

## Rules for records

- Don't change a published ID. Add the old one to `previous_ids` if a rename is truly needed.
- Don't delete a record. Set `status` to `deprecated` or `replaced`.
- Only contribute text you wrote or may license under CC-BY-SA 4.0, and credit sources.
- `bodyweight_fraction` and `energy.met` need a citation.

## Changing the format

Open an issue with real records it would improve. New vocabulary terms are additions; renaming a term or field is a breaking change.

## Releasing

Every pull request a consumer would notice adds an entry under `Unreleased` in [CHANGELOG.md](CHANGELOG.md), and keeps its `Still to do` list current: add what you leave undone, remove what ships.

Maintainers cut a release when records or the format have changed enough for apps to take:

1. Move the `Unreleased` entries in [CHANGELOG.md](CHANGELOG.md) under a new `## vX.Y.Z (date)`
   heading, saying what a consumer must act on (renames, retirements, replacements, alias changes
   that change matches, format changes), and merge.
2. Tag the merge and push the tag: `git tag v0.2.0 && git push origin v0.2.0`.
3. The Release workflow validates, and publishes a GitHub release with the changelog entry as notes
   and `catalogue.json` attached.

Until v1, versions are `v0.N.P`: `N` for releases that add, rename, retire or replace records or
change the format, `P` for corrections that don't.

