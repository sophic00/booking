# Project Guidelines

## Version Control System (VCS)
- **Use Jujutsu (`jj`)**: Use `jj` instead of `git` for all version control operations.
- **Non-interactive Paging**: Always include the `--no-pager` flag with all `jj` commands to avoid interactive paging (e.g., `jj --no-pager status`, `jj --no-pager diff`, `jj --no-pager log`).

## Development Environment & Toolchain
- **Use Nix Flake**: Run all build, test, lint, and runtime commands inside the Nix dev environment:
  - Format: `nix develop --command <command>` (e.g., `nix develop --command go test ./...`, `nix develop --command npm run build`).
