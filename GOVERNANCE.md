# Governance

Open Exercise covers the Open Exercise Schema, its vocabularies, the `oes` tools and the Open Exercise Catalogue. It's a small project, run in the open.

## Maintainers

| Name | GitHub | Role |
| --- | --- | --- |
| Bryan Morrow | [myteedev](https://github.com/myteedev) | Lead maintainer |

Write to the maintainers at maintainers@openexercise.org, or open an issue. Someone who has made sustained, careful contributions can be invited to become a maintainer by the existing ones.

## How decisions are made

- **Records** (`catalogue/`): a pull request that follows [CONTRIBUTING.md](CONTRIBUTING.md) and passes `oes validate` is merged by a maintainer. Published IDs never change and records are never deleted.
- **The format** (`spec/`, `schema/`, `vocab/`): a change starts as an issue with real records it would improve, and a maintainer decides after at least a week for comment. New vocabulary terms are additions; renaming a term or field is breaking.
- **Versions:** v0 is a development draft, so breaking changes can still land. v1 freezes the format; after it, breaking changes need a new major version.

Maintainers aim for agreement in the issue. When there isn't any, the lead maintainer decides and says why in the issue.

## Where the source of truth is

This repository is the source of truth for the schema and the catalogue. Lifterclub, the first app to adopt them, takes a snapshot and sends its corrections back here as pull requests.

## Conduct and security

Everyone taking part follows the [code of conduct](CODE_OF_CONDUCT.md). Report security problems privately, as [SECURITY.md](SECURITY.md) describes.
