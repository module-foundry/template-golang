---
name: commit
description: Use ONLY when the user asks to commit or an explicit commit point is reached in Russian or English ("закоммить", "commit", "сделай коммит", "зафиксируй"). Creates a Conventional Commits message and runs the Go commitlint; never commits unprompted.
---

# Commit

## When

- The user asked for a commit, or the agent reached an explicit commit point
  (a completed logical step). Never commit silently.

## Format (Conventional Commits)

```
<type>(<scope>): <subject>

[body]

[footer]
```

- Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`,
  `ci`, `chore`, `revert`.
- Scope is the module/package: `auth`, `apperror`, `openapi`, `docs`, `ci`.
- Subject is imperative, lowercase, no trailing dot, <= 72 chars total header.
- Body explains **why**, wrapped at 100 chars.
- Breaking public API change: `!` after scope and a `BREAKING CHANGE:` footer.

## Workflow

1. Inspect changes: `git status`, `git diff --staged` (stage only intended
   files).
2. Compose the message.
3. Validate: `task commit:lint` (or pipe the message into
   `commitlint lint`).
4. Commit; do not amend, force-push or skip hooks.

## Examples

```
feat(auth): add telegram mini-app token endpoint
fix(apperror): truncate messages to 120 runes
docs(skills): describe the commit workflow
feat(api)!: rename token response field
```
