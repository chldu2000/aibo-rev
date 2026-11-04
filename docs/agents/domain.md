# Domain docs

## Layout

This repository uses a single context:

- Root glossary: `CONTEXT.md`.
- Architectural decisions: `docs/adr/`.

## Before exploring

Read [CONTEXT.md](../../CONTEXT.md) and the
[ADRs](../adr/) relevant to the work.

Use the [README](../../README.md) to locate the current direction
and distinguish current scope from historical proposals. Check ADR
status and supersession metadata before applying a decision;
superseded decisions provide historical context.

If domain files are absent, proceed silently. The `/domain-modeling`
skill creates them lazily when terms or decisions are resolved.

## Vocabulary

Use glossary terms in issue titles, proposals, hypotheses, and tests.
Respect explicitly discouraged synonyms. If a needed concept is
missing, reconsider the wording or note the gap for `/domain-modeling`.

## ADR conflicts

Explicitly identify any conflict between a proposal and a current ADR,
including the decision and the reason for reconsidering it.
