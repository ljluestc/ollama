Suggested PR Title:
cloud, gemini: preserve thought_signature across tool-calling turns

## Summary
Fixes #15785.

Gemini 3 on Cloud includes a `thought_signature` in function-calling responses.
If that signature is not returned with the next tool result turn, Gemini responds
with HTTP 400. This change preserves and forwards `thought_signature` during the
tool-calling loop so function calling works for `gemini-3-flash-preview:cloud`.

## Problem
- Tool/function calling works with most models, but fails for Gemini 3 on Cloud.
- Gemini 3 requires `thought_signature` continuity between model function-call
  output and the subsequent tool-result input.
- Current Cloud handling does not propagate this field, causing request rejection.

## What this PR changes
1. Capture `thought_signature` from Gemini function-calling responses.
2. Persist the captured signature for the active tool-calling exchange.
3. Attach the signature when submitting tool results back to Gemini.
4. Support multi-tool turns by preserving signature association across each
   function-call/result round trip.
5. Keep behavior unchanged for non-Gemini providers/models.
6. Improve diagnostics for provider 4xx failures in tool-calling paths.

## Why this fix
- Matches Gemini 3 protocol expectations for function calling.
- Removes model-specific 400 failures without changing user-facing tool APIs.
- Keeps existing behavior stable for other providers.

## Validation
- Unit tests covering:
  - parsing/capturing `thought_signature` from Gemini responses
  - forwarding signature in tool-result requests
  - no-op behavior for providers/models that do not use the field
- Integration test covering a full function-call round trip with
  `gemini-3-flash-preview:cloud` request/response fixtures.
- Manual verification by reproducing the reported workflow from #15785:
  - before fix: HTTP 400 on tool-result turn
  - after fix: tool call completes successfully

## Compatibility and risk
- No breaking API changes.
- Change is scoped to provider-specific request/response handling in the Cloud
  Gemini tool-calling path.
- Low regression risk for non-Gemini models due to guarded execution paths.

## Release note
```release-note
Fix a Cloud Gemini 3 function-calling bug by preserving and forwarding
`thought_signature` across tool-calling turns, preventing HTTP 400 errors when
submitting tool results.
```
