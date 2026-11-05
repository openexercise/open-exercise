# Open Exercise Schema v0

**Status: draft.** Anything may change until v1. The key words MUST, SHOULD and MAY are used as in RFC 2119.

The Open Exercise Schema (OES) describes one exercise so any tool can identify it, search it, explain it, and record it the same way as any other tool. The Open Exercise Catalogue (OEC) is a public set of records in this format.

## 1. Why this exists

Exercise data doesn't travel. Every app has its own list and its own names, and the most damaging gap is in what a logged number means: "50 kg" on a dumbbell press may be one hand or both. OES fixes three things:

1. **Identity.** A stable ID that survives renames and merges.
2. **Meaning of a record.** What must be logged, and what the numbers mean.
3. **Honest origin.** Who or what produced each value, and whether a person has checked it.

## 2. Documents

An exercise is one JSON object, UTF-8, valid against `schema/v0/exercise.json`. Only `id`, `name` and `measurement` are required.

| Conformance | A record has |
| --- | --- |
| **Minimal** | `id`, `name`, `measurement` |
| **Standard** | Minimal, plus `classification`, `anatomy.muscles`, `equipment`, `provenance.license` |
| **Complete** | Standard, plus `summary`, `anatomy.joints`, `instructions.steps`, `provenance.sources`, `provenance.review` |

Consumers MUST ignore members starting `x-` that they don't understand. Producers MAY add `x-<name>` members at the top level of a record. Values in vocabularies MAY be `x-<term>`.

## 3. Identity

```
id = namespace ":" slug
```

