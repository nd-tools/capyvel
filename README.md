# CAPYVEL Project

Welcome to the **CAPYVEL** project. This repository contains the source code and documentation for the project.

## Versions

This project follows a version control scheme based on **Semantic Versioning** (SemVer).

### Latest Version: v1.1

#### Changes in this version:
- Initial stable version of the project.
- Basic functions implemented.
- Project structure configured.
  



---

### How to Add New Versions?

When you want to add a new version, simply commit your changes and use the following command to create a new tag:

```bash
git tag -a v1.1 -m "Description of the new version"
git push origin v1.1


1. Get a specific version of the package:

```bash
go get github.com/nd-tools/capyvel@v1.0

    This command retrieves the specific version v1.0 of the package capyvel from the GitHub repository github.com/nd-tools/capyvel.
    It ensures that your project uses exactly version v1.0.
    The version number you specify must exist as a tag in the repository (for example, v1.0, v1.1, etc.).
    After running this, the go.mod file will reflect the exact version v1.0 of the package, for example:

    require github.com/nd-tools/capyvel v1.0.0

2. Get the latest version of the package:

```bash
go get github.com/nd-tools/capyvel@latest

    This command fetches the latest version of the capyvel package from the repository.
    The latest keyword refers to the most recent tagged version available in the repository, which might be the latest stable release.
    It is particularly useful if you want to ensure your project is using the most up-to-date version of a package.
    After running this, the go.mod file will reflect the latest version of the package (e.g., v1.1.0 if that is the most recent release).