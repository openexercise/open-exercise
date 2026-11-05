# Review lists

2026-11-05. Lists of records worth a person's look before anyone reviews the catalogue record by record.

**How they were made.** An AI assistant scanned all 815 records in `catalogue/exercises` with scripts (name and alias collisions after normalising words like single/one-arm and rear/reverse, `setup` facets the name doesn't fix, grip, stance and tempo words in instructions, required equipment) and read the wger entries behind each of the 106 merged groups in `REPORT.md` (`sources/wger/exercises.json`). It then judged each candidate. **No person has reviewed these lists.** Every row is a proposal: the maintainer ticks it, and the fix lands as a pull request that passes `oes validate` and follows [CONTRIBUTING.md](../CONTRIBUTING.md) (IDs never change, records are never deleted, so the fixes use `deprecated`, `replaced` + `replaced_by`, aliases and `previous_ids`).

**Decisions (2026-11-05).** The maintainer went through every list. Ticked rows are applied; the two unticked rows in section 3 are the bulk facet cleanup (its own pull request) and the bench rack, a format question for an issue. Where a row offered a choice:

- Section 1: "90/90 Hamstring Stretch" is dropped and the merge kept, with no split. `oec:close-grip-push-up`, `oec:dumbbell-side-lying-external-rotation` (wger 406 names the infraspinatus and teres minor, both external rotators, so "towards stomach" reads as a slip) and `oec:cable-lateral-raise` (wger 1752's "Front" twin of the behind-the-back raise is the standard cable lateral raise) keep their merges. "Single Arm Row" becomes `legacy`, not dropped.
- Section 2: meditation stays set aside; the catalogue stays physical exercise. Boxing, Radio Taiso and Lower-Body Stretching are written as new `machine-generated` records, as is `oec:machine-seated-dip`.
- Section 4: both single-arm cable lateral raises are replaced by `oec:cable-lateral-raise`, which stays unilateral (in practice the cable lateral raise is done one arm at a time). The low-confidence replacements were checked against the steps and applied. Replaced records keep their own `external_ids`, aliases and `previous_ids`, as the catalogue's earlier replaced records do, so a lookup by any of them finds the record and follows `replaced_by`; the survivor gains their names as aliases and their sources.
- Section 5: `by` is the latest reviewer's handle, and git history keeps the rest ([GUIDE.md](../GUIDE.md)). `review.scope` waits until an expert reviews. Set-aside records are `unreviewed`.
- Section 6: "Neck Extension" is dropped. Of the other misspelling aliases, 17 are real misspellings and stay; 7 become `common`, 2 `brand` (TRX, Hammer Strength) and 14 more `legacy`.

**What a fix does in Lifterclub.** Lifterclub moves stored references when an ID shows up in a record's `previous_ids`. A `deprecated` or `replaced` record is retired on load, but Lifterclub doesn't follow `replaced_by` today, so logs already on a replaced record stay on it. Merging those logs needs a Lifterclub change, not just a catalogue one. Moving a wger ID or legacy ID between records changes nothing for data Lifterclub has already loaded.

Confidence: **high** means the sources plainly say so, **med** means likely but worth a look, **low** means a judgement call.

## 1. Suspect merges

All 106 groups in `REPORT.md` were checked against their wger names and descriptions. 95 look right. These 11 are wrong or doubtful:

| ✓ | Record ID | What's wrong | Proposed fix | Confidence |
| --- | --- | --- | --- | --- |
| [x] | `oec:wall-slide` | "Serratus Wall Slide Stretch" (wger 2523) faces the wall with forearms on it and protracts; "Wall Slides" (716) has the back to the wall | New record `oec:serratus-wall-slide` "Serratus Wall Slide"; move wger 2523 and `oec:serratus-wall-slide-stretch` from `previous_ids` to it | high |
| [x] | `oec:cable-triceps-pushdown` | "Tricep Pushdown on Cable" (wger 805) is described as the rope push-down, and `oec:cable-rope-triceps-pushdown` exists | Move wger 805, the alias and `oec:tricep-pushdown-on-cable` to `oec:cable-rope-triceps-pushdown` | high |
| [x] | `oec:cable-single-arm-lat-pulldown` | The merge is right (wger 1637 is a seated single-arm pulldown despite its name), but the alias "Single Arm Row" (`misspelling`) would match one-arm row logs to a pulldown | Keep the merge; change that alias's kind to `legacy` (or drop it) | high |
| [x] | `oec:barbell-bench-press` | "Supine press" (wger 1583) is a close-grip press filed under Back; it is not the standard bench press | Move wger 1583, the alias and `oec:supine-press` to `oec:barbell-close-grip-bench-press` | med |
| [x] | `oec:cable-seated-row` | "Rowing seated, narrow grip" (wger 512) is the close-grip row, which has its own record | Move wger 512, the alias and `oec:rowing-seated-narrow-grip` to `oec:cable-close-grip-seated-row` | med |
| [x] | `oec:standing-elevated-hamstring-stretch` | "90/90 Hamstring / Barre-Style Stretch" (wger 2648, placeholder text): a 90/90 hamstring stretch is usually supine; the alias "90/90 Hamstring Stretch" will mislead | Keep the merge ("barre-style" fits); drop the alias "90/90 Hamstring Stretch", or split as `oec:supine-90-90-hamstring-stretch` if the maintainer reads it as supine | med |
| [x] | `oec:incline-dumbbell-y-raise` | Source 1825 describes a Y-raise, but the alias "Chest-Supported Rear Delt Raise" names what `oec:incline-dumbbell-reverse-fly` is | Keep the merge; change that alias's kind to `legacy` | med |
| [x] | `oec:single-leg-glute-bridge` | "Unilateral Hip Thrust" (wger 1913) is described lying down, so a bridge; the alias will catch single-leg hip thrust logs (`oec:dumbbell-single-leg-hip-thrust`) | Keep the merge; change the alias's kind to `legacy` | med |
| [x] | `oec:close-grip-push-up` | "Triceps push up" (wger 2684) keeps hands under the shoulders with elbows pinned; close-grip has hands almost touching | Keep (most people log them as one); or split as `oec:triceps-push-up` "Triceps Push-Up" | low |
| [x] | `oec:dumbbell-side-lying-external-rotation` | "Lying Rotator Cuff Exercise" (wger 406) says "rotate towards stomach", which is internal rotation, yet names the infraspinatus | Keep; a reviewer checks the source and moves it to `oec:dumbbell-side-lying-internal-rotation` if internal | low |
| [x] | `oec:cable-lateral-raise` | "Side lateral raise - Front (Cable)" (wger 1752) stands with the cable in front and targets the rear delt, nearer a cable Y-raise | Keep; or move to `oec:cable-cross-body-y-raise` | low |

## 2. Records set aside

The 16 records `REPORT.md` set aside are `deprecated` with no `provenance.review`. Six are real things people log:

| ✓ | Record ID | What's wrong | Proposed fix | Confidence |
| --- | --- | --- | --- | --- |
| [x] | `oec:boxing` | Heavy-bag boxing (wger 2542) is a real conditioning activity; `oec:straight-punch` covers one drill only | Reinstate (`active`), name "Boxing", summary as bag work; profile `duration` (optional `rounds` as reps, `heart-rate`, `calories`); load `none`; modalities `conditioning`, `sport` | high |
| [x] | `oec:bag-training` | Same activity as Boxing (wger 1092: bag work for cardio) | `replaced`, `replaced_by: oec:boxing`; add "Bag Training", "Heavy Bag" as aliases on Boxing | high |
| [x] | `oec:radio-taiso` | A real, widely done Japanese calisthenics routine, logged as one timed activity | Reinstate as "Radio Taiso" (aliases "Rajio Taisō", "Radio Calisthenics"), modality `calisthenics`/`mobility`, profile `duration` | med |
| [x] | `oec:leg-and-hip-stretch` | Vague, but a real thing people log: a general lower-body stretching session | Reinstate as "Lower-Body Stretching" (keep the ID), profile `duration`, `movement_patterns: [stretch]`; aliases "Leg and Hip Stretch", "Leg Stretches" | med |
| [x] | `oec:guided-or-free-meditation` | Meditation is a real practice logged by duration, next to `oec:diaphragmatic-breathing` | Reinstate as "Meditation", profile `duration`, modality `other`; or keep set aside if the catalogue stays physical-only (maintainer's call on scope) | low |
| [x] | `oec:machine-dip-or-triceps-extension` | An either/or, but the seated dip machine is a real exercise with no record (`oec:machine-assisted-dip` is the counterweight dip) | New record `oec:machine-seated-dip` "Machine Seated Dip"; mark this one `replaced` by it | med |
| [x] | `oec:limber-11` | A named 11-step routine, not an exercise; its parts exist (`foam-roller-it-band`, `foam-rolling-adductor`, `bent-knee-iron-cross`, `rollover-to-v-sit`…) | Keep set aside; it belongs in an app's routines | high |
| [x] | `oec:neutral-grip-pull-ups-or-trx-rows` | Either/or of `oec:neutral-grip-pull-up` and `oec:suspension-trainer-row` | Keep set aside (`replaced_by` takes one target) | high |
| [x] | `oec:lying-dumbbell-row-ss-seated-shrug` | A superset of `oec:incline-dumbbell-row` and `oec:dumbbell-shrug` | Keep set aside | high |
| [x] | `oec:row` | Generic "row" with no implement or position | Keep set aside; "Row" is too generic to alias anywhere | high |
| [x] | `oec:blaze` | A brand's HIIT class | Keep set aside | high |
| [x] | `oec:wall-drills` | Three linked knee drills with no text of their own | Keep set aside; add a record per drill only if someone asks | med |
| [x] | `oec:banded-shoulder-drills` | A video's sequence of band drills | Keep set aside | med |
| [x] | `oec:upper-external-oblique` | A muscle, not an exercise | Keep set aside | high |
| [x] | `oec:rest-for-timed-workouts` | Rest, not an exercise | Keep set aside | high |
| [x] | `oec:biceps` | Junk entry (wrong muscle, no description) | Keep set aside | high |

## 3. Over-specified details

The scan found `setup` facets the name doesn't fix in most records: `stance.width` in 389 active records, `grip.width` in 233, a non-neutral `grip.orientation` in 178, and `incline_degrees` in 10. SPEC §6 says to leave out what the exercise doesn't fix. The clearest cases:

| ✓ | Record ID | What's wrong | Proposed fix | Confidence |
| --- | --- | --- | --- | --- |
| [ ] | all active records (bulk) | `setup.stance.width`, `setup.grip.width` and `grip.orientation` filled in by the generator where the name fixes none of them | One scripted pull request: remove these facets unless the name or a sibling record depends on them (wide-grip, close-grip, sumo, narrow-stance…), then spot-check 20 | med |
| [x] | `oec:barbell-back-squat` | `grip.width: wide` and `load_position: high-bar`, while its aliases include both "high bar squat" and "low bar squat" | Remove `setup.grip` and `load_position` | high |
| [x] | `oec:smith-machine-squat` | `grip.width: wide`, `load_position: high-bar` | Remove `setup.grip` and `load_position` | high |
| [x] | `oec:pull-up` | `grip.width: wide`, though `oec:wide-grip-pull-up` exists | Remove `grip.width` (keep `pronated`) | high |
| [x] | `oec:cable-lat-pulldown` | `grip.width: wide`, and steps say "the wide lat pulldown bar", though `oec:wide-grip-lat-pulldown` exists | Remove `grip.width`; say "a lat pulldown bar" | high |
| [x] | `oec:barbell-deadlift` | `grip.orientation: pronated`; double overhand, mixed and hook grips are all the same lift | Remove `setup.grip` | high |
| [x] | `oec:barbell-bench-press` | `grip.width: wide`; standard bench is a moderate grip (instructions already say "slightly wider than shoulder-width") | Remove `grip.width` | med |
| [x] | `oec:assisted-pull-up`, `oec:machine-assisted-pull-up` | `grip.width: wide` and "wider than shoulder" on a generic pull-up | Remove `grip.width` | med |
| [x] | `oec:barbell-romanian-deadlift` | `grip.orientation: pronated`; grip doesn't define an RDL | Remove `setup.grip` | med |
| [x] | `oec:hack-squat` | `incline_degrees: 45`; machines differ | Remove `incline_degrees` | low |
| [x] | `oec:barbell-overhead-press` | Rack is `required`; the bar can be cleaned from the floor | Move `squat-rack`/`power-rack` to `optional` | med |
| [x] | `oec:barbell-forward-lunge`, `oec:barbell-reverse-lunge`, `oec:barbell-walking-lunge` | Rack is `required` | Move the rack to `optional` | med |
| [ ] | `oec:barbell-bench-press` and 8 other bench variants | Rack `required`, so a flat bench with built-in uprights doesn't count | Format issue: propose a vocab term (e.g. `bench-with-uprights`) as an alternative in the rack group, or make the rack optional | med |
| [x] | ~20 records (e.g. `oec:machine-standing-calf-raise`, `oec:levitation-crunch`, `oec:bilateral-band-external-rotation`) | Steps prescribe a 1–2 s hold at the top; that's tempo, not the exercise (SPEC §10) | Reword to "pause briefly" or drop; leave genuine pause variants (`barbell-pause-*`, `pause-hack-squat`, `bodyweight-tempo-squat`) as they are | low |

## 4. Near-duplicates

Active records that are the same exercise under two or three IDs. Found by alias and name collisions, then checked against summaries, `setup` and wger sources. The record that stays is the one with more sources, or the one following the implement-first naming rule. "Replace" means `status: replaced`, `replaced_by` the survivor, its names and wger IDs moved to the survivor as aliases, and relations pointing at it repointed.

| ✓ | Record ID | What's wrong | Proposed fix | Confidence |
| --- | --- | --- | --- | --- |
| [x] | `oec:standing-biceps-stretch` | wger's left and right versions (1232/1233) became two records, `oec:wall-biceps-stretch` the other side | Replace `oec:wall-biceps-stretch` → `oec:standing-biceps-stretch` | high |
| [x] | `oec:lateral-squat` | wger's left and right side split squats (986/987) became two records, `oec:lateral-split-squat` the other side | Replace `oec:lateral-split-squat` → `oec:lateral-squat`; keep `oec:cossack-squat` separate | high |
| [x] | `oec:dumbbell-biceps-curl` | `oec:dumbbell-curl` is the same standing supinating curl | Replace `oec:dumbbell-curl` → `oec:dumbbell-biceps-curl` | high |
| [x] | `oec:alternating-dumbbell-curl` | `oec:dumbbell-alternating-biceps-curl` has the same setup and steps | Replace `oec:dumbbell-alternating-biceps-curl` → `oec:alternating-dumbbell-curl` | high |
| [x] | `oec:cable-biceps-curl` | `oec:cable-curl` is the same low-pulley curl | Replace `oec:cable-curl` → `oec:cable-biceps-curl` | high |
| [x] | `oec:dumbbell-side-lying-external-rotation` | `oec:side-lying-dumbbell-external-rotation` and `oec:dumbbell-external-rotation` are both side-lying | Replace both → `oec:dumbbell-side-lying-external-rotation` | high |
| [x] | `oec:cable-external-rotation` | `oec:cable-shoulder-external-rotation` is identical | Replace `oec:cable-shoulder-external-rotation` → `oec:cable-external-rotation` | high |
| [x] | `oec:cable-reverse-fly` | `oec:cable-rear-delt-fly` and `oec:standing-cable-reverse-fly` are the same standing cable fly | Replace both → `oec:cable-reverse-fly` | high |
| [x] | `oec:dumbbell-seated-rear-delt-raise` | `oec:seated-bent-over-dumbbell-rear-delt-raise` and `oec:seated-dumbbell-bent-over-lateral-raise` are the same seated bent-over raise | Replace both → `oec:dumbbell-seated-rear-delt-raise`; keep standing `oec:dumbbell-bent-over-reverse-fly` | high |
| [x] | `oec:cable-single-arm-triceps-pushdown` | `oec:single-arm-cable-pushdown` is identical | Replace `oec:single-arm-cable-pushdown` → it | high |
| [x] | `oec:cable-single-arm-overhead-triceps-extension` | `oec:single-arm-cable-overhead-triceps-extension` is identical (and carries the alias "Neck Extension") | Replace it → `oec:cable-single-arm-overhead-triceps-extension`; don't carry "Neck Extension" over | high |
| [x] | `oec:cable-behind-the-back-lateral-raise` | `oec:behind-the-back-cable-lateral-raise` is identical | Replace it → `oec:cable-behind-the-back-lateral-raise` | high |
| [x] | `oec:cable-single-arm-lateral-raise` | `oec:single-arm-cable-lateral-raise` is identical; `oec:cable-lateral-raise` is also marked unilateral | Replace `oec:single-arm-cable-lateral-raise` → it; decide whether `oec:cable-lateral-raise` is the two-handle version (and fix its laterality) or replace it too | high |
| [x] | `oec:cable-half-kneeling-single-arm-lat-pulldown` | `oec:half-kneeling-single-arm-lat-pulldown` is identical | Replace `oec:half-kneeling-single-arm-lat-pulldown` → it | high |
| [x] | `oec:hack-squat` | `oec:machine-hack-squat` is the same machine squat ("Hackenschmidt machine") | Replace `oec:machine-hack-squat` → `oec:hack-squat`; keep `oec:barbell-hack-squat` | high |
| [x] | `oec:seated-dumbbell-shoulder-press` | `oec:seated-dumbbell-overhead-press` is the same | Replace `oec:seated-dumbbell-overhead-press` → it | high |
| [x] | `oec:smith-machine-shoulder-press` | `oec:smith-machine-overhead-press` is the same seated Smith press | Replace `oec:smith-machine-overhead-press` → it | high |
| [x] | `oec:chest-supported-dumbbell-row` | `oec:incline-dumbbell-row` is the same prone incline row | Replace `oec:incline-dumbbell-row` → it; keep `oec:dumbbell-helms-row` | high |
| [x] | `oec:prone-y-t-w-raise` | `oec:prone-ytw-raise` is identical | Replace `oec:prone-ytw-raise` → it | high |
| [x] | `oec:cat-cow` | `oec:cat-cow-stretch` is identical | Replace `oec:cat-cow-stretch` → it | high |
| [x] | `oec:bodyweight-forward-lunge` | `oec:forward-lunge` is the same bodyweight lunge | Replace `oec:forward-lunge` → it (and settle laterality: unilateral vs alternating) | high |
| [x] | `oec:cobra-pose` | `oec:cobra-stretch` is the same prone extension | Replace `oec:cobra-stretch` → it | med |
| [x] | `oec:butcher-s-block-stretch` | `oec:bench-prayer-stretch` is the same kneeling, elbows-on-bench stretch | Replace `oec:bench-prayer-stretch` → it | med |
| [x] | `oec:couch-stretch` | `oec:rear-foot-elevated-hip-flexor-stretch` is the same position | Replace it → `oec:couch-stretch` | med |
| [x] | `oec:seated-figure-four-stretch` | `oec:seated-piriformis-stretch` is the same crossed-leg seated stretch | Replace it → `oec:seated-figure-four-stretch` | med |
| [x] | `oec:supine-figure-four-stretch` | `oec:glute-stretch` (placeholder wger text) is a supine glute stretch, most likely the same | Replace `oec:glute-stretch` → it, after reading the steps | low |
| [x] | `oec:half-kneeling-cable-high-row` | `oec:single-arm-cable-high-row` is the same half-kneeling single-arm high-pulley row | Replace `oec:single-arm-cable-high-row` → it | med |
| [x] | `oec:reverse-grip-lat-pulldown` | `oec:cable-reverse-grip-lat-pulldown` has the same setup (shoulder-width supinated) | Replace it → `oec:reverse-grip-lat-pulldown`; keep `oec:close-grip-supinated-lat-pulldown` | med |
| [x] | `oec:dumbbell-standing-calf-raise` | `oec:dumbbell-calf-raise` is also standing with dumbbells at the sides | Replace `oec:dumbbell-calf-raise` → it | med |
| [x] | `oec:dumbbell-hex-press` | `oec:dumbbell-close-grip-bench-press` holds the dumbbells together, neutral, which is the hex/crush press | Replace `oec:dumbbell-close-grip-bench-press` → it, or keep if "close grip" means apart but narrow | med |
| [x] | `oec:weighted-crunch` | `oec:dumbbell-crunch` is a weighted crunch whose implement doesn't change the number | Replace `oec:dumbbell-crunch` → it | med |
| [x] | `oec:wall-angel` | `oec:wall-slide` (once the serratus slide is split out, section 1) is the same back-to-wall drill | Replace `oec:wall-slide` → it | med |
| [x] | `oec:running` | `oec:jogging` differs only in pace, which isn't an exercise facet (Zone 2 running was already merged in) | Replace `oec:jogging` → `oec:running` | med |
| [x] | `oec:machine-assisted-pull-up` | `oec:assisted-pull-up` has the same machine and setup; its source (1929) allows band or machine | Replace `oec:assisted-pull-up` → it | med |
| [x] | `oec:ez-bar-skull-crusher` | `oec:barbell-lying-triceps-extension` lists `ez-bar` as an alternative to the barbell, so the two overlap | Make the barbell record barbell-only (no replace) | med |
| [x] | `oec:pelvic-tilt` | Overlaps `oec:posterior-pelvic-tilt`; its steps describe a posterior tilt | Replace `oec:pelvic-tilt` → `oec:posterior-pelvic-tilt`, or rewrite it as the anterior/posterior rocking drill | low |
| [x] | `oec:dumbbell-lateral-and-front-raise` | Close to `oec:dumbbell-lateral-to-front-raise`, but one alternates raises and the other sweeps through horizontal adduction | Keep both; check the relations | low |
| [x] | `oec:leg-press` | `oec:narrow-stance-leg-press` and `oec:wide-stance-leg-press` differ only by stance | Keep: SPEC §3 makes stance a distinct record and `variation-of` already links them | low |

## 5. Review process

The schema already has what's needed: `provenance.review.status` (`unreviewed`, `machine-generated`, `community-reviewed`, `expert-reviewed`) with `at` and `by`, and `provenance.generated`, the JSON pointers no person has checked.

1. **Who.** Anyone may review a record through a pull request; a maintainer merges it. That makes it `community-reviewed`. `expert-reviewed` needs a reviewer the maintainers list in GOVERNANCE.md as qualified (for example, a certified coach or physiotherapist), reviewing in their field.
2. **What they check,** against a source they cite in `provenance.sources`:
   - **Names:** `name` follows SPEC §4; aliases are things people really call it; misleading wger names become `legacy`.
   - **Equipment:** `required` holds only what the exercise can't be done without.
   - **Muscles:** prime movers and the most specific term that's supported, never a term and its parent.
   - **Instructions and safety:** correct, safe, no sets, reps, loads or tempo.
   - **Measurement:** `load.kind` and `convention` say what a logged number means (per hand or total), and `reps_counted`.
   - **Setup:** only the facets the name fixes (section 3).
3. **How it's recorded.** The reviewer removes each pointer they checked from `provenance.generated`. When the list is empty, they set `review` to `{status: "community-reviewed", at, by}` (an expert sets `expert-reviewed`) and drop `model`. A partly checked record stays `machine-generated` with a shorter list, so field-level progress is kept.
4. **What Lifterclub does with it.** Lifterclub shows a record's `summary` and `instructions` only when its status is `community-reviewed` or `expert-reviewed`, and never shows text whose pointer is still in `generated` (SPEC §13: machine-generated safety text isn't authoritative). Names, equipment and measurement are used either way.
5. **Order.** Fix sections 1–4 first, then review records in the order Lifterclub users log them.

Proposals for the schema (each is a format change, so it goes through an issue):

- `by` holds one name, so a second reviewer overwrites the first. Either agree that `by` is the latest reviewer's GitHub handle and git history holds the rest, or add an optional `review.history` array of `{status, at, by}`.
- Nothing ties an `expert-reviewed` record to the reviewer's field. An optional `review.scope` (e.g. `["instructions", "anatomy"]`) would say what the expert checked.
- The 16 set-aside records have no `provenance.review`. Give them `unreviewed`, so every record states its review status.

## 6. Misleading aliases

The generator kept 45 wger names as `misspelling` aliases. Some aren't misspellings but different names, and a matcher would trust them. The schema has a `legacy` kind for these.

| ✓ | Record ID | What's wrong | Proposed fix | Confidence |
| --- | --- | --- | --- | --- |
| [x] | `oec:single-arm-cable-overhead-triceps-extension` | Alias "Neck Extension" (wger 1717's wrong name) | Drop, or `legacy` | high |
| [x] | `oec:cable-overhead-curl` | Alias "Overhand Cable Curl", which is a different curl | `legacy` | high |
| [x] | `oec:dumbbell-seated-external-rotation` | Alias "External Rotation Stretch" | `legacy` | high |
| [x] | `oec:plank-tuck-jump` | Alias "Dynamic Planche" | `legacy` | high |
| [x] | all records (bulk) | The other 41 `misspelling` aliases | Keep real misspellings ("Pendelay Row", "Facepull"); turn wrong names into `legacy` | med |
