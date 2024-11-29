# ND Project - Your Project Name

Welcome to the **ND** project. This repository contains the source code and documentation for the project.

## Versions

This project follows a version control scheme based on **Semantic Versioning** (SemVer).

### Latest Version: v1.0

#### Changes in this version:
- Initial stable version of the project.
- Basic functions implemented.
- Project structure configured.
  
### Version History

| Version | Date       | Description                                      |
|---------|------------|--------------------------------------------------|
| **v1.0** | Nov 29, 2024 | Initial version of the project. First stable release. |
| **v0.1** | Nov 10, 2024 | Preliminary version with initial features. |


### Usage

Instructions on how to use the project or the tools it contains.

### Contributing

If you'd like to contribute to the project, please follow these steps:

1. Fork the repository.
2. Create a branch for your changes (`git checkout -b feature/new-feature`).
3. Make your changes and commit them.
4. Push your changes to your fork (`git push origin feature/new-feature`).
5. Create a Pull Request for us to review your changes.

---

### Explanation of the `README.md`:

- **Versions**: The **Versions** section contains the list of versions for the project, following the versioning scheme. You can add more versions here as you update the project.
- **Version History**: A table where you can add the version, the date, and a brief description of the changes.
- **Installation and Usage**: This section contains instructions for other developers to install and use the project.
- **Contributing**: Includes instructions for other users to contribute to the project.
- **License**: Links to the license under which the project is distributed (if applicable).

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