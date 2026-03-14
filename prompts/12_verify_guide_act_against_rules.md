# Prompt 12: Verify Guide Act Against Editorial Rules

```text
I want you to verify one specific act of the campaign guide against the editorial rules we established while refining Act 1, and then apply the necessary edits directly in that act.

Context files to use as source of truth:
- guide source: `guides/stormburst_campaign.md`
- editorial rules: `docs/zasady_redakcji_poradnika.md`

Act to verify:
- Act [PASTE_ACT_NUMBER_HERE]

Your task:
1. Read only the relevant act from the guide source and the editorial rules document.
2. Audit the act step by step.
3. Check whether every step is consistent with the rules.
4. Apply the necessary edits directly in the analyzed act so it complies with the rules.
4. Focus especially on the issues that mattered during Act 1 refinement:
   - each numbered step should describe one real action,
   - important area-entry steps should stay visible as separate route milestones,
   - steps should be procedural and executable when they appear,
   - no vague wording like `reward gemowy`, `po odblokowaniu`, `wróć do wejścia`,
   - reward and vendor steps should name the exact NPC whenever possible,
   - route logistics should explicitly say whether to return by portal, logout, or waypoint when that matters,
   - steps that are really one continuous action should be merged,
   - standalone gear reminders or vague tips should be removed or merged into a neighboring step,
   - optionality should be used only when it reflects a real branch or decision.

Execution requirements:
- Edit only the requested act in `guides/stormburst_campaign.md`.
- Do not rewrite other acts.
- Preserve existing formatting, colors, and numbering style used by the guide.
- Prefer the smallest set of edits that brings the act in line with the rules.
- If two steps should be merged, actually merge them in the file.
- If a step should be removed, actually remove it and renumber the remaining steps in that act.

Output requirements:
- Start with findings only.
- Order findings by severity.
- For each finding, include:
  - step number,
  - short explanation of why it violates the rules,
  - the final rewrite applied in Polish.
- If a step was removed, say that explicitly.
- If two or more steps were merged, say which ones and show the final merged wording.
- If no problems are found, say that explicitly and do not make unnecessary edits.

Constraints:
- Do not rewrite the entire act unless it is necessary to satisfy the rules.
- Do not analyze other acts.
- Base your review on the rules document, not on personal preference.
- If some detail is uncertain and cannot be verified from the repository context, call that out explicitly instead of guessing.

Respond in Polish.
Use English for code, identifiers, API names, database schema names, and technical labels unless explicitly asked otherwise.
Keep user-facing explanations in Polish.
Work in small, verifiable steps.
```