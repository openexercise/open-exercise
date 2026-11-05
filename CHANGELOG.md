# Changelog

Releases of the Open Exercise Schema and the Open Exercise Catalogue, newest first. Pin a release
rather than `main` ([GUIDE.md](GUIDE.md)). Until v1 the format may still change, and a release says
so when it does. Records are never deleted and IDs never reused; a rename adds the old ID to
`previous_ids`.

Each entry lists what a consumer has to act on: records added, renamed, retired or replaced, alias
changes that change what matches, and format changes.

## Unreleased

## v0.1.0 (2026-11-05)

The first release: the format and the catalogue as they stand, so apps can pin them.

- **Format:** Open Exercise Schema v0 (`schema/v0/exercise.json`, `spec/v0/SPEC.md`) and its
  vocabularies. `oes validate` now also refuses a `replaced_by` chain that loops, and a record
  marked `community-reviewed` or `expert-reviewed` without `by` and `at` or with values still in
  `provenance.generated`.
- **Catalogue:** 815 records generated from wger community data (`catalogue/REPORT.md`): 796
  active, 16 deprecated (set aside: routines, either/or entries, a muscle, a junk entry) and 3
  replaced (`oec:stationary-cycling` and `oec:stationary-bike-cycling` by `oec:stationary-bike`;
  `oec:treadmill-running` by `oec:treadmill-run`). 730 previous IDs keep the merged and renamed
  wger entries resolving. 2,982 aliases: 2,721 common, 119 abbreviations, 53 brands, 45
  misspellings and 44 regional.
- **Since the first commit (2026-10-31):** Hiking, Barbell Upright Row, Barbell Good Morning and an
  air bike added; aliases other apps use; "Calf Raise" as an alias; one
  push press alias fixed; the bike and treadmill duplicates replaced.
- **Review:** no record has been reviewed by a person yet. 795 are `machine-generated`; treat their
  instructions and summaries accordingly. `catalogue/REVIEW.md` lists what to look at first, and
  `oes review` marks a record reviewed.
