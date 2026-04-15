# Velox Development Rules & Conventions

> Living document — split per platform. For project overview and build commands, see [CLAUDE.md](../CLAUDE.md).

## Index

- [Backend Rules (Go)](development-rules-backend.md) — Handler/Service/Repository/Model, SQLite, error handling, testing.
- [Webapp Rules (React / TypeScript)](development-rules-webapp.md) — React 19 + Compiler, TailwindCSS 4, React Query, Zustand.
- [Mobile Rules (Android / Kotlin)](development-rules-mobile.md) — Jetpack Compose + Hilt + Media3, Clean Architecture + MVVM, coroutines/Flow.

---

## Shared Conventions

### Commit Convention
```
Add(scope): new feature
Fix(scope): bug fix
Enhance(scope): improve existing feature
Refactor(scope): structural change, no behavior change
Chore: tooling, deps, config
```

### Language
- Code comments and variable names in **English**.
- Plan/spec files may contain **Vietnamese**.
- Commit messages in **English**.

### No Premature Abstractions
- Don't create helpers/wrappers for one-time operations.
- Three similar lines > one premature abstraction.
- Only abstract when the pattern repeats 3+ times.

### Pre-commit Hooks (Husky + lint-staged)
- `.ts/.tsx` → Prettier
- `.go` → gofmt
- `.kt` → ktlint (if configured)
- Runs automatically on `git commit`.
