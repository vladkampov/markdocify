# CI/CD Setup Guide for markdocify

This guide walks you through setting up the complete CI/CD pipeline for secure, automated releases.

## 🔐 Security-First Architecture

Our CI/CD pipeline follows security best practices:
- **OIDC authentication** (no long-lived tokens)
- **Minimal permissions** for each workflow
- **Dependency scanning** with automated updates
- **Code signing** with cosign
- **Supply chain security** with SLSA attestations
- **Container scanning** with Trivy

## 📋 Prerequisites

1. **GitHub repository** with admin access
2. **GitHub Personal Access Token** (for Homebrew tap)
3. **Package registry access** (optional, for containers)

## 🚀 Quick Setup

### 1. Repository Secrets

Add these secrets to your GitHub repository (`Settings > Secrets and variables > Actions`):

#### Required Secrets:
```bash
# For Homebrew formula updates
HOMEBREW_TAP_GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx

# For Codecov (optional)
CODECOV_TOKEN=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

#### Optional Secrets (for additional package managers):
```bash
# For Windows Scoop bucket
SCOOP_TAP_GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx

# For AUR (Arch Linux)
AUR_KEY=-----BEGIN OPENSSH PRIVATE KEY-----
```

### 2. Create Homebrew Tap Repository

1. Create a new repository: `https://github.com/vladkampov/homebrew-tap`
2. Initialize with a README
3. The release workflow will automatically create/update the formula

### 3. Enable Security Features

1. **CodeQL**: Go to `Security > Code scanning > Set up CodeQL`
2. **Dependabot**: Already configured via `.github/dependabot.yml`
3. **Secret scanning**: Enable in repository settings

## 🔧 Configuration Files

### Workflow Files Created:

```
.github/
├── workflows/
│   ├── ci.yml          # Main CI: test, lint, build
│   ├── release.yml     # Release automation
│   └── security.yml    # Security scanning
└── dependabot.yml      # Dependency updates
```

### Build Configuration:
```
├── .goreleaser.yml     # Cross-platform builds & packaging
└── Dockerfile         # Container builds
```

## 🔄 Workflow Triggers

### CI Workflow (`ci.yml`)
- **Triggers**: Push to `main`/`develop`, PRs to `main`
- **Actions**: Test, lint, security scan, build artifacts
- **Matrix**: Go 1.21 & 1.22

### Release Workflow (`release.yml`)
- **Triggers**: Git tags (`v*`), manual dispatch
- **Actions**: Build, sign, package, publish to:
  - GitHub Releases
  - Homebrew
  - Container Registry
  - Snap Store
  - Windows Scoop

### Security Workflow (`security.yml`)
- **Triggers**: Daily schedule, pushes, PRs
- **Actions**: Vulnerability scanning, license checks, SAST

## 📦 Package Managers Supported

### ✅ Automatically Configured:
- **GitHub Releases** (binaries + checksums)
- **Homebrew** (macOS/Linux)
- **Docker/OCI** (containers)
- **Snap** (Ubuntu/Linux)
- **Scoop** (Windows)

### 🔧 Manual Setup Required:
- **AUR** (Arch Linux) - requires SSH key
- **Chocolatey** (Windows) - requires account

## 🏃‍♂️ Creating Your First Release

### Method 1: Git Tag (Recommended)
```bash
# Create and push a version tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Method 2: Manual Trigger
1. Go to `Actions > Release`
2. Click `Run workflow`
3. Enter version (e.g., `v1.0.0`)

## 🔍 Monitoring & Maintenance

### Daily Automated Tasks:
- **Dependency updates** (Dependabot)
- **Security scanning** (scheduled workflow)
- **License compliance** checks

### Manual Tasks:
- **Review security alerts** in GitHub Security tab
- **Approve dependency updates** PRs
- **Monitor release success** in Actions tab

## 🛡️ Security Features

### Code Signing
All releases are signed with **cosign**:
```bash
# Verify a release
cosign verify-blob \
  --certificate markdocify_1.0.0_checksums.txt.pem \
  --signature markdocify_1.0.0_checksums.txt.sig \
  markdocify_1.0.0_checksums.txt
```

### SLSA Attestations
Build provenance is recorded for supply chain security:
- Build environment details
- Source code commit hash
- Build steps executed

### Vulnerability Scanning
- **Go modules**: govulncheck, nancy
- **Container images**: Trivy
- **Code**: CodeQL, gosec
- **Dependencies**: GitHub Security Advisories

## 🔧 Customization

### Adding New Package Managers

Edit `.goreleaser.yml` to add new targets:
```yaml
# Example: Add Chocolatey
chocolateys:
  - name: markdocify
    owners: vladkampov
    title: markdocify
    # ... configuration
```

### Custom Build Flags
Modify the `builds` section in `.goreleaser.yml`:
```yaml
builds:
  - ldflags:
      - -X main.customFlag=value
```

### Additional Security Scans
Add steps to `.github/workflows/security.yml`:
```yaml
- name: Custom security check
  run: |
    # Your custom security tools
```

## 📊 Metrics & Observability

### GitHub Insights
- **Actions**: Workflow success rates
- **Security**: Vulnerability trends
- **Community**: Download statistics

### External Services
- **Codecov**: Test coverage tracking
- **Snyk**: Advanced vulnerability scanning (optional)

## 🆘 Troubleshooting

### Common Issues:

1. **Release fails with permission error**
   - Check `HOMEBREW_TAP_GITHUB_TOKEN` has `repo` scope
   - Verify token hasn't expired

2. **Security scan fails**
   - Check if new vulnerabilities were introduced
   - Review gosec findings for false positives

3. **Build fails on specific platform**
   - Check Go version compatibility
   - Review cross-compilation requirements

### Debug Mode:
Enable debug logging by adding to workflow:
```yaml
env:
  ACTIONS_STEP_DEBUG: true
```

## 🎯 Best Practices

### Security:
- ✅ Regular dependency updates
- ✅ Pin action versions to commit hashes
- ✅ Use minimal permissions
- ✅ Sign all releases
- ✅ Scan for vulnerabilities

### Reliability:
- ✅ Test on multiple Go versions
- ✅ Cross-platform builds
- ✅ Comprehensive test coverage
- ✅ Automated rollback on failures

### Maintainability:
- ✅ Clear commit messages
- ✅ Automated changelog generation
- ✅ Version tagging consistency
- ✅ Documentation updates

## 📞 Support

- **GitHub Issues**: Bug reports and feature requests
- **Security Issues**: Email security@vladkampov.com
- **Documentation**: Check repository wiki

---

**🔒 This pipeline prioritizes security without sacrificing simplicity.**