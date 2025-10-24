# Version Management

## Overview

The TypeScript package uses a dual-version system:

1. **Package Version** (`package.json` → `version`): NPM package version
2. **Native Library Version** (`library-version.json` → `version`): Pre-built native library version

This allows you to release patch updates to the TypeScript bindings without rebuilding the native libraries.

## Example Scenario

```
Package v0.4.1 uses native library v0.4.0
Package v0.4.2 uses native library v0.4.0
Package v0.5.0 uses native library v0.5.0
```

## Configuration Files

### `package.json`
```json
{
  "name": "@ideamans/libnextimage",
  "version": "0.4.1",
  ...
}
```

This version is used for:
- NPM package versioning
- Package publishing
- User-facing version number

### `library-version.json`
```json
{
  "version": "0.4.0",
  "description": "Version of the pre-built native library to download",
  "releaseUrl": "https://github.com/ideamans/libnextimage/releases/tag/v0.4.0"
}
```

This version is used for:
- Downloading the correct native library during `npm install`
- Runtime library version information

## Release Process

### Patch Release (TypeScript only)

When you only update TypeScript code without changing native libraries:

```bash
# 1. Update package.json version
npm version patch  # 0.4.0 → 0.4.1

# 2. Keep library-version.json unchanged
# (Still points to v0.4.0)

# 3. Build and publish
npm run build
npm publish
```

### Major/Minor Release (with native library changes)

When you update both TypeScript and native code:

```bash
# 1. Build and release native libraries
cd ..
bash scripts/build-c-library.sh
# Create GitHub Release v0.5.0 with artifacts

# 2. Update library-version.json
cd typescript
cat > library-version.json <<EOF
{
  "version": "0.5.0",
  "description": "Version of the pre-built native library to download",
  "releaseUrl": "https://github.com/ideamans/libnextimage/releases/tag/v0.5.0"
}
EOF

# 3. Update package.json version
npm version minor  # 0.4.1 → 0.5.0

# 4. Build and publish
npm run build
npm publish
```

## How It Works

### Installation Flow

1. User runs: `npm install @ideamans/libnextimage`
2. NPM downloads the package (version 0.4.1)
3. NPM runs the `postinstall` script
4. `postinstall.js` reads `library-version.json` (version 0.4.0)
5. Downloads `libnextimage-{platform}.tar.gz` from GitHub Releases v0.4.0
6. Extracts native library to `lib/{platform}/`

### Runtime

```typescript
import { getLibraryVersion } from '@ideamans/libnextimage'

console.log(getLibraryVersion()) // "0.4.0" (from library-version.json)
```

## Testing Version Management

### Test postinstall script locally

```bash
# Clean lib directory
rm -rf lib

# Run postinstall
node scripts/postinstall.js

# Verify library was downloaded
ls -la lib/*/
```

### Test with different versions

```bash
# Temporarily change library version
echo '{"version":"0.3.0","description":"Test"}' > library-version.json

# Clean and reinstall
rm -rf lib
node scripts/postinstall.js

# Verify it downloaded v0.3.0
```

## Troubleshooting

### "Failed to download: HTTP 404"

The specified library version doesn't exist on GitHub Releases.

**Solutions:**
1. Check that the GitHub Release exists: https://github.com/ideamans/libnextimage/releases/tag/v{VERSION}
2. Verify the release has the required artifacts (`libnextimage-{platform}.tar.gz`)
3. Update `library-version.json` to an existing version

### Version Mismatch Warning

If you see warnings about version mismatches, verify:

```bash
# Check package version
cat package.json | grep version

# Check library version
cat library-version.json | grep version

# Check installed library
ls -la lib/*/
```

## Best Practices

1. **Always test postinstall** before publishing
2. **Document version changes** in CHANGELOG.md
3. **Use semantic versioning**:
   - Patch (0.4.0 → 0.4.1): Bug fixes in TypeScript code
   - Minor (0.4.0 → 0.5.0): New features (may include native changes)
   - Major (0.5.0 → 1.0.0): Breaking changes
4. **Keep library-version.json in sync** with actual GitHub Releases
5. **Test on all platforms** before releasing

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Verify native library download
  run: |
    cd typescript
    rm -rf lib
    node scripts/postinstall.js
    ls -la lib/*/
```

This ensures the postinstall script works correctly for the specified library version.
