---
PLAN: "refactor!: cash_manager on the crud-by-declaration framework — Allow and Rules in definition.go, persistence by crud, only domain rules in rule files"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 10543397699119517187
PR: https://github.com/veltylabs/cash_manager/pull/9
REVIEWED_COMMIT: 140187f8a44351b58e4071c31b58a8b98a6ea6c9
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work. Commit after every stage.
> Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work.
> **Push early:** after EVERY stage run `git push origin HEAD`; after Stage 3 open the pull request and keep pushing to it.
> **Never ask a yes/no question about progress** («am I on the right track?», «shall I proceed?», «shall I push?»): the answer is always yes. Ask only when a compile error or a signature in this plan is wrong, and name the symbol.
> Never edit `definition_gen.go`. Never write `go.mod`/`go.sum` lines by hand: only `go get` and `go mod tidy` do. No scratch files and no sed/regex/python scripts over the source: edit by hand, file by file, and run `gotest` after each file.

# Plan: `cash_manager` on the crud-by-declaration framework

Two earlier sessions of this plan stalled on questions it did not answer. This version is split into stages, one file each; **every answer they needed is written inside the stage that needs it**: read the stage, do it, `gotest`, commit, push, next.

Today `ops.go` mounts six operations by hand (open and close a shift, today's shift, stored facts, and the reports `reorder` and `thirty_day`); `StoredFact` is written only by `Accept(money.Fact)`. After: two tables (`Shift` with rules, `StoredFact` read-only), three flat read models, and `Accept` over the store.

Reference module, already migrated (copy its shape): https://github.com/veltylabs/patient_directory/blob/main — `definition.go`, `rule_patient_create.go`, `rule_patient_update.go`, `store.go`, `import.go`, `seed/demo.go`, `tests/setup_test.go`, `tests/crud_test.go`, `tests/demo_test.go`.

## Stages

| # | File | What | Files touched |
|---|---|---|---|
| 1 | [PLAN_STAGE_1_TOOLS_DEFINITION.md](PLAN_STAGE_1_TOOLS_DEFINITION.md) | tools, `definition.go`, `errors.go` | `definition.go`, `errors.go`, `go.mod`, `go.sum` |
| 2 | [PLAN_STAGE_2_MODULE_STORE.md](PLAN_STAGE_2_MODULE_STORE.md) | `Module.store`, helpers, `Accept` | `module.go`, `store.go`, `rows.go` |
| 3 | [PLAN_STAGE_3_RULES.md](PLAN_STAGE_3_RULES.md) | generate; shift rules; the three read models | `rule_*.go` |
| 4 | [PLAN_STAGE_4_BROWSER.md](PLAN_STAGE_4_BROWSER.md) | delete `ops.go`; `view.go`; demo | `ops.go` (deleted), `view.go`, `seed/demo.go`, `web/` (deleted) |
| 5 | [PLAN_STAGE_5_TESTS_DOCS.md](PLAN_STAGE_5_TESTS_DOCS.md) | tests, `docs/TESTS.md`, README | `tests/*.go`, `docs/TESTS.md`, `README.md`, `docs/ARCHITECTURE.md` |

## Operation mapping (every operation of today has exactly one row)

`<entity>.<verb>` is `router.Operation{Module: ModelName, Entity: "<entity>", Action: model.<Verb>}`; `ModelName` is the constant `definitionc` writes (`"cash_manager"`). The old module name `"cashmanager"` disappears: use the constant, never a literal.

| Today | After |
|---|---|
| `OpenShift` | `cash_shift.create` with `{opening_cash, currency}` (the cashier is the caller) |
| `CloseShift` | `cash_shift.update` with `Fields: ["closing_cash_counted"]` |
| `GetToday` | read model `cash_today` |
| `ListStoredFacts` | `stored_fact.read` |
| `Reorder` | read model `reorder` |
| `ThirtyDay` | read model `thirty_day` with filter `{from, to, scope}` |

## Delivery check

1. `gotest` ends green on EVERY destination (`wasm` and `browser` included) and «tests index» ✅.
2. `langc sync .` reports zero untranslated keys for `es` and zero reuse warnings.
3. `grep -rn "reg.Operation\|router.Context\|rpc.WriteError\|m\.db\." --include=*.go . | grep -v "_test.go\|^./ui/\|^./definition_gen.go"` → only `m.db` inside `New` and `store.go`.
4. `ls *.py *.sh commit_message.txt 2>/dev/null` → empty; `grep -c XXXX go.sum` → 0.
5. `git diff --stat main` lists only the files of the stages table plus badge files `gotest` rewrites.
