# Changelog

Releases of the Open Exercise Schema and the Open Exercise Catalogue, newest first. Pin a release
rather than `main` ([GUIDE.md](GUIDE.md)). Until v1 the format may still change, and a release says
so when it does. Records are never deleted and IDs never reused; a rename adds the old ID to
`previous_ids`.

Each entry lists what a consumer has to act on: records added, renamed, retired or replaced, alias
changes that change what matches, and format changes.

## Unreleased

- **Guide:** the worked matching example now says how Lifterclub treats `legacy` aliases (a
  suggestion, never a sure match), and no longer groups `misspelling` with them: the table trusts
  misspellings for matching what people type.

- **Replaced (41):** near-duplicates from `catalogue/REVIEW.md` section 4 now point at one record
  each through `replaced_by` (for example `oec:dumbbell-curl` → `oec:dumbbell-biceps-curl`,
  `oec:jogging` → `oec:running`, both single-arm cable lateral raises → `oec:cable-lateral-raise`),
  plus `oec:bag-training` → `oec:boxing` and `oec:machine-dip-or-triceps-extension` →
  `oec:machine-seated-dip`. Move logs on a replaced record to its `replaced_by`.
- **Reinstated (3):** `oec:boxing`, `oec:radio-taiso` and `oec:leg-and-hip-stretch` (now "Lower-Body
  Stretching") are `active` again, timed activities. An app that retired them on loading v0.1.0
  should restore them.
- **Added (2):** `oec:serratus-wall-slide`, split from `oec:wall-slide` (it takes wger 2523 and the
  previous ID `oec:serratus-wall-slide-stretch`), and `oec:machine-seated-dip`.
- **Moved between records:** wger 805 and `oec:tricep-pushdown-on-cable` to
  `oec:cable-rope-triceps-pushdown`; wger 1583 and `oec:supine-press` to
  `oec:barbell-close-grip-bench-press`; wger 512 and `oec:rowing-seated-narrow-grip` to
  `oec:cable-close-grip-seated-row`.
- **Aliases that change matches:** 19 wrong or vague names become `legacy` (for example "Single Arm
  Row" on the single-arm pulldown, "Overhand Cable Curl", "Dynamic Planche", "Unilateral Hip
  Thrust"); "Neck Extension" and "90/90 Hamstring Stretch" are gone; 7 misspellings become
  `common` and 2 `brand`.
- **Setup and equipment:** facets the name doesn't fix are gone from the back and Smith squats,
  pull-up, lat pulldown, deadlift, RDL, bench press, assisted pull-up and hack squat; the rack is
  optional for the overhead press and barbell lunges; 14 steps no longer prescribe a hold time;
  `oec:barbell-lying-triceps-extension` is barbell-only.
- **Setup widths:** `stance.width` and `grip.width` are gone from 439 records whose name doesn't fix
  them (SPEC §6). An app that filtered or displayed by them should read them only where present.
- **Review:** the 13 set-aside or replaced-from-set-aside records say `unreviewed`. `review.by` is
  the latest reviewer ([GUIDE.md](GUIDE.md)).

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
