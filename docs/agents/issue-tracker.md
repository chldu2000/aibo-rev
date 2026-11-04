# Issue tracker: Local Markdown

Issues and specs live as Markdown files under `.scratch/`.
All paths below are relative to the repository root.

## Conventions

- One feature per directory: `.scratch/<feature-slug>/`.
- Spec: `.scratch/<feature-slug>/spec.md`.
- Implementation tickets: `.scratch/<feature-slug>/issues/<NN>-<slug>.md`,
  numbered from `01`, one file per ticket.
- Record triage state in a `Status:` line near the top, using
  [triage label mappings](triage-labels.md).
- Append comments and conversation history under `## Comments`.

## Publish to the issue tracker

Create the spec or individual ticket at the corresponding path above.
Create directories when needed.

## Fetch the relevant ticket

Read the referenced file. Resolve ticket numbers within their feature
directory; ask for the feature if the number is ambiguous.

## Wayfinding operations

Used by `/wayfinder`. A map has one child file per ticket.

- Map: `.scratch/<effort>/map.md`, containing Notes, Decisions-so-far,
  and Fog.
- Child: `.scratch/<effort>/issues/<NN>-<slug>.md`, numbered from `01`,
  with the question in the body.
- Type: record `research`, `prototype`, `grilling`, or `task` in a
  `Type:` line.
- Lifecycle: wayfinding tickets use `Status: open`, `Status: claimed`,
  or `Status: resolved`.
- Blocking: record `Blocked by: NN, NN` near the top. A ticket is
  unblocked when every listed ticket is resolved.
- Frontier: select the first numbered open, unblocked ticket.
- Claim: set `Status: claimed` and save before starting work.
- Resolve: append the answer under `## Answer`, set `Status: resolved`,
  and append a gist plus link to the map's Decisions-so-far section.
