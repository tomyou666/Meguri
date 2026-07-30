---
name: plan-skill-annotate
description: >-
  Scans .cursor/skills/ and annotates a plan-mode markdown with applicable
  "Use the `skill-name` skill" notes on matching steps. Use only when explicitly
  invoked with a Cursor plan MD (e.g. .cursor/plans/*.plan.md).
disable-model-invocation: true
---

# Plan Skill Annotate

When explicitly invoked with a plan-mode markdown, check project skills and annotate the plan where they apply.

## When to run

- User attaches or names this skill **and** a plan MD (typically under `.cursor/plans/`).
- Do **not** auto-run from ambient context.

## Workflow

1. Read the attached / named plan MD.
2. List skills under `.cursor/skills/*/SKILL.md` (project skills only).
3. For each skill, read frontmatter `name` + `description` (and skim body only if the match is unclear).
4. Match skills to plan steps by task type (language, layer, file kind, changelog, tests, i18n, Wire, etc.).
5. Edit the plan MD: append a skill note on each matching step.
6. Reply briefly: which skills were added, and which steps got them. If none match, say so and do not invent notes.

## Annotation format

Per matching implementation step, add (or merge into) a bullet:

```markdown
- **Skill:** Use the `skill-name` skill.
```

Optional short reason in parentheses when helpful:

```markdown
- **Skill:** Use the `go-docstring-style` skill. (new/edited Go funcs, methods, struct fields)
```

Rules for the note:

- English only.
- Use the skill's `name` field exactly (backticks).
- Prefer one line per step; multiple skills on one step → same bullet, joined with "; " or a second **Skill:** bullet.
- Place the note at the **top of that step's bullet list** (before file/task bullets).
- Do not duplicate if the step already says to use that skill (any wording).

## Matching rules

- Annotate only when the step clearly triggers the skill's description.
- Skip this skill (`plan-skill-annotate`) itself.
- Skip skills that do not apply to any step.
- Do not recommend personal (`~/.cursor/skills/`) or built-in (`~/.cursor/skills-cursor/`) skills unless the user asks.
- Prefer existing project skill names over paraphrased advice.

## Examples (typical matches in this repo)

| Plan work | Skill |
|-----------|--------|
| Go funcs / methods / struct fields | `go-docstring-style` |
| Wire / providers / wire_gen / composition root | `go-wire` |
| TSX UI copy / labels / aria | `tsx-i18n-messages` |
| Adding or editing tests | `test-overview-style` |
| Notable feature/fix after impl | `update-changelog` |
| DESIGN.md → shadcn CSS tokens | `design-to-shadcn-css` |

Other `.cursor/skills/*` entries: match from their `description` the same way.

## Do not

- Rewrite the plan's technical content unrelated to skill notes.
- Add a global "use all skills" dump.
- Mark todos completed or change plan frontmatter status unless asked.
