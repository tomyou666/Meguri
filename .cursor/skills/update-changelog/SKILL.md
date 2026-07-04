---
name: update-changelog
description: Add an entry to CHANGELOG.md after implementing a feature, fix, or other notable change in this repo. Use when the user asks to update the changelog, or right after finishing a code change that should be logged.
---

# Update Changelog

Edit `CHANGELOG.md` at repo root after implementing a change.

## Steps

1. Open `CHANGELOG.md`.
2. Add the entry under `## [Unreleased]` (create this section right after the intro lines if missing).
3. Under `[Unreleased]`, use one of these exact Japanese category subheadings:
   - `### 追加` - new feature
   - `### 修正` - fix / change to existing behavior
   - `### その他` - other (docs, chore, refactor)
4. If the category subheading already exists under `[Unreleased]`, append a bullet to it instead of duplicating the heading.
5. Write one bullet per change, in Japanese, matching the terse style of existing entries (e.g. `〜を追加`, `〜を修正`), no trailing period.
6. For a sub-detail, nest a bullet with 2-space indent under the parent bullet.
7. Do NOT bump the version number or add a new `## [x.y.z] - date` heading unless the user explicitly asks to release/bump version.
8. Never edit already-released version sections.

## Format reference

```
## [Unreleased]

### 追加

- 新機能の説明

### 修正

- 修正内容の説明
```
