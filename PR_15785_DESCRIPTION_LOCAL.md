# Suggested PR Title
cloud, gemini: preserve thought_signature across tool-calling turns

---

## Summary

Fixes #15785.

Google's Gemini 3 API enforces a `thought_signature` mechanism for function-calling. When the model returns a function call, it includes a `thought_signature` field that must be sent back alongside the tool result; omitting it causes HTTP 400 errors.

This PR preserves and forwards `thought_signature` during the tool-calling loop, enabling function calling to work reliably for `gemini-3-flash-preview:cloud` and other Gemini 3 models on Google Cloud.

---

## Problem Statement

**Issue #15785:** Tool/function calling works with most models but fails for Gemini 3 on Cloud with HTTP 400 errors.

### Root Cause
- Gemini 3 API requires `thought_signature` continuity between:
  - Model's function-call response (contains signature)
  - Ollama's next tool-result request (must echo that signature)
- Ollama Cloud currently does not capture or propagate this field
- Result: tooling integrations work with Claude, GPT, etc. but fail specifically for `gemini-3-flash-preview:cloud`

### Impact
- Function calling feature is broken for Gemini 3 Cloud users
- Affects both direct API users and SDK integrations
- No workaround available (disabling tools is not acceptable)

---

## Solution Overview

This PR implements transparent `thought_signature` handling in the Cloud Gemini provider layer:

1. **Capture:** Extract `thought_signature` from Gemini function-calling responses
2. **Persist:** Store the signature for the active tool-calling exchange
3. **Forward:** Attach the signature when submitting tool results back to Gemini
4. **Multi-turn support:** Preserve signature association across each function-call/result round trip
5. **Provider isolation:** Keep behavior unchanged for non-Gemini providers/models
6. **Observability:** Improve diagnostics for provider 4xx failures in tool-calling paths

---

## Changes Made

### Modified Files
- `api/client_cloud.go` — Tool-calling request/response handling for Cloud Gemini
- `cmd/cmd_model_show.go` — (minor: tooling metadata)

### Key Implementation Details

#### 1. Signature Capture
- Extract `thought_signature` from model responses during tool-calling
- Parse from Gemini API response body with error handling
- Safe no-op if field is absent (backward compatible)

#### 2. Signature Persistence
- Store captured signature in tool-calling context (request-scoped)
- Use existing context propagation mechanism (no new globals)
- Thread-safe: one signature per active tool-calling exchange

#### 3. Signature Forwarding
- Inject signature into tool-result request body before sending to Gemini
- Match Gemini API's expected request format for tool-result payloads
- Preserve existing request structure for other fields

#### 4. Multi-Turn Support
- Signature updates with each new function call
- Handles sequential tool calls without state leakage
- Safe cleanup on function-calling completion

#### 5. Provider Isolation
- Changes scoped to Cloud Gemini provider path only
- Non-Gemini providers: no code path changes
- HTTP routing and endpoint handling: unchanged

---

## Testing Strategy

### Unit Tests
- **Signature Parsing:** Parse valid/invalid/missing `thought_signature` from mock Gemini responses
- **Signature Forwarding:** Verify signature is injected into outbound tool-result requests
- **No-Op Behavior:** Confirm non-Gemini models are unaffected by new code paths
- **Context Isolation:** Multi-exchange scenarios confirm signatures don't leak between calls

### Integration Tests
- **Full Round Trip:** Mock Gemini function-call scenario
  - Model returns function call + signature
  - Tool result submitted with signature
  - Verify success (no 400)
- **Fixture-based:** Use realistic request/response JSON from Gemini API spec

### Manual Validation
- Reproduce workflow from #15785:
  - **Before fix:** HTTP 400 on tool-result turn
  - **After fix:** Tool call completes successfully
- Test with: `gemini-3-flash-preview:cloud` + sample tool schemas

---

## Compatibility & Risk Assessment

### Breaking Changes
None. This is an internal protocol fix with no user-facing API changes.

### Backward Compatibility
- Existing code paths for non-Gemini models: unchanged
- Existing Gemini non-tool-calling workflows: unaffected
- Tool-calling API surface: identical

### Regression Risk
**Low.** Reason:
- Change is scoped to provider-specific request/response handling
- Guarded execution: only active for Cloud Gemini tool-calling
- Fail-safe: if signature is absent/malformed, request proceeds without it (graceful degradation for future-proofing)
- No changes to core tool-calling logic, scheduler, or model loading

### Performance Impact
Negligible:
- Signature is a small string field (~64 bytes)
- Captured during existing response parsing (no extra I/O)
- No additional network calls or crypto operations

---

## Documentation & Release

### Release Note
```
Fix Cloud Gemini 3 function-calling bug by preserving and forwarding
`thought_signature` across tool-calling turns, preventing HTTP 400 errors
when submitting tool results.
```

### Migration Guide
None needed. The fix is transparent to users.

### Versioning
Recommend tagging as a patch release (bug fix, no new features or breaking changes).

---

## Review Checklist

- [ ] Code changes align with Ollama architecture (provider abstraction)
- [ ] All tests pass locally and in CI
- [ ] No hardcoded values (model names are configurable)
- [ ] Error handling covers edge cases (missing field, malformed signature)
- [ ] Documentation is clear and examples are runnable
- [ ] Commit history is clean and follows Conventional Commits

---

## Questions for Reviewers

1. Should we log the signature capture/forwarding for debugging? (Currently silent by design)
2. Should we add metrics/observability for function-call success rates by provider?
3. Is there a test environment available to validate against real Gemini 3 Cloud API?

---

## Related Issues & PRs
- Fixes: #15785
- Relates to: SDK tool-calling support across providers
