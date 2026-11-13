# Changelog

Releases of the Open Exercise Schema and the Open Exercise Catalogue, newest first. Pin a release
rather than `main` ([GUIDE.md](GUIDE.md)). Until v1 the format may still change, and a release says
so when it does. Records are never deleted and IDs never reused; a rename adds the old ID to
`previous_ids`.

Each entry lists what a consumer has to act on: records added, renamed, retired or replaced, alias
changes that change what matches, and format changes.

## Still to do

Known gaps and deferred work, removed when it ships (and recorded under Unreleased).

- **Person review of 307 records** the verification pipeline handed on: 68 with questions from the
  fix pass and 239 with issues from the check (`catalogue/verification/batch-2.json`).
- **No record is `community-reviewed` or `expert-reviewed` yet:** 382 are still
  `machine-generated` and 17 `unreviewed`. Review in the order Lifterclub users log them
  (`catalogue/REVIEW.md` §5).
- **No qualified expert reviewers** are listed in [GOVERNANCE.md](GOVERNANCE.md), so nothing can be
  `expert-reviewed` yet (`catalogue/REVIEW.md` §5).
- **`review.scope`** (what an expert checked) waits until an expert reviews (`catalogue/REVIEW.md`,
  decisions on section 5).
- **v1:** freeze the format and set up the RFC process for breaking changes after it
  ([GOVERNANCE.md](GOVERNANCE.md), Versions).

## Unreleased

- **Guides (additions):** a [developer guide](docs/developer-guide.md) for apps building on the
  schema: getting a release, every field with examples from the catalogue, units and what a load
  number means, vocabularies, validating your own records and `x-` extensions with `oes validate`,
  loading a new release (with a `jq` release diff), licences and versioning. CONTRIBUTING.md now
  covers setup, the repository layout, a worked correction, adding, renaming and retiring records,
  sources, format changes, the checks CI runs, and commit and changelog conventions. No record or
  format changes.

## v0.5.0 (2026-11-06)

Verification across the catalogue. 418 records are now `machine-verified`, 382 still
`machine-generated` and 17 `unreviewed`. SPEC §6 gains a wording rule for conventional setup facets.
Records whose aliases became `legacy` stop deciding matches by those names; an app that shows
descriptions by review status will show more as checked automatically.

- **Format (wording):** SPEC §6 now says `position`, `range`, `load_position` and the stance's `kind`
  may describe the standard way an exercise is done, as conventional grips do, when its
  instructions agree, and are left out only where they'd contradict it. No field or term changes.
- **Verification batch 2:** every remaining active machine-generated record went through the
  pipeline (`catalogue/verification/batch-2.json`). 217 records were fixed (268 fixes: aliases made
  `legacy`, summaries, equipment, muscles, load conventions, setup facets and instructions), and
  397 are now `machine-verified`, for 418 in all. The other 307 wait for a person: 68 with
  questions from the fix pass, 239 with issues from the check.
- **Checker:** `oes-verify/2` reads the schema, so it no longer calls valid values invalid, and
  applies the §6 rule above. Records whose only issues were setup facets were rechecked with it.

## v0.4.0 (2026-11-05)

Automated verification. A format addition, the `machine-verified` review status, and its first
use: 21 of the 54 exercises Lifterclub's programs use are verified, and 33 of those 54 were
corrected. An app may now show a verified record's summary and instructions, labelled as checked
automatically. Records whose aliases became `legacy` stop deciding matches by those names.

- **Verification pilot (batch 1):** the 54 exercises Lifterclub's programs use went through the
  pipeline. 33 records were fixed: summaries, aliases (misleading ones now `legacy`), equipment,
  load conventions, setup facets and instructions. 21 are now `machine-verified`; the other 33
  wait for a person. `catalogue/verification/batch-1.json` lists every fix with its reason, and
  every open question. The check ran on `gemini-3.8-flash`, the model that generated the records,
  because it was the only Gemini model available; the fix pass, from another provider, is the
  independent half.

- **Format (addition):** review status `machine-verified`, between `machine-generated` and
  `community-reviewed`, with `review.models` naming the independent models that checked the record
  (openexercise/open-exercise#15). `oes validate` refuses one without `by`, `at` and two distinct
  models. Consumers may show such a record's text, labelled as checked automatically.

## v0.3.0 (2026-11-05)

Benches with built-in uprights. A format change, though only an addition: two new equipment terms,
and SPEC §8 now says one item may fill more than one equipment group. An app that checks equipment
should let one item a person has satisfy every group it appears in.

- **Vocabulary (additions):** `flat-bench-with-uprights` and `incline-bench-with-uprights`, a bench
  with hooks that hold the bar (openexercise/open-exercise#10). SPEC §8 now says one item may fill
  more than one equipment group.
- **Equipment:** the barbell bench press and its six flat variants accept a flat bench with
  uprights for both the bench and the rack, and the two barbell incline presses an incline one.
  The incline presses no longer accept a flat `bench`. Pin and overhead presses still need a rack.

## v0.2.0 (2026-11-05)

The maintainer's pass over `catalogue/REVIEW.md`: duplicates folded together, wrong merges and
misleading aliases fixed, and setup details the names don't fix removed. 817 records: 762 active,
11 deprecated and 44 replaced. Apps that loaded v0.1.0 act on the replaced, reinstated and moved
entries below.

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
- **Format (wording only):** SPEC §6 now says a name fixes the grip its exercise is conventionally
  done with (a curl `supinated`, a bench press `pronated`), and no width unless it says one or the
  exercise is defined by it. No field or term changes.
- **Review:** the 13 set-aside or replaced-from-set-aside records say `unreviewed`. `review.by` is
  the latest reviewer ([GUIDE.md](GUIDE.md)). No record has been reviewed one by one yet: all active records are still
  `machine-generated`.

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
