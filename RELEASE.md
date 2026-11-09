# Release Process

This document describes the release process for libnextimage, which includes C libraries, Go bindings, and TypeScript bindings.

## Overview

The release process is automated using GitHub Actions and follows this flow:

1. **Tag and Release** - Create a version tag (e.g., `v0.5.0`)
2. **Build Binaries** - CI builds all platforms and creates GitHub Release
3. **Update Modules** - Automated PR adds binaries to Go/TypeScript modules
4. **Merge PR** - Review and merge the automated PR
5. **Publish** - Tag Go modules, and push a `typescript/v*` tag to publish TypeScript to npm (CI)

## Detailed Steps

### 1. Create Release Tag

```bash
# Ensure main is up to date
git checkout main
git pull

# Create and push tag
git tag v0.5.0
git push origin v0.5.0
```

### 2. Automated Build and Release

When you push a `v*` tag, the `release.yml` workflow automatically:

- Builds C libraries for all platforms:
  - darwin-arm64 (macOS Apple Silicon)
  - darwin-amd64 (macOS Intel)
  - linux-amd64 (Linux x64)
  - linux-arm64 (Linux ARM64)
  - windows-amd64 (Windows x64)

- Creates both static (`.a`) and shared (`.dylib`/`.so`/`.dll`) libraries

- Creates a GitHub Release with:
  - `libnextimage-v0.5.0-{platform}.tar.gz` (static library + headers)
  - `libnextimage-shared-v0.5.0-{platform}.tar.gz` (shared library + headers)

### 3. Automated Binary Update

`release.yml` starts the `update-binaries.yml` workflow (workflow_dispatch) right after it creates the release.
Both workflows use `GITHUB_TOKEN` only (no PAT). If the release was created by hand, start it with
`gh workflow run update-binaries.yml -f tag=v0.5.0`.

**The PR it opens is created with `GITHUB_TOKEN`, so CI does not run on it.** Check out the PR branch and run the
Go and TypeScript tests locally (or push an empty commit from your account) before merging.

The workflow:

- Downloads all platform binaries from the Release
- Organizes them into the project structure:
  ```
  golang/
    cwebp/
      lib/
        darwin-arm64/libnextimage.a
        darwin-amd64/libnextimage.a
        linux-amd64/libnextimage.a
        linux-arm64/libnextimage.a
        windows-amd64/libnextimage.a
      include/
        nextimage.h
        nextimage/
          cwebp.h
    avifenc/
      lib/{platform}/libnextimage.a
      include/...
    # ... other modules

  typescript/
    lib/
      darwin-arm64/libnextimage.dylib
      darwin-amd64/libnextimage.dylib
      linux-amd64/libnextimage.so
      linux-arm64/libnextimage.so
      windows-amd64/libnextimage.dll
  ```

- Updates TypeScript `package.json` to version `0.5.0`
- Removes `typescript/library-version.json` (no longer needed)
- Updates Go module CGO directives to use embedded libraries
- Creates a Pull Request with all changes

### 4. Review and Merge PR

1. Check the automated PR: `Update binaries for v0.5.0`
2. Review the changes:
   - All platform binaries are included
   - TypeScript version is updated
   - Go CGO directives are correct
3. Merge the PR to `main`

### 5. Tag Go Modules

After merging the PR, tag each Go submodule:

```bash
# Update local main
git checkout main
git pull

# Tag Go modules
git tag golang/v0.5.0
git tag golang/cwebp/v0.5.0
git tag golang/dwebp/v0.5.0
git tag golang/avifenc/v0.5.0
git tag golang/avifdec/v0.5.0
git tag golang/gif2webp/v0.5.0
git tag golang/webp2gif/v0.5.0

# Push all tags
git push --tags
```

### 6. Publish TypeScript to npm

After the binaries PR is merged (it sets `typescript/package.json` to the new version and adds `typescript/lib/`),
push a `typescript/v*` tag. `.github/workflows/publish.yml` tests and publishes with npm **Trusted Publishing**
(GitHub Actions OIDC). No npm token is used and nobody runs `npm publish` locally.

```bash
git checkout main && git pull
git tag typescript/v0.5.0
git push origin typescript/v0.5.0

# Verify
npm view libnextimage version
```

The job fails if the tag differs from `typescript/package.json`, and does nothing if the version is already on npm.
The plain `v*` tag cannot be used for this: at that point the binaries and the version bump are not on `main` yet.
To publish without a tag (e.g. a version that was merged but never released), start "Publish to npm" with
workflow_dispatch on `main`.

Prerequisite: on npmjs.com, `libnextimage` lists GitHub Actions / `ideamans/libnextimage` / `publish.yml` with the
**publish** permission as its Trusted Publisher.

## Manual Workflow Trigger

If the automated workflow fails or needs to be re-run:

```bash
# Trigger manually via GitHub CLI
gh workflow run update-binaries.yml -f tag=v0.5.0

# Or via GitHub web UI:
# Actions → Update Binaries → Run workflow → Enter tag
```

## Versioning Strategy

- **Main library**: `v0.5.0`
- **Go modules**: `golang/v0.5.0`, `golang/cwebp/v0.5.0`, etc.
- **TypeScript**: `0.5.0` on npm, released by the `typescript/v0.5.0` tag

## Platform Support

| Platform | Architecture | Static Library | Shared Library | Status |
|----------|-------------|----------------|----------------|--------|
| macOS | ARM64 | ✅ | ✅ | Supported |
| macOS | AMD64 | ✅ | ✅ | Supported |
| Linux | AMD64 | ✅ | ✅ | Supported |
| Linux | ARM64 | ✅ | ✅ | Supported |
| Windows | AMD64 | ✅ | ✅ | Supported |

## Troubleshooting

### Workflow failed to download binaries

Check that all platform builds completed successfully in the release workflow.

### Go modules can't find libraries

Ensure the PR was merged and Go module tags were pushed after the merge.

### TypeScript npm publish fails

1. Check `typescript/package.json` has the correct version
2. Verify you're logged in to npm: `npm whoami`
3. Check that the version doesn't already exist: `npm view libnextimage versions`

## Rollback

If a release needs to be rolled back:

```bash
# Delete tags
git tag -d v0.5.0
git push origin :refs/tags/v0.5.0

# Delete Go module tags
git tag -d golang/v0.5.0
# ... delete other module tags

# Unpublish from npm (within 72 hours)
npm unpublish libnextimage@0.5.0

# Delete GitHub Release
gh release delete v0.5.0
```
