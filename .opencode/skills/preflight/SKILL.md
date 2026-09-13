---
name: preflight
description: Use before saying "done", "готово", "проверь" or before a release in Russian or English ("готово", "проверь", "перед продом", "done", "check", "review"). Runs the definition-of-done gate and blocks the task until everything is green.
---

# Preflight (definition of done)

Run in order and do not declare the task complete until every item passes.

## 1. Gates

```bash
task verify      # fmt + vet + tests + coverage gate pkg/* + build
task api:check   # routes snapshot + OpenAPI golden (contract not changed)
```

## 2. Checklist

- [ ] Tests cover new/changed behaviour; `pkg/*` coverage gate >= 90%.
- [ ] Hot path changes have benchmarks with before/after numbers.
- [ ] `docs/<topic>/{en,ru}.md` updated in both languages; README only if
      requirements/quick start changed.
- [ ] `.env.example` updated for every new variable.
- [ ] Error codes reused where possible; new codes added to the registry and
      `docs/error-codes/{en,ru}.md`.
- [ ] Migrations have `Down`, no destructive change without a backup note.
- [ ] Public API diff reviewed: breaking changes marked `BREAKING CHANGE`.
- [ ] `.gitignore`/`README.md`/`LICENSE` still correct.
- [ ] No secrets, no local utils, no new unjustified dependencies.

## 3. Report

Summarize: what changed, commands run, results, remaining risks. If something
could not be verified locally (e.g. Docker unavailable), say so explicitly.
