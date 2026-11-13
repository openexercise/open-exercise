# Contributing

Open Exercise is the Open Exercise Schema, its vocabularies, the `oes` tools and the Open Exercise
Catalogue. Anyone can help through issues and pull requests. This guide is for someone outside the
project: how to set up, where things are, how to correct, add and review records, how to propose a
format change, and what a pull request needs to be merged.

If you're building an app on the schema rather than changing it, read the
[developer guide](docs/developer-guide.md) and [GUIDE.md](GUIDE.md) instead.

Everyone taking part follows the [code of conduct](CODE_OF_CONDUCT.md). Report security problems
privately, as [SECURITY.md](SECURITY.md) describes. [GOVERNANCE.md](GOVERNANCE.md) says who
decides and how.

## Ways to help

In order of how much they're needed:

1. **Review records.** Every catalogue record was written by a model from [wger](https://wger.de)
   data, and no person has checked one yet. Reviewing is how a record becomes trustworthy.
   [`catalogue/REVIEW.md`](catalogue/REVIEW.md) and the open questions in
   `catalogue/verification/batch-2.json` say where to start.
2. **Correct a record** you know is wrong, even if you don't review the rest of it.
3. **Add a record** for an exercise people log that the catalogue lacks.
4. **Report a problem** in an issue: a wrong value, a duplicate, a misleading alias.
5. **Propose a format or vocabulary change**, with real records it would improve.
6. **Improve the tools** in `cmd/` and `internal/`.

## Set up

You need Git and Go at the version in `go.mod` (`go 1.26.2`). [`jq`](https://jqlang.org) helps for
reading records and isn't required.

Fork the repository on GitHub, then:

```bash
git clone https://github.com/<you>/open-exercise
cd open-exercise
go test ./...
go run ./cmd/oes validate
go run ./cmd/oes site dist
```

On a clean checkout these end with:

```
ok  	github.com/openexercise/open-exercise/internal/oes	0.286s
825 documents, 0 problems
site: 821 pages for 806 exercises in dist
```

`validate` checks the examples and the catalogue against the schema, the vocabularies and the
cross-record rules. `site` builds the website into `dist/`, which git ignores; open
`dist/index.html` or serve the folder to look at it.

## What's where

| Path | What it holds | Licence |
| --- | --- | --- |
| `spec/v0/SPEC.md` | The specification: what a record means | Apache-2.0 |
| `schema/v0/exercise.json` | The JSON Schema (draft 2020-12) | Apache-2.0 |
| `vocab/v0/*.json` | Muscles, joints, joint actions, movement patterns, modalities, equipment, metrics | Apache-2.0 |
| `examples/` | Eight records that show the format. They illustrate the schema and aren't reviewed training advice | Apache-2.0 |
| `catalogue/exercises/<slug>.json` | The catalogue, one file per record, named after the ID's slug | CC-BY-SA 4.0 |
| `catalogue/REVIEW.md` | Lists of records worth a person's look, with the maintainer's decisions | CC-BY-SA 4.0 |
| `catalogue/verification/batch-*.json` | What the verification pipeline fixed in each record, and its open questions | CC-BY-SA 4.0 |
| `catalogue/REPORT.md` | What the first import's merge did to the wger data | CC-BY-SA 4.0 |
| `sources/wger/exercises.json` | The wger snapshot the catalogue was built from (fetched 2026-10-29) | CC-BY-SA 4.0 |
| `sources/enrichment/`, `sources/relations/` | The model answers the first import was built from | CC-BY-SA 4.0 |
| `internal/oes/` | Validation (`oes.go`), the record type (`types.go`), marking reviews (`review.go`) and their tests | Apache-2.0 |
| `cmd/oes/` | The `oes` command: `validate`, `review`, `site`, and the website's templates | Apache-2.0 |
| `cmd/enrich/`, `cmd/verify/` | The import and verification pipelines (they call Vertex AI) | Apache-2.0 |
| `GUIDE.md`, `docs/developer-guide.md` | For apps using the catalogue and the schema | Apache-2.0 |
| `CHANGELOG.md` | Every release, plus `Still to do` and `Unreleased` | Apache-2.0 |
| `.github/workflows/` | CI, the website deploy and releases | Apache-2.0 |

The website is built from `main` and served at <https://openexercise.org>. It renders the spec,
GUIDE.md and CHANGELOG.md, so a change to those shows there after it merges.

## How the catalogue was made, and why you edit records by hand

The first import turned 939 wger entries into records with Gemini (`go run ./cmd/enrich generate`
and `relate`) and merged entries with the same canonical name (`go run ./cmd/enrich merge`). Since
then every change has been made to the files in `catalogue/exercises` directly: the
[REVIEW.md](catalogue/REVIEW.md) decisions, new records, and the verification batches
(`cmd/verify`).

So **edit `catalogue/exercises/<slug>.json` directly**, and **don't run `enrich merge`**: it
deletes `catalogue/exercises` and rebuilds it from the October snapshot, which throws away every
later fix, every verification and six records.

## Correct a record

A worked example, using an open question the verification pipeline left for a person.

**1. Find the question.** `catalogue/verification/batch-2.json` lists, per record, the fixes the
pipeline made and what it wasn't sure of:

```bash
jq '.records[] | select(.id == "oec:back-extension-y-w-combo") | .uncertain' catalogue/verification/batch-2.json
```

```
[
  "Load is bodyweight-plus-external with convention total and implements 2, though the setup allows a single plate; a person should settle the load convention.",
  "Stability ball is listed as a required alternative, but the setup and steps only describe locking the feet into a Roman chair or GHD."
]
```

**2. Read the source.** The record's `external_ids.wger` is `1487`. The wger text it came from is
in the snapshot:

```bash
jq -r '.exercises[] | select(.id == 1487) | .name, .description' sources/wger/exercises.json
```

The description begins "Use a glute-ham raise or a stability ball for this exercise". So the
equipment is right, and the instructions are what's missing: they only describe the Roman chair
and the GHD. Check against a source you can cite, not only against wger, and say which in the pull
request.

**3. Edit the record.** Open `catalogue/exercises/back-extension-y-w-combo.json` in an editor and
fix the setup step. Then update `provenance.generated`, which lists JSON pointers to values no
person has checked. You checked all of `equipment`, so remove `/equipment`. You checked only the
setup part of `instructions`, so replace `/instructions` with pointers to the parts you didn't
check:

```diff
     "setup": [
-      "Lock your feet securely into a Roman chair, 45-degree hyperextension apparatus, or GHD with your hips resting against the pad.",
+      "Lock your feet securely into a Roman chair, 45-degree hyperextension apparatus, or GHD with your hips resting against the pad, or lie face down over a stability ball with your feet braced.",
       "Hinge forward at the hips to lower your torso toward the floor.",
@@
       "/anatomy",
-      "/equipment",
       "/measurement",
-      "/instructions",
+      "/instructions/steps",
+      "/instructions/breathing",
+      "/instructions/cues",
+      "/instructions/common_faults",
+      "/instructions/safety",
       "/setup",
```

Leave `provenance.review` alone: the record is still `machine-generated` until a person has checked
every value.

**4. Validate.**

```bash
go run ./cmd/oes validate
```

```
825 documents, 0 problems
```

Run it with no arguments. Validating one file on its own reports every relation as pointing at a
record that doesn't exist, because the cross-record checks only see the files you give them.

**5. Add a changelog entry and open a pull request** (see below). The second question, the load
convention, can stay open: say so in the pull request.

When you edit by hand, keep the file's style: two-space indentation and a newline at the end. Most
records have their top-level keys in alphabetical order; keep the order the file already has.
Nothing enforces it, but a small diff is easier to review.

## Review a record

Reviewing is how records go from `machine-generated` to trusted. [`catalogue/REVIEW.md`](catalogue/REVIEW.md)
lists what to look at first, and its section 5 says what to check: names and aliases, required
equipment, muscles, instructions and safety, measurement, and setup.

1. Check the values against your knowledge and a source you can cite.
2. Correct what's wrong.
3. Remove each pointer you checked from `provenance.generated`, or check them all.
4. Mark it: `go run ./cmd/oes review --by <your name or handle> oec:<slug>`. It sets
   `provenance.review` to `community-reviewed` with `by` and today's date, and refuses while
   `provenance.generated` lists values; `--checked` says you checked every one. A qualified coach
   or clinician uses `--level expert-reviewed`.
5. Open a pull request with the records you reviewed. `oes validate` refuses a reviewed record that
   still lists generated values or lacks `by` and `at`.

What the command does, on a record whose pointers are still all there:

```bash
go run ./cmd/oes review --by octocat oec:dumbbell-bench-press
```

```
oec:dumbbell-bench-press: provenance.generated still lists /name, /summary, /classification, /anatomy, /equipment, /measurement, /instructions, /setup, /relations: check them, remove each you checked, or say you checked them all
```

Once you've checked every value:

```bash
go run ./cmd/oes review --by octocat --checked oec:dumbbell-bench-press
```

```
oec:dumbbell-bench-press: community-reviewed by octocat on 2026-11-13
```

It changes only `provenance`: `generated` goes, and `review` becomes
`{"status": "community-reviewed", "at": "2026-11-13", "by": "octocat"}`. `--at` sets another date.
Use your GitHub handle for `by`; it holds the latest reviewer, and git history keeps earlier ones.

`expert-reviewed` needs a reviewer the maintainers list in [GOVERNANCE.md](GOVERNANCE.md) as
qualified, reviewing in their field. None is listed yet; if you're a certified coach,
physiotherapist or similar and want to review, open an issue.

## Add a record

Look first. Search <https://openexercise.org/catalogue/> and the aliases: the exercise may exist
under another name, in which case add an alias instead. An exercise is a separate record only when
its numbers aren't directly comparable with an existing one: another implement, stance, grip,
angle, or one side at a time (spec section 3).

Then:

1. **Pick the ID.** `oec:` plus the name in lowercase, with each run of other characters turned
   into one hyphen: "Barbell Good Morning" is `oec:barbell-good-morning`. Check that no record has
   it as its `id` or in `previous_ids`.
2. **Create `catalogue/exercises/<slug>.json`.** The file name must be the slug: `oes review` finds
   records by it. Copy a close sibling and change it, or start from one of the `examples/`.
3. **Fill in what you can support.** Required: `id`, `name`, `measurement`. Aim for what the spec
   calls Standard: `classification`, `anatomy.muscles`, `equipment`, `provenance.license`. Follow
   the spec on names (section 4), setup facets (section 6), muscles (section 7), equipment groups
   (section 8) and especially `measurement.load` (section 9.1).
4. **Set `provenance`.** `"license": "CC-BY-SA-4.0"`, `sources` for anything you drew on (see
   below), and `"review": {"status": "unreviewed"}`. Don't list `generated`: no machine wrote it.
   The record stays `unreviewed` until someone else reviews it. `oec:hiking` and
   `oec:barbell-good-morning` were added this way.
5. **Link it.** Add `relations` from the new record, and to it from close siblings where it helps.
   If an alias on another record really means the new exercise, move it.
6. **Validate**, add a changelog entry, and open a pull request.

Don't add `measurement.load.bodyweight_fraction` or `energy` without a published source to put in
their `source`: the spec says they're never estimated.

## Rename, retire or replace a record

Apps store IDs, so these rules keep their data resolving:

- **Don't change a published ID.** If a rename is truly needed, create the record under the new ID
  and add the old one to its `previous_ids`. `oes validate` refuses an ID that is both current and
  another record's previous ID, or a previous ID claimed by two records.
- **Don't delete a record.** Set `status` to `deprecated` (keep resolving, stop offering it for new
  logs) or `replaced` with `replaced_by` naming the record to use. A `replaced_by` chain may not
  loop.
- When one record replaces another (a duplicate merged), the survivor gains the old record's names
  as aliases and its sources. The replaced record keeps its own `external_ids`, aliases and
  `previous_ids`, so a lookup by any of them still finds it and follows `replaced_by`.
- An alias that misleads (it names another exercise too) becomes `legacy` rather than being
  dropped, so old imports still resolve.

Each of these is something an app must act on, so its changelog entry says so plainly.

## Sources and credits

The catalogue is CC-BY-SA 4.0 because the wger data it derives from requires it.

- Only contribute text you wrote, or text you may license under CC-BY-SA 4.0. Don't paste from
  books, websites or other apps' databases.
- Credit what you drew on in `provenance.sources`: each entry has a `title` (required) and
  optionally `url`, `license` and `author`. The wger-derived records use
  `{"title": "wger: <name>", "url": "https://wger.de/en/exercise/<id>/view-base", "license": …, "author": …}`.
- When you correct a value against a published source, cite it in the pull request. Add it to
  `sources` when the record now draws on it.
- `external_ids` maps a registered system (`wger`, `free-exercise-db`) to that system's ID, several
  comma-separated. Use `x-<name>` for an unregistered one, or propose registering it by pull
  request.

## Change the format

The format is `spec/`, `schema/` and `vocab/`.

1. **Open an issue first**, with real records the change would improve. A maintainer decides after
   at least a week for comment ([GOVERNANCE.md](GOVERNANCE.md)).
2. **Then one pull request** that changes, together:
   - `spec/v0/SPEC.md`, the meaning;
   - `schema/v0/exercise.json`, the shape;
   - `vocab/v0/*.json` for new terms (`id`, `label`, and `parent` and `definition` where they help);
   - `internal/oes/types.go` and `oes.go` if `oes validate` should check a new rule, with a test in
     `internal/oes/oes_test.go`;
   - the records and examples that use it;
   - [GUIDE.md](GUIDE.md) or the [developer guide](docs/developer-guide.md) if apps should act on it;
   - `CHANGELOG.md`, marked as a format change.

New vocabulary terms and new optional fields are additions. Renaming or removing a term or field,
or changing what a value means, is breaking: until v1 it can still happen, with the changelog
saying so; after v1 it needs a new major version. `v0.3.0` is an example of an addition (two
equipment terms and a sentence in SPEC section 8).

## Change the tools

The Go code follows the standard library and the existing style: `gofmt`, errors wrapped with
context, no new dependencies without a reason. Add or update a test in `internal/oes` for any rule
`validate` checks. `cmd/enrich` and `cmd/verify` call Vertex AI and need Google Cloud credentials
and `OES_GCP_PROJECT`; you don't need them for anything else.

## Before you open a pull request

Run what CI runs (`.github/workflows/ci.yml`):

```bash
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
go run ./cmd/oes validate
go run ./cmd/oes site dist
```

Each must pass. Then check:

- [ ] Every record you touched validates, and its `provenance.generated` lists only what no person
      checked.
- [ ] No published ID changed and no record was deleted.
- [ ] New text is yours or CC-BY-SA 4.0 compatible, and sources are credited.
- [ ] `CHANGELOG.md` has an entry under `## Unreleased` if a consumer would notice the change.

## Commits and pull requests

- Work on a branch, never on `main`. Branch names follow the change: `fix/…`, `feat/…`, `docs/…`.
- Commit messages are [conventional commits](https://www.conventionalcommits.org):
  `type(scope): description`. Types used here: `feat`, `fix`, `docs`, `chore`, `ci`, `refactor`,
  `test`. Scopes used so far: `catalogue`, `vocab`, `format`, `spec`, `verify`, `review`, `guide`,
  `governance`, `changelog`, `release`. For example `fix(catalogue): verify batch 2, the rest of
  the catalogue` or `feat(vocab): benches with uprights fill the bench and the rack`.
- Write the description as what changed, in plain words. Put the why in the body.
- Keep a pull request to one topic. A batch of reviewed records is one topic; a format change and
  unrelated record fixes are two.
- In the pull request, say what you checked and against which source, and list anything you left
  open.

## The changelog

[CHANGELOG.md](CHANGELOG.md) is the running record. Every pull request a consumer would notice
adds an entry under `## Unreleased`, in the same pull request: records added, renamed, retired or
replaced, alias changes that change what matches, corrections, and format changes. Follow the
existing style: a bullet with a bold label (**Format (addition):**, **Vocabulary (additions):**,
**Equipment:**) and one or two sentences on what changed and what an app has to do. Pure refactors,
tests and typo fixes don't need one.

Keep `## Still to do` current: add what you leave undone, and remove an item when it ships.

## Releasing

Maintainers cut a release when records or the format have changed enough for apps to take:

1. Move the `Unreleased` entries in [CHANGELOG.md](CHANGELOG.md) under a new `## vX.Y.Z (date)`
   heading, saying what a consumer must act on (renames, retirements, replacements, alias changes
   that change matches, format changes), and merge.
2. Tag the merge and push the tag: `git tag v0.2.0 && git push origin v0.2.0`.
3. The Release workflow validates, and publishes a GitHub release with the changelog entry as notes
   and `catalogue.json` attached.

Until v1, versions are `v0.N.P`: `N` for releases that add, rename, retire or replace records or
change the format, `P` for corrections that don't.

## Becoming a maintainer

Someone who has made sustained, careful contributions can be invited to become a maintainer by the
existing ones ([GOVERNANCE.md](GOVERNANCE.md)).
