# ⚠️ README.md Configuration Notice

This README.md file is configured to be **excluded from GitLab** but **included in GitHub**.

## How It Works

### For GitHub (Normal Behavior)
- README.md is included normally
- All content is visible
- Standard GitHub README display

### For GitLab (Excluded)
- README.md is excluded via `.gitattributes` (`export-ignore`)
- Or via GitLab CI/CD configuration
- File exists in repository but is not displayed/exported

## Configuration Files

1. **`.gitattributes`** - Uses `export-ignore` to exclude from GitLab exports
2. **`.gitlab-ci-exclude-readme.yml`** - CI/CD script to remove README.md from GitLab

## Manual Exclusion (If Needed)

If automatic exclusion doesn't work, you can manually exclude README.md from GitLab:

```bash
# Add to .gitignore (not recommended - affects all remotes)
echo "README.md" >> .gitignore

# Or use GitLab's .gitlab-ci.yml to remove it
# See .gitlab-ci-exclude-readme.yml for example
```

## Verification

To verify README.md is excluded from GitLab:

```bash
# Check GitLab export
git archive --format=tar HEAD | tar -t | grep README.md
# Should return empty (file excluded)

# Check GitHub (normal)
git ls-files | grep README.md
# Should return README.md (file included)
```
