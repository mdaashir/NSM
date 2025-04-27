# New package: nsm

## What does this PR do?

Introduces `nsm` (Nix Shell Manager), a tool to manage Nix development environments.

## Package description

NSM provides features like:

- Initialize new Nix shell environments
- Add/remove packages easily
- List installed packages
- Convert between shell.nix and flake.nix
- Run Nix shells
- Manage Nix channel versions
- Clean up unused packages

## Testing

- Built and tested on:
  - [ ] NixOS
  - [ ] Linux (non-NixOS)
  - [ ] macOS
  - [ ] Other (specify)

## Checklist

- [ ] Checked package builds (`nix-build -A nsm`)
- [ ] Added package to pkgs/top-level/all-packages.nix
- [ ] Added myself as maintainer
- [ ] Package has proper documentation
- [ ] Dependencies are correctly specified
- [ ] License is valid and files properly licensed

## Additional Info

- Project homepage: https://github.com/mdaashir/NSM
- Latest release: v1.1.6
