# Feature: Auto-size num_ctx to a user VRAM budget

## Summary
Closes #12353

This PR introduces a new capability to automatically sizing the context window (`num_ctx`) based on the available VRAM capabilities of the user's GPU(s). This solves a common issue where users must guess a safe `num_ctx` to avoid overflowing VRAM and falling back to CPU (which causes severe performance degradation).

Two new flags are introduced:
- `--fit-vram`: Automatically calculates the largest safe context length that fits in the available GPU memory.
- `--max-vram`: (Optional) Sets a strict VRAM budget (e.g., `8GB`, `4096MB`) to leave headroom for other applications or strictly limit Ollama's usage.

## Changes

### 1. API (`api/types.go`)
- Added `FitVRAM` (bool) and `MaxVRAM` (uint64) fields to the `Runner` struct to transport these options from the client to the server.

### 2. CLI (`cmd/cmd.go`)
- Added parsing for `--fit-vram` and `--max-vram` flags in the `ollama run` command.
- Implemented a `parseBytes` helper to handle human-readable memory strings (e.g., "12GB", "512MB").
- Passes these new options into the runner configuration.

### 3. Server Logic (`llm/server.go`)
- Implemented `estimateMemoryUsage` function to accurately predict the memory footprint of:
    - Model weights
    - KV Cache (at a specific context length)
    - Graph / Workspace memory
- Updated `NewLlamaServer` initialization flow:
    - If `FitVRAM` is enabled, it queries the free memory of all visible GPUs.
    - Reserves a safety headroom (defaulting to 15%) to avoid fragmentation and system overhead.
    - Performs a **binary search** between the minimum context (2048) and the model's training limit (e.g. 128k) to find the maximum `num_ctx` that fits within the calculated budget.
    - Clamps the result to the model's `trainCtx`.
    - Updates `opts.NumCtx` and `loadRequest.KvSize` dynamically before the model is allocated.

## Usage Example

**Basic Usage (Fit to available GPU memory):**
```bash
ollama run llama3:8b --fit-vram
```

**With a Strict Budget (e.g., on a 24GB card, leaving space for other tasks):**
```bash
ollama run llama3:70b --fit-vram --max-vram=16GB
```

**Output:**
The server logs will confirm the adjustment:
```text
level=WARN source=server.go:123 msg="minimal context does not fit in VRAM" ...
level=INFO source=server.go:145 msg="auto-sized num_ctx" original=8192 new=6144 available_vram=6.2GB
```

## Impact
- **Stability**: Drastically reduces OOM crashes and unintended CPU fallbacks.
- **Performance**: Ensures models run entirely on GPU whenever possible.
- **UX**: Removes the need for users to manually calculate and trial-and-error their context window settings.