- `oec:` is reserved for the catalogue. Anyone else uses their own namespace (`acme:my-press`). Private, per-user exercises SHOULD use `usr:`, which is meaningful only inside the system that made it.
- From v1, a published ID never changes meaning and is never reused.
- `previous_ids` lists IDs that once meant this exercise. A consumer holding an ID MUST look there if it finds no record with that ID.
- `status` is `active` (default), `deprecated` (keep resolving; don't offer for new logs) or `replaced` (`replaced_by` is required). Records are retired, never deleted.

An exercise is a distinct record when it differs in any way that changes how it's trained or compared: implement, stance, grip, angle, unilateral or not. If a person wouldn't compare the numbers directly, they're different exercises. Link them with `relations`.

## 4. Names

`name` is the canonical English name: title case, singular, implement first when it defines the exercise. `aliases` hold everything else people call it. `kind` tells a matcher how far to trust it. `translations` holds other languages, keyed by BCP 47 tag.

## 5. Classification

`classification` answers what kind of exercise this is. `modalities` and `movement_patterns` use the vocabularies in `vocab/v0`, most important first.

| Field | Values | Meaning |
| --- | --- | --- |
| `mechanics` | compound, isolation, not-applicable | Compound moves two or more joints under the working effort |
| `force` | push, pull, static, mixed, not-applicable | Direction of the working effort |
| `kinetic_chain` | open, closed, mixed, not-applicable | Closed: the far end of the limb is fixed |
| `laterality` | bilateral, unilateral, alternating, contralateral, not-applicable | How the sides work |
| `planes` | sagittal, frontal, transverse | Planes of the main motion |
| `contraction` | dynamic, isometric, eccentric-emphasis, ballistic | Character of the effort |
| `level` | beginner, intermediate, advanced | Technical demand, not how heavy it can be |

## 6. Setup

`setup` holds the facets that tell a variation from its siblings: `position`, `incline_degrees`, `grip`, `stance`, `load_position`, `range`. Leave out what the exercise doesn't fix. A name fixes the grip its exercise is conventionally done with, so a curl's grip is `supinated` and a bench press's `pronated`; a variant with another grip is a record of its own ("Hammer Curl", "Reverse-Grip Bench Press"). A name fixes no grip or stance width unless it says one ("Wide-Grip Pull-Up") or the exercise is defined by it ("Diamond Push-Up").

## 7. Anatomy

`anatomy.muscles` lists muscles with a `role`:

- `prime-mover` does most of the work.
- `synergist` assists.
- `stabilizer` holds position.
- `stretched` is the target of a stretch.

Terms come from the muscle tree in `vocab/v0/muscles.json`: region, group, muscle, head. A term implies its ancestors, so a record lists the most specific term it can support and never both a term and its parent. To answer "does this work the chest?", walk up from each term. Muscle percentages aren't in v0: published activation figures vary too much by study and method to state as a property of an exercise.

`anatomy.joints` lists each joint and its actions during the working phase.

## 8. Equipment

`equipment.required` is a list of groups. All groups are needed; any one item of a group is enough.

```json
{ "required": [["barbell"], ["squat-rack", "power-rack"]], "optional": ["strap"] }
```

An empty `required` means no equipment. Absent means unknown. A parent term in the vocabulary means any of its children. One item may fill more than one group: a bench press lists `flat-bench-with-uprights` in both its bench group and its rack group, because that one station is both.

## 9. Measurement

This section is what makes logs comparable. Metric units for exchange are fixed: kilograms, metres, seconds, as listed in `vocab/v0/metrics.json`. Tools MAY show anything they like.

`measurement.profiles` lists ways to record a set, the usual one first. A set matches a profile when it has every `required` metric.

### 9.1 What the load number means

`measurement.load` is REQUIRED when a profile records `load`.

| `kind` | The logged number is | Example |
| --- | --- | --- |
| `external` | The weight | Bench press |
| `bodyweight` | Nothing is logged | Push-up |
| `bodyweight-plus-external` | The weight added | Weighted pull-up |
| `bodyweight-minus-assistance` | The assistance | Assisted dip |
| `elastic` | A tension; varies through the rep | Band pull-apart |
| `none` | Nothing is loaded | Running |

| `convention` | The number counts | Example |
| --- | --- | --- |
| `total` | Everything moved, bar included | 100 kg barbell squat |
| `per-implement` | One of the `implements` held | 30 kg dumbbells: 60 kg moved |
| `per-side` | One side of a plate-loaded machine or sled | Plate-loaded row |
| `weight-stack` | The number on the pin; not comparable between machines | Cable row |

To compute moved load: `total` is the number; `per-implement` multiplies by `implements`; `per-side` multiplies by 2; `weight-stack` MUST NOT be compared across machines.

### 9.2 Counting

`reps_counted` and `duration_counted` say whether a logged value is `total` or `per-side`. Ten single-arm rows `per-side` means ten for each arm.

### 9.3 Computed numbers

`records.e1rm` says an estimated one-rep max means something here; `records.volume` says load times reps is a fair measure of work. If false or absent, tools SHOULD NOT present them. `bodyweight_fraction` and `energy.met` MUST cite a source, and are never to be estimated by a generator.

## 10. Instructions

`setup`, `steps`, `breathing`, `cues`, `common_faults` and `safety`. Instructions are general education, not medical advice. They SHOULD NOT prescribe sets, reps or loads, which belong to a program.

## 11. Relations

Read each as "this exercise *type* target".

| `type` | Meaning |
| --- | --- |
| `variation-of` | The same movement with one facet changed |
| `regression-of` | Easier than the target, and leads to it |
| `progression-of` | Harder than the target, and builds on it |
| `substitute-for` | A sensible swap when the target can't be done |
| `antagonist-of` | Works the opposing action |
| `part-of` | A component of the target |
| `same-as` | The same exercise in another namespace |

Relations aren't assumed to be symmetric, except `same-as` and `antagonist-of`.

## 12. External IDs

`external_ids` maps a system to its identifier. Registered systems: `wger`, `free-exercise-db`. Multiple IDs in one system are comma-separated. Use `x-<name>` for others; propose registration by pull request.

## 13. Provenance

Every value is only as good as where it came from, so a consumer must be able to tell.

- `license` is the SPDX ID of the record. `sources` credit what it derives from.
- `review.status` is, from least to most checked, `unreviewed`, `machine-generated`, `machine-verified`, `community-reviewed` or `expert-reviewed`.
- `machine-verified` means an automated pipeline checked the record against its sources with at least two independent models from different providers, they agreed nothing is left wrong, and the record validates. No person has checked it. `review.by` names the pipeline and its version, `review.at` the day, and `review.models` the models that checked it. It never replaces a person's review.
- `generated` lists JSON pointers to values a machine produced and no person has checked. A person who checks a value removes its pointer. A `machine-verified` record keeps its list; a record a person reviewed SHOULD have an empty one.

Consumers MAY present a `machine-verified` record's text, marked as checked automatically. Consumers SHOULD NOT present machine-generated or machine-verified safety text as authoritative.

## 14. Vocabularies

Each file in `vocab/v0` has terms with `id`, `label`, and optional `parent` and `definition`. Term IDs are stable from v1. New terms are additions, not breaking changes, so consumers MUST tolerate terms they don't know.

## 15. Versioning

The schema URL carries the major version. In v0 anything may change. From v1, additions stay in the same major version; removals or changes of meaning take a new one. A document SHOULD name its schema in `$schema`.

## 16. Relationship to other standards

The [Open Workout Schema](https://openworkout.dev) records what someone did; its entries refer to exercises by OES ID. OES says what those exercises are.

## 17. Licences

The schema, spec, vocabularies and tools are Apache-2.0. The catalogue is CC-BY-SA 4.0, which the wger data it derives from requires; records built on it carry the same terms and must credit the sources they list.
