# Developer guide

For people building an app, a service or a data pipeline on the Open Exercise Schema and the Open
Exercise Catalogue. It covers getting the data, what each field of a record means in practice,
checking your own records with the `oes` tool, and staying correct as releases change.

Three other documents sit beside this one:

- The [specification](../spec/v0/SPEC.md) says what a record means. Where this guide and the spec
  disagree, the spec wins.
- [GUIDE.md](../GUIDE.md), "Using the catalogue", has the rules for resolving stored IDs, alias
  kinds, the normalization rule, a worked matching example and how far to trust a record's text.
  This guide links to it rather than repeating it.
- [CONTRIBUTING.md](../CONTRIBUTING.md) is for sending corrections and changes back.

Every command and program below was run against release `v0.5.0`, and the output shown is what it
printed. The Go examples use only the standard library.

## 1. Get the catalogue

### From a release (use this in production)

Each [release](https://github.com/openexercise/open-exercise/releases) carries one asset,
`catalogue.json`: every record in the catalogue, as a JSON array sorted by name.

```bash
curl -sSLO https://github.com/openexercise/open-exercise/releases/download/v0.5.0/catalogue.json
jq length catalogue.json
```

```
817
```

The array holds every record, retired ones included, because a stored ID must keep resolving:

```bash
jq -r 'group_by(.status) | map("\(.[0].status) \(length)") | .[]' catalogue.json
```

```
active 762
deprecated 11
replaced 44
```

Store the tag you loaded (`v0.5.0`) next to your data. Before loading a newer one, read its
[changelog](../CHANGELOG.md) entry: it lists what a consumer has to act on.

### From a git tag

To get the schema, vocabularies and tools at the same version as the records, clone the tag:

```bash
git clone --depth 1 --branch v0.5.0 https://github.com/openexercise/open-exercise
cd open-exercise
go run ./cmd/oes validate
```

```
825 documents, 0 problems
```

The 825 documents are the 817 catalogue records and the 8 records in `examples/`.

### From openexercise.org (for browsing, not pinning)

The website is rebuilt from `main` on every push, so its files move between releases. It serves
them with `Access-Control-Allow-Origin: *`, which makes it handy for previews and tools:

| Path | What it is |
| --- | --- |
| `/catalogue.json` | Every record, as in a release asset, but from `main` |
| `/exercises/<slug>.json` | One record, such as `/exercises/barbell-bench-press.json` |
| `/catalogue/index.json` | The site's search index: ID, name, summary, alias texts, movement patterns, and muscles and equipment with their parent terms. It leaves out `deprecated` records but keeps `replaced` ones |
| `/schema/v0/exercise.json` | The JSON Schema, at the URL every record names in `$schema` |
| `/vocab/v0/<name>.json` | The vocabularies |

`go run ./cmd/oes site dist` builds the same files locally into `dist/`.

## 2. A record, field by field

A record is one JSON object, valid against [`schema/v0/exercise.json`](../schema/v0/exercise.json)
(JSON Schema draft 2020-12). Only `id`, `name` and `measurement` are required. Unknown top-level
members are refused unless they start with `x-` (section 6).

The examples below come from `oec:dumbbell-bench-press` in `v0.5.0`. The count after each field is
how many of the 817 records have it.

### Identity and lifecycle

```json
{
  "$schema": "https://openexercise.org/schema/v0/exercise.json",
  "id": "oec:dumbbell-bench-press",
  "status": "active",
  "previous_ids": ["oec:benchpress-dumbbells"]
}
```

| Field | Records | Meaning |
| --- | --- | --- |
| `$schema` | 817 | Always `https://openexercise.org/schema/v0/exercise.json` in v0 (the schema makes it a constant) |
| `id` | 817 | `namespace:slug`, matching `^[a-z][a-z0-9]*:[a-z0-9]+(-[a-z0-9]+)*$` (section 5) |
| `status` | 817 | `active`, `deprecated` or `replaced`. Absent means `active` |
| `replaced_by` | 44 | Required when `status` is `replaced`: the record to use instead |
| `previous_ids` | 634 | IDs that once meant this record |

How to act on these when a release changes them is in [section 7](#7-load-a-new-release) and in
GUIDE.md, [Resolve an ID you stored](../GUIDE.md#resolve-an-id-you-stored).

### Names

```json
{
  "name": "Dumbbell Bench Press",
  "summary": "A compound free-weight exercise performed lying on a flat bench to build chest, anterior shoulder, and triceps strength and hypertrophy.",
  "aliases": [
    { "text": "dumbbell bench", "kind": "common" },
    { "text": "flat dumbbell press", "kind": "common" },
    { "text": "DB bench press", "kind": "abbreviation" },
    { "text": "benchpress dumbbells", "kind": "common" }
  ]
}
```

| Field | Records | Meaning |
| --- | --- | --- |
| `name` | 817 | Canonical English name, title case, singular, implement first |
| `summary` | 804 | One plain sentence |
| `aliases` | 800 | Objects with `text`, optional `lang` (default `en`) and `kind` (default `common`) |
| `translations` | 0 | Other languages, keyed by BCP 47 tag: `name`, `summary`, `aliases` (plain strings) and `instructions` |

The six alias kinds and how far to trust each for matching are in GUIDE.md,
[Names and aliases](../GUIDE.md#names-and-aliases). Show `name`; never show a `misspelling` alias.

### Classification

```json
{
  "classification": {
    "modalities": ["strength", "hypertrophy"],
    "movement_patterns": ["horizontal-push"],
    "mechanics": "compound",
    "force": "push",
    "kinetic_chain": "open",
    "laterality": "bilateral",
    "planes": ["transverse", "sagittal"],
    "contraction": ["dynamic"],
    "level": "beginner"
  }
}
```

804 records have it. `modalities` and `movement_patterns` are vocabulary terms, most important
first. The other members are closed lists in the schema; the spec's
[section 5](../spec/v0/SPEC.md#5-classification) gives each value's meaning. `tags` (free strings)
is allowed, and no catalogue record uses it yet. `level` is technical demand, not how heavy the exercise can be.

### Setup

```json
{
  "setup": {
    "grip": { "orientation": "pronated" },
    "incline_degrees": 0,
    "position": "supine",
    "range": "full",
    "stance": { "kind": "parallel" }
  }
}
```

801 records have it. These facets tell a variation from its siblings: `position`,
`incline_degrees` (−90 to 90, positive is incline), `grip.orientation`, `grip.width`,
`stance.width`, `stance.kind`, `load_position` and `range`. A facet is left out when the exercise
doesn't fix it, so absent means "any", not "unknown". Spec [section 6](../spec/v0/SPEC.md#6-setup)
says when a facet belongs.

Use setup to filter and to tell siblings apart ("show me seated presses"), not to describe how
someone did a particular set. A different grip is a different record.

### Anatomy

```json
{
  "anatomy": {
    "muscles": [
      { "muscle": "pectoralis-major-sternal", "role": "prime-mover" },
      { "muscle": "pectoralis-major-clavicular", "role": "prime-mover" },
      { "muscle": "deltoid-anterior", "role": "synergist" },
      { "muscle": "triceps-brachii", "role": "synergist" },
      { "muscle": "rotator-cuff", "role": "stabilizer" }
    ],
    "joints": [
      { "joint": "glenohumeral", "actions": ["horizontal-adduction", "flexion"] },
      { "joint": "elbow", "actions": ["extension"] }
    ]
  }
}
```

(The record lists two more joints; this excerpt shortens it.) 802 records have anatomy. `role` is
`prime-mover`, `synergist`, `stabilizer` or `stretched`. Muscles are the most specific term the
record can support, and never a term together with its parent, so to answer "does this work the
chest?" walk up the muscle tree (section 4). There are no percentages in v0.

### Equipment

```json
{
  "equipment": {
    "required": [["dumbbell"], ["bench", "adjustable-bench"]]
  }
}
```

804 records have it. `required` is a list of groups: every group is needed, and any one item of a
group fills it. An empty `required` means no equipment; an absent `equipment` means unknown.
`optional` lists items that help but aren't needed. A parent term means any of its children, and
one item may fill several groups. Section 4 has code for "can this person do it with what they
have?".

### Measurement

```json
{
  "measurement": {
    "profiles": [{ "required": ["reps", "load"], "optional": ["rpe"] }],
    "load": { "kind": "external", "convention": "per-implement", "implements": 2 },
    "reps_counted": "total",
    "records": { "e1rm": true, "volume": true }
  }
}
```

Every record has it. It's the part most worth reading carefully, since it decides what a logged
number means. Section 3 goes through it.

### Instructions

803 records have `instructions`: `setup` and `steps` (arrays of sentences), `breathing` (a
string), `cues` (short phrases), `common_faults` (objects with `fault` and `correction`) and
`safety` (`precautions`, `spotter`: `required`, `recommended` or `not-needed`, and `safeties`, a
boolean for rack safeties). They are general education and never prescribe sets, reps or loads.

Most of this text was written by a model. Before showing it, read
[Show what's trustworthy](../GUIDE.md#show-whats-trustworthy) in GUIDE.md.

### Relations

```json
{
  "relations": [
    { "type": "substitute-for", "target": "oec:barbell-bench-press" },
    { "type": "progression-of", "target": "oec:dumbbell-floor-press" }
  ]
}
```

670 records have relations. Read each as "this exercise *type* target": the dumbbell bench press
is a progression of the floor press. Types are `variation-of`, `regression-of`, `progression-of`,
`substitute-for`, `antagonist-of`, `part-of` and `same-as`. Only `same-as` and `antagonist-of` are
symmetric, so build reverse edges yourself if you need them ("what progresses from this?").
`oes validate` checks that every target exists in the catalogue.

### External IDs

```json
{ "external_ids": { "wger": "75" } }
```

812 records have a `wger` ID. When several wger entries were merged into one record, they're
comma-separated (`oec:barbell-back-squat` has `"wger": "615,1801"`). The registered systems are `wger` and `free-exercise-db`;
anything else uses an `x-<name>` key.

### Energy and media

Both are in the schema and no catalogue record uses them yet. `energy` is `met` with a `source`
URI and must cite one. `media` is a list of objects with `kind` (`image`, `video`, `animation`,
`illustration`, `model-3d`), `url`, `license`, and optional `view`, `alt` and `author`.

### Provenance

```json
{
  "provenance": {
    "license": "CC-BY-SA-4.0",
    "sources": [
      {
        "title": "wger: Benchpress Dumbbells",
        "url": "https://wger.de/en/exercise/75/view-base",
        "license": "CC-BY-SA-3.0",
        "author": "wger.de"
      }
    ],
    "review": { "status": "machine-generated", "model": "gemini-3.8-flash", "at": "2026-10-31" },
    "generated": ["/name", "/summary", "/classification", "/anatomy", "/equipment", "/measurement", "/instructions", "/setup", "/relations"]
  }
}
```

Every record has it. `license` is the record's SPDX ID. `sources` are what it derives from, and
you must credit them (section 9). `review.status` says who has checked it; `generated` lists JSON
pointers (RFC 6901) to values no person has checked. In `v0.5.0`:

```bash
jq -r 'group_by(.provenance.review.status) | map("\(.[0].provenance.review.status) \(length)") | .[]' catalogue.json
```

```
machine-generated 382
machine-verified 418
unreviewed 17
```

A `machine-verified` record also has `review.models` (at least two) and `review.by`, the pipeline
and its version, such as `oes-verify/1`.

## 3. Record sets with `measurement`

### Units

Exchange data in the units fixed by [`vocab/v0/metrics.json`](../vocab/v0/metrics.json):

| Metric | Unit |
| --- | --- |
| `reps` | count |
| `load` | kilograms, meaning set by `measurement.load` |
| `duration` | seconds |
| `distance`, `height` | metres |
| `speed` | metres per second |
| `pace` | seconds per kilometre |
| `power` | watts |
| `energy` | kilojoules |
| `calories` | kilocalories, as a machine reports them |
| `heart-rate` | beats per minute |
| `cadence` | cycles per minute |
| `incline` | percent grade |
| `resistance-level` | a machine setting, no unit |
| `band-tension` | a named band colour or level |
| `rpe` | 1 to 10 |
| `rir` | repetitions in reserve |
| `tempo` | eccentric, pause, concentric and top pause in seconds, as in `3-1-1-0` |

Show whatever units your users want. Convert at the edge.

### Profiles

`measurement.profiles` lists ways to record a set, the usual one first. A set fits a profile when
it has every `required` metric. `oec:running` has two:

```json
{
  "profiles": [
    { "required": ["distance", "duration"], "optional": ["pace", "speed", "heart-rate", "cadence", "calories", "rpe"] },
    { "required": ["duration"], "optional": ["distance", "pace", "speed", "heart-rate", "cadence", "calories", "rpe"] }
  ]
}
```

Build your logging form from the first profile, and accept a set that fits any of them.

### What the load number means

`measurement.load` is present whenever a profile records `load`. Its `kind` says whether the
number is the weight, something added to bodyweight, or assistance; its `convention` says what
the number counts. Real records from `v0.5.0`:

| Record | `kind` | `convention` | A logged 40 means |
| --- | --- | --- | --- |
| `oec:barbell-bench-press` | `external` | `total`, `implements: 1` | 40 kg on the bar, bar included |
| `oec:dumbbell-bench-press` | `external` | `per-implement`, `implements: 2` | Two 40 kg dumbbells, 80 kg moved |
| `oec:plate-loaded-chest-press` | `external` | `per-side`, `implements: 2` | 40 kg on each side, 80 kg moved |
| `oec:cable-seated-row` | `external` | `weight-stack` | Pin at 40; not comparable with another machine |
| `oec:back-extension` | `bodyweight-plus-external` | `total` | 40 kg added to bodyweight |
| `oec:machine-assisted-dip` | `bodyweight-minus-assistance` | `weight-stack` | 40 on the assistance stack: a bigger number is easier |
| `oec:band-pull-apart` | `elastic` | `not-applicable` | No load is logged; `band-tension` is optional |
| `oec:plank` | `bodyweight` | (none) | No load is logged; the profile records `duration` |

For the bodyweight kinds, `load.bodyweight_fraction` (`value`, between 0 and 1, and a `source`
URI) says what fraction of body mass the exercise moves. It must cite a source, and no catalogue
record has one yet, so don't estimate it from bodyweight yourself and present it as data.

To compare loads across exercises or apps, turn the logged number into kilograms moved (spec
[section 9.1](../spec/v0/SPEC.md#91-what-the-load-number-means)):

```go
type Load struct {
	Kind       string `json:"kind"`
	Convention string `json:"convention"`
	Implements int    `json:"implements"`
}

// MovedLoad turns a logged load number into the kilograms moved (SPEC §9.1).
// ok is false when the number can't be compared: a weight stack, or no external load.
func MovedLoad(logged float64, l Load) (kg float64, ok bool) {
	if l.Kind != "external" {
		return 0, false // bodyweight kinds and elastic need more than the logged number
	}
	switch l.Convention {
	case "total":
		return logged, true
	case "per-implement":
		return logged * float64(l.Implements), true
	case "per-side":
		return logged * 2, true
	}
	return 0, false // weight-stack: the pin number means nothing on another machine
}
```

With 100 `total`, 30 `per-implement` ×2, 40 `per-side` and 70 `weight-stack`, it returned
`100 true`, `60 true`, `80 true` and `0 false`.

### Counting reps and time

`reps_counted` and `duration_counted` are `total`, `per-side` or `not-applicable`. Ten reps of a
`per-side` exercise means ten for each side. Store what the person did on one side and show
"each side"; don't double it.

### Computed numbers

`records.e1rm` says an estimated one-rep max means something for this exercise, and
`records.volume` that load × reps is a fair measure of work. When a flag is false or absent, don't
show the number. `oec:running` has `"records": {}`, so an app shows neither.

## 4. Vocabularies

Seven files in [`vocab/v0`](../vocab/v0) hold the terms records use:

| File | Used in |
| --- | --- |
| `modalities.json` | `classification.modalities` |
| `movement-patterns.json` | `classification.movement_patterns` |
| `muscles.json` | `anatomy.muscles[].muscle`: a tree of region, group, muscle, head |
| `joints.json` | `anatomy.joints[].joint` |
| `joint-actions.json` | `anatomy.joints[].actions` |
| `equipment.json` | `equipment.required` and `equipment.optional` |
| `metrics.json` | `measurement.profiles` |

Each file is `{"vocabulary", "description", "terms"}`, and each term has `id`, `label`, and
optionally `parent` and `definition`. Show `label`, store `id`.

Three rules to build in:

1. **Walk up the tree.** A term implies its ancestors. To find every record that works the chest,
   compare against each muscle's ancestors, not the muscle alone:

   ```bash
   jq -r --arg term pectoralis-major-sternal \
     '(.terms | map({key: .id, value: .parent}) | from_entries) as $parent
      | [$term | recurse($parent[.] // empty)] | join(" > ")' vocab/v0/muscles.json
   ```

   ```
   pectoralis-major-sternal > pectoralis-major > chest > trunk-anterior
   ```

2. **Tolerate terms you don't know.** New terms are additions, not breaking changes, so a newer
   release may use a term your copy of the vocabulary lacks. Show the ID, or fall back to its
   parent once you have the newer vocabulary.
3. **Expect `x-` terms.** Anyone may use `x-<term>` for a private term. Treat one you don't know as
   unknown.

For equipment, a record's term covers the item a person has when it's the same term or one of its
ancestors, and one item may fill several groups. This checks a record's `equipment.required`
against what someone has:

```go
// covers reports whether an item a person has is the term or one of its children.
func covers(term, item string, parent map[string]string) bool {
	for at := item; at != ""; at = parent[at] {
		if at == term {
			return true
		}
	}
	return false
}

// canDo reports whether the items fill every group of equipment.required.
// One item may fill more than one group (SPEC §8).
func canDo(required [][]string, have []string, parent map[string]string) bool {
	for _, group := range required {
		filled := slices.ContainsFunc(group, func(term string) bool {
			return slices.ContainsFunc(have, func(item string) bool { return covers(term, item, parent) })
		})
		if !filled {
			return false
		}
	}
	return true
}
```

With `parent` built from `vocab/v0/equipment.json` and the barbell bench press's groups
(`[["barbell"], ["bench", "flat-bench-with-uprights"], ["power-rack", "squat-rack", "flat-bench-with-uprights"]]`),
`{"barbell", "flat-bench-with-uprights"}` returned `true` (the bench fills two groups),
`{"barbell", "bench"}` returned `false` (no rack), and a record needing `free-weights` was
satisfied by `{"dumbbell"}`.

## 5. ID namespaces

| Namespace | Who uses it | Meaning |
| --- | --- | --- |
| `oec:` | This catalogue only | Resolvable against a release |
| `usr:` | Any app, for a person's own exercise | Means nothing outside the system that made it |
| your own, such as `acme:` | You, for exercises you publish | Yours to manage |

Rules that follow:

- Never mint an `oec:` ID. If the catalogue lacks an exercise, use your own namespace and propose
  the record here ([CONTRIBUTING.md](../CONTRIBUTING.md)). When it lands, it gets an `oec:` ID.
- Store `oec:` IDs, not names. Store the name you showed with it too, so a log still reads well if
  an ID ever fails to resolve.
- To say your record is the catalogue's exercise under another ID, add
  `{"type": "same-as", "target": "oec:…"}` to your record.
- An exercise is a separate record when its numbers aren't directly comparable: another implement,
  stance, grip, angle or side. Don't merge them in your app either.

## 6. Your own records and extensions

Your records use the same schema. Here is a record an app might publish for an exercise the
catalogue doesn't have:

```json
{
  "$schema": "https://openexercise.org/schema/v0/exercise.json",
  "id": "acme:half-kneeling-landmine-press",
  "name": "Half-Kneeling Single-Arm Landmine Press",
  "classification": {
    "modalities": ["strength"],
    "movement_patterns": ["vertical-push"],
    "laterality": "unilateral"
  },
  "setup": { "position": "half-kneeling" },
  "equipment": { "required": [["barbell"], ["landmine"]] },
  "measurement": {
    "profiles": [{ "required": ["reps", "load"], "optional": ["rpe"] }],
    "load": { "kind": "external", "convention": "total", "implements": 1 },
    "reps_counted": "per-side",
    "records": { "e1rm": true, "volume": true }
  },
  "relations": [{ "type": "variation-of", "target": "oec:landmine-press" }],
  "provenance": { "license": "CC0-1.0" },
  "x-acme-video": "https://example.com/videos/hk-landmine-press.mp4"
}
```

### Extensions

- Add your own data as top-level `x-<name>` members, such as `x-acme-video`. The name after `x-`
  is lowercase letters, digits and hyphens. Consumers ignore ones they don't understand.
- Only the top level takes `x-` members. Inside `measurement` or any other object, unknown members
  are refused.
- Vocabulary values may be `x-<term>` anywhere a term is expected.
- An `external_ids` key for a system that isn't registered is `x-<name>`.

### Validate with the `oes` tool

Run `oes` from a checkout of this repository (it reads `schema/` and `vocab/` from the current
directory). Give it a directory or file:

```bash
go run ./cmd/oes validate ~/acme/exercises
```

`validate` checks each file against the JSON Schema, then the rules a schema can't state: every
term is in its vocabulary (or `x-`), a profile that records `load` has a `measurement.load`, an
`external` load has a real convention, `per-implement` has `implements` of 2 or more, `e1rm` only
with external load, no muscle listed twice or with its parent, `provenance.generated` pointers
that resolve, and review statuses with their `by`, `at` and `models`. Then it checks the files of
each directory as one catalogue: unique IDs, relation and `replaced_by` targets that exist, no
`replaced_by` loops, and no ID that's both current and another record's previous ID.

Because of that last step, a record that relates to an `oec:` record fails on its own (path
shortened):

```
/Users/me/acme/exercises: acme:half-kneeling-landmine-press: relation target oec:landmine-press doesn't exist
1 documents, 1 problems
validation failed
```

Validate your records together with the catalogue instead. Copy both into one directory, yours
last, and check the file names don't collide (here `landmine-press.json` would have overwritten
the catalogue's own):

```bash
mkdir /tmp/oes-check
cp catalogue/exercises/*.json /tmp/oes-check/
cp ~/acme/exercises/*.json /tmp/oes-check/
go run ./cmd/oes validate /tmp/oes-check
```

```
818 documents, 0 problems
```

What failures look like (paths shortened):

```
load.json: load kind external needs a convention
load.json: external load needs a real convention
nested.json: jsonschema validation failed with 'file:///…/schema/v0/exercise.json#'
- at '/measurement': additional properties 'x-acme-note' not allowed
```

The command exits with status 1 when it finds a problem, so it works as a CI step.

Any JSON Schema draft 2020-12 validator can check the schema part, using
`https://openexercise.org/schema/v0/exercise.json`. Only `oes validate` checks the vocabulary and
cross-record rules.

## 7. Load a new release

Records are never deleted and IDs never reused, but a release can rename, retire or replace
records, and move an alias between records. Before loading one:

1. Read its [changelog](../CHANGELOG.md) entry.
2. Diff it against the release you have. This `jq` program lists records added, IDs that stopped
   being current (and the record whose `previous_ids` took them), and status changes:

   ```jq
   # Usage: jq -r -n --slurpfile old OLD.json --slurpfile new NEW.json -f release-diff.jq
   ($old[0] | map({key: .id, value: .}) | from_entries) as $before
   | ($new[0] | map({key: .id, value: .}) | from_entries) as $after
   | ($new[0] | map(.id as $id | (.previous_ids // [])[] | {key: ., value: $id}) | from_entries) as $renamed
   | ($new[0][] | select($before[.id] == null) | "added    \(.id)"),
     ($old[0][] | select($after[.id] == null)
       | "renamed  \(.id) -> \($renamed[.id] // "MISSING")"),
     ($new[0][] | select($before[.id] != null)
       | select(($before[.id].status // "active") != (.status // "active"))
       | "status   \(.id): \($before[.id].status // "active") -> \(.status)\(if .replaced_by then " (use \(.replaced_by))" else "" end)")
   ```

   From `v0.1.0` to `v0.5.0` it printed 46 lines, starting:

   ```
   added    oec:machine-seated-dip
   added    oec:serratus-wall-slide
   status   oec:assisted-pull-up: active -> replaced (use oec:machine-assisted-pull-up)
   status   oec:bag-training: deprecated -> replaced (use oec:boxing)
   ```

   From `v0.4.0` to `v0.5.0` it printed nothing: that release changed records' content, not their
   identity.
3. Apply the changes:
   - **Renamed** (an ID you hold is now in another record's `previous_ids`): move every stored
     reference to the new ID, including IDs inside JSON you keep.
   - **`deprecated`:** keep resolving and showing it in old logs. Stop offering it for new logs and
     stop matching new text to it.
   - **`replaced`:** the same, and follow `replaced_by` to the end of the chain for new logs. Offer
     to move a person's old logs, or move them yourself when the two are the same exercise.
   - **Back to `active`:** offer it again. (`oec:boxing` was deprecated in `v0.1.0` and is active
     in `v0.5.0`.)
   - **Alias changes:** rebuild your name index, since an alias that became `legacy` should stop
     deciding matches.
4. Re-derive anything you computed from records (e1RM, volume, equipment filters), since
   `measurement` and `equipment` can be corrected.

GUIDE.md has the full rules: [Resolve an ID you stored](../GUIDE.md#resolve-an-id-you-stored) and
[Deprecated and replaced records](../GUIDE.md#deprecated-and-replaced-records).

## 8. Resolve IDs and match names

The program below loads a release, resolves stored IDs (by `id`, then `previous_ids`), follows
`replaced_by`, and does exact matching of typed names with GUIDE.md's
[normalization rule](../GUIDE.md#normalize-before-comparing). It skips `deprecated` and `replaced`
records and `legacy` aliases when matching.

```go
// Resolve stored IDs and match typed names against an Open Exercise release.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Alias struct {
	Text string `json:"text"`
	Kind string `json:"kind"`
}

type Exercise struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	ReplacedBy  string   `json:"replaced_by"`
	PreviousIDs []string `json:"previous_ids"`
	Aliases     []Alias  `json:"aliases"`
}

type Catalogue struct {
	byID       map[string]*Exercise
	byPrevious map[string]*Exercise
	byName     map[string][]*Exercise
}

func Load(path string) (*Catalogue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var all []*Exercise
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, err
	}
	c := &Catalogue{map[string]*Exercise{}, map[string]*Exercise{}, map[string][]*Exercise{}}
	for _, e := range all {
		c.byID[e.ID] = e
		for _, p := range e.PreviousIDs {
			c.byPrevious[p] = e
		}
		if e.Status == "deprecated" || e.Status == "replaced" {
			continue // never match new text to a retired record
		}
		c.add(e.Name, e)
		for _, a := range e.Aliases {
			if a.Kind != "legacy" { // legacy aliases only suggest
				c.add(a.Text, e)
			}
		}
	}
	return c, nil
}

func (c *Catalogue) add(name string, e *Exercise) {
	key := Normalize(name)
	for _, have := range c.byName[key] {
		if have == e {
			return
		}
	}
	c.byName[key] = append(c.byName[key], e)
}

// Resolve finds the record for a stored ID: by id, then by previous_ids (a rename).
func (c *Catalogue) Resolve(id string) (*Exercise, bool) {
	if e, ok := c.byID[id]; ok {
		return e, true
	}
	e, ok := c.byPrevious[id]
	return e, ok
}

// Current follows replaced_by to the end of the chain: the record to use for new logs.
func (c *Catalogue) Current(e *Exercise) *Exercise {
	for e.Status == "replaced" {
		e = c.byID[e.ReplacedBy]
	}
	return e
}

// Normalize applies GUIDE.md's rule: lowercase, separators to one space, trim.
func Normalize(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if isSeparator(r) {
			space = true
			continue
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	return b.String()
}

func isSeparator(r rune) bool {
	switch {
	case r >= 0x01 && r <= 0x2F, r >= 0x3A && r <= 0x40, r >= 0x5B && r <= 0x60, r >= 0x7B && r <= 0x7F:
		return true
	case r == 0xA0, r == 0xB7, r == 0xD7, r == 0x2026:
		return true
	case r >= 0x2010 && r <= 0x2015, r >= 0x2018 && r <= 0x201F:
		return true
	}
	return false
}

func main() {
	c, err := Load(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, id := range []string{"oec:barbell-bench-press", "oec:bench-press", "oec:bag-training", "oec:no-such-thing"} {
		e, ok := c.Resolve(id)
		if !ok {
			fmt.Printf("%s -> not found: keep it and show the stored name\n", id)
			continue
		}
		fmt.Printf("%s -> %s (%s), current %s\n", id, e.ID, e.Status, c.Current(e).ID)
	}
	for _, text := range []string{"Bench Press!", "DB bench press", "pull-up", "Skull-Crusher", "row"} {
		var ids []string
		for _, e := range c.byName[Normalize(text)] {
			ids = append(ids, e.ID)
		}
		fmt.Printf("%q -> %q %v\n", text, Normalize(text), ids)
	}
}
```

`go run . catalogue.json` printed:

```
oec:barbell-bench-press -> oec:barbell-bench-press (active), current oec:barbell-bench-press
oec:bench-press -> oec:barbell-bench-press (active), current oec:barbell-bench-press
oec:bag-training -> oec:bag-training (replaced), current oec:boxing
oec:no-such-thing -> not found: keep it and show the stored name
"Bench Press!" -> "bench press" [oec:barbell-bench-press]
"DB bench press" -> "db bench press" [oec:dumbbell-bench-press]
"pull-up" -> "pull up" [oec:pull-up]
"Skull-Crusher" -> "skull crusher" [oec:barbell-lying-triceps-extension oec:ez-bar-skull-crusher]
"row" -> "row" []
```

Notes on what it shows:

- `oec:bench-press` is an old ID. It resolves through `previous_ids`; store the new one.
- "Skull Crusher" is an alias of two records. One name naming several records decides nothing:
  show both and let the person choose.
- Exact matching misses a lot ("row" finds nothing, and plurals and typos miss). GUIDE.md's
  [worked example](../GUIDE.md#a-worked-example-lifterclubs-matching) adds abbreviation expansion,
  plural forms and trigram similarity with thresholds. A generic name like "row" should always ask.

## 9. Licences and attribution

Two licences apply, to different parts:

| Part | Licence |
| --- | --- |
| Schema, spec, vocabularies, the `oes` tools (`schema/`, `spec/`, `vocab/`, `cmd/`, `internal/`) | [Apache-2.0](../LICENSE) |
| The catalogue (`catalogue/`, `sources/`) and the `catalogue.json` release asset | [CC-BY-SA 4.0](../catalogue/LICENSE) |

What that means in practice:

- **Your own records** in the schema can carry the licence you choose. Put its SPDX ID in
  `provenance.license`.
- **Showing catalogue records** (names, summaries, instructions): credit them. The catalogue is
  derived from [wger](https://wger.de), and each record lists its sources in `provenance.sources`
  with `title`, `url`, `license` and `author`. A credit line that follows GUIDE.md is the record's
  sources plus "Open Exercise Catalogue, CC-BY-SA 4.0", with a link to the licence.
- **Publishing changed catalogue records** (an edited copy, or records built from them): CC-BY-SA
  4.0's share-alike terms apply to what you publish. Keep `provenance.license` as `CC-BY-SA-4.0`
  and keep the sources. Better still, send the change back (section 11).

This is a summary of the files in this repository, not legal advice. Read the licences for
anything that matters.

## 10. Versioning and what stays stable

The project is a **v0 development draft**. In the spec's words, anything may change until v1. In
practice the project already keeps these promises for the catalogue
([CONTRIBUTING.md](../CONTRIBUTING.md), [GOVERNANCE.md](../GOVERNANCE.md)):

- Records are never deleted. A record is retired with `status`.
- A published ID is never reused for another exercise. A rename adds the old ID to
  `previous_ids`, and no ID is ever both current and another record's previous ID.
- Every release has a changelog entry naming what a consumer must act on.

Release numbers are `v0.N.P` until v1: `N` goes up for releases that add, rename, retire or
replace records or change the format; `P` for corrections that don't. So a `P` release changes no
IDs, statuses or format, and an `N` release is worth reading before you load it.

What can still change in v0, with a changelog entry when it does:

- **The format.** Fields and enum values can be added, and in v0 also renamed or removed. A
  renamed vocabulary term or field is called out as breaking. New vocabulary terms are additions.
- **Record content.** Corrections to any value, including `measurement`, `equipment` and alias
  kinds.
- **Review status** moves up as people and pipelines check records.

From v1 the format freezes: additions stay in the same major version, and removals or changes of
meaning take a new one, through a written proposal. The schema URL carries the major version
(`/schema/v0/`), so a v1 record names a different schema.

To be robust now: ignore `x-` members you don't know, tolerate unknown vocabulary terms, and
treat an absent optional field as "not stated" rather than as false.

## 11. Corrections and help

Found a wrong value? Fix it here, not in your copy: [CONTRIBUTING.md](../CONTRIBUTING.md) explains
how to correct, add and review records. An app with its own admin edits should send them back as
pull requests. Questions about the format go in an issue. Security problems go to the address in
[SECURITY.md](../SECURITY.md).
