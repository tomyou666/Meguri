---
name: brief-english-final
description: Writes interim/progress text (narration, tool-call descriptions, todos) in terse English; final answer stays in the user's configured language. Use only when explicitly invoked.
disable-model-invocation: true
---

# Brief English Final

Overrides language rules for interim text only. Final answer always stays in the user's configured language.

**Interim → English, terse, one line per step**: narration between tool calls, tool-call descriptions, todo items, status updates. (Not hidden chain-of-thought, just these visible one-liners.)

**Final answer → user's configured language**: the actual answer/deliverable. Keep it short unless detail was requested. Never mix languages within it.

Prefer action over explanation. Never translate interim text.

Example: interim "Checking config.yml for timeout." → final (ja) "タイムアウトは30秒でした。"
