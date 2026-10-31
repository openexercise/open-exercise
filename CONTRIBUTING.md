# Contributing

Open an issue or pull request. Run `go run ./cmd/oes validate` and `go test ./...` first.

## Reviewing a record

1. Check the values against your knowledge and a source you can cite.
2. Correct what's wrong.
3. Remove each pointer you checked from `provenance.generated`.
4. When the list is empty, set `provenance.review` to `community-reviewed` with `at` and `by`.

## Rules for records

- Don't change a published ID. Add the old one to `previous_ids` if a rename is truly needed.
- Don't delete a record. Set `status` to `deprecated` or `replaced`.
- Only contribute text you wrote or may license under CC-BY-SA 4.0, and credit sources.
- `bodyweight_fraction` and `energy.met` need a citation.

## Changing the format

Open an issue with real records it would improve. New vocabulary terms are additions; renaming a term or field is a breaking change.
