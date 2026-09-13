---
name: add-skill
description: Use when adding or updating an AI skill or the routing rules in Russian or English ("добавь скилл", "обнови правила AI", "add skill", "skill"). Registers new skills in opencode.json, AGENTS.md routing and docs/skills/{en,ru}.md.
---

# Add or update a skill

## Steps

1. Create `.opencode/skills/<name>/SKILL.md` with frontmatter:

```markdown
---
name: <name>
description: Use when <what it does> in Russian or English ("триггер 1", "trigger 1"). <one-line scope>.
---
```

- `name` matches the folder, lowercase with hyphens.
- `description` front-loads literal trigger words (RU + EN); include "Use ONLY
  when ..." if the skill must stay quiet on adjacent topics.

2. Body: self-contained procedure — steps, commands, templates, settings and a
   local definition of done. No rules duplicated from `instructions`.
3. Add a row to the routing table in `AGENTS.md` (RU + EN triggers).
4. Update `docs/skills/{en,ru}.md` (both languages).
5. Register the path only if it lives outside `.opencode/skills`:
   `opencode.json` -> `skills.paths`.
6. Check trigger collisions with existing skills; rename or narrow if needed.
7. Remind the user to restart opencode: config and skills are loaded at startup.

## Rules

- One skill = one responsibility, one clear trigger set.
- Skills may reference each other (e.g. `add-module` -> `tdd-tests`) but must not
  require loading all of them.
- Keep skills executable: exact file paths, exact commands.
