# Fix embedding handlers to return early on empty input and guard runner acquisition

## Summary
This PR tightens the `/api/embed` and `/api/embeddings` handlers in `server/routes.go` so they fail fast on empty requests and handle runner acquisition/context failures explicitly.

## What changed
- `/api/embed`
  - Return an empty embedding response immediately when `input` is empty, before scheduling a runner.
  - Check for a nil runner after `scheduleRunner` and return a server error instead of continuing.
  - Check `c.Request.Context().Err()` after scheduling and return a timeout-style error if the request context is no longer valid.

- `/api/embeddings`
  - Return an empty embedding response immediately when `req.Prompt` is empty.
  - Check for a nil runner after `scheduleRunner`.
  - Check `c.Request.Context().Err()` after scheduling and return a timeout-style error if the request context has been canceled or expired.

## Why
These handlers should not proceed with model scheduling when the request is already empty, canceled, or otherwise unable to continue safely. The early return keeps the empty-request behavior intact while avoiding unnecessary runner work and making failure modes explicit.

## Behavior
- Empty embed requests still return success with an empty embedding payload.
- Empty embeddings requests still return success with an empty embedding payload.
- If the scheduler cannot provide a runner, the handlers now return a clear server-side error instead of proceeding with invalid state.
- If the request context has already expired, the handlers return a timeout response with context details.

## Testing
Not run locally.

## Notes
This change is limited to request handling in `server/routes.go`. It does not change model behavior, response schemas, or scheduler logic.
