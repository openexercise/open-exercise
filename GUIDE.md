# Using the catalogue

For apps that store, match or show Open Exercise IDs. The [specification](spec/v0/SPEC.md) says what
a record means; this says how to use the catalogue well.

## Pin a release

Take the catalogue from a tagged release (`v0.1.0`), not from `main`: a release has a
[changelog](CHANGELOG.md) entry saying what changed, so you can review the difference before you load
it. Each release on GitHub carries `catalogue.json`, every record in one file. Store the tag you
loaded with your data.

A release can rename, retire or replace records, never delete them, and never reuse an ID. Loading a
newer release is safe if you follow the next two sections.

## Resolve an ID you stored

IDs are `namespace:slug`. `oec:` is the catalogue's; `usr:` is a person's own exercise in your app
and means nothing elsewhere. To find the record for an `oec:` ID you stored:

1. Look for a record with that `id`.
2. If there's none, look for the record whose `previous_ids` lists it. That's a rename: move your
   stored references to the new ID (every one, including IDs inside JSON you keep), so you don't
   look it up again.
3. If neither finds it, keep the ID and show the name you stored with it. Don't drop the log.

No ID is ever both a current `id` and another record's previous ID, and no previous ID belongs to two
records (`oes validate` checks both), so the answer is never ambiguous.

## Deprecated and replaced records

- **`deprecated`**: keep resolving it and showing it in old logs, but don't offer it when someone
  picks an exercise for a new log, and don't match new text to it.
- **`replaced`**: the same, and `replaced_by` names the record to use instead. Offer to move a
  person's logs to it, or move them yourself when the two are the same exercise under two records
  (a duplicate merged later). Follow `replaced_by` to the end of the chain; a chain never loops.

## Names and aliases

`name` is the canonical English name. `aliases` are everything else people call the exercise, each
with a `kind` saying how far to trust it for matching:

| Kind | What it is | Trust it for |
| --- | --- | --- |
| `common` | Another name people use ("Bench Press" for Barbell Bench Press) | Matching, search and display |
| `abbreviation` | A short form ("RDL", "OHP") | Matching and search; don't display it as the name |
| `regional` | A name used in one region or community | Matching and search |
| `brand` | A machine or product name ("Smith Machine", "Pec Deck") | Matching and search |
| `misspelling` | A spelling people type ("pullup" is common; this is for real errors) | Matching what people type; never show it |
| `legacy` | A name that once meant this record, kept so old data still resolves, but that may mislead | Resolving old imports only: ask before matching new text to it |

When one name is an alias of more than one record, it doesn't decide anything: show the candidates
and let the person choose.

## Normalize before comparing

Compare a person's text with names and aliases after normalizing both, the same way every time:

1. Lowercase.
2. Replace each run of these characters with one space: ASCII control characters, punctuation and
   symbols (U+0001–U+002F, U+003A–U+0040, U+005B–U+0060, U+007B–U+007F, which include the space
   and tab), the no-break space (U+00A0), the middle dot (U+00B7), the multiplication sign
   (U+00D7), dashes (U+2010–U+2015), curly quotes (U+2018–U+201F) and the ellipsis (U+2026).
3. Trim spaces from both ends.

Letters and digits of every script stay, so "Développé couché" and "жим лежа" can match.
"Pull-Up", "pull up" and "PULL UP!" all become `pull up`.

Plurals and abbreviations are a matcher's choice, not part of the rule. The worked example below
handles both.

## A worked example: Lifterclub's matching

[Lifterclub](https://lifterclub.com/developers) matches names to catalogue IDs for apps
(`POST /v1/match`) and for what people log. For a name such as "DB bench":

1. **Normalize** it: `db bench`.
2. **Expand abbreviations** it knows: `db` → `dumbbell`, `bb` → `barbell`, `kb` → `kettlebell`,
   `ez` → `ez bar` or `sz bar`. The phrase is now tried as `db bench` and `dumbbell bench`.
3. **Exact match:** normalized names and aliases equal to any form, first as written, then with
   each word's plural `s` dropped ("kb swings" → "kb swing"), then with one added. One record found
   is a sure match (`catalogue`). Several are a choice to offer (`close`, with candidates).
   A generic name ("row", "press", "curl") always offers a choice, even when it hits exactly.
4. **Similarity:** otherwise, trigram similarity against names and aliases. At 0.85 or above it's
   sure; from 0.60 it's `close`, to confirm with the person; below that, `unmatched`, with the
   candidates as guesses.

Every result carries its candidates, best first. Records that are `deprecated` or `replaced` never
match. Today Lifterclub trusts every alias kind the same; following the table above would stop
`legacy` and `misspelling` aliases from deciding a match on their own.

## Show what's trustworthy

`provenance.review.status` says who has checked a record:

| Status | Meaning |
| --- | --- |
| `machine-generated` | A model wrote it from the sources; `provenance.generated` lists the values no person has checked |
| `community-reviewed` | A person checked every value (`by`, `at`) |
| `expert-reviewed` | A qualified person checked it, such as a coach or physiotherapist |
| `unreviewed` | Neither generated nor checked, such as a record set aside |

Until a record is reviewed, use its IDs, names, aliases and measurement, but don't present its
instructions, safety text or summary as authoritative. Once reviewed, show them with the credit the
licence asks for: the record's `provenance.sources` (title, author, licence), and "Open Exercise
Catalogue, CC-BY-SA 4.0".

## Send corrections back

Fix a record here rather than patching your copy: open a pull request that passes
`oes validate` and follows [CONTRIBUTING.md](CONTRIBUTING.md). An app that edits records in its own
admin tools (Lifterclub does) should export those edits as pull requests.
