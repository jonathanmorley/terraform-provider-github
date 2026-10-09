# Fork control plane

This branch (`main`) is the **control plane** for this fork. It
contains no provider code — only the machinery that maintains it:

- `.github/workflows/` — the synthesize + update callers
- `synthesis.json` / `synthesis.lock` — the declarative spec + pins
- `.github/chainguard/` — OIDC trust policy for automation tokens

The generated product lives on `release`, rebuilt from upstream plus the
declared patches on every run. Never commit code here; never commit
directly to `release` (it gets force-pushed). Persistent fork content goes
on a patch branch declared in `synthesis.json`.
