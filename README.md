## README.md for PowerBuf CLI

---

### PowerBuf CLI

The `pbuf` is a Command Line Interface (CLI) tool for PowerBuf, allowing you to manage and vendor modules easily.

---

### Installation

To use `pbuf`, you must first install' Go' on your system. After that, install the application from the source using the following commands:

```bash
go install github.com/pbufio/pbuf-cli
```

---

### Usage

#### General Structure

The general structure of commands is:

```bash
pbuf [command] [arguments...]
```

#### Available Commands

1. **Vendor**

   The vendor command allows you to vendor modules from the provided configuration.

   ```bash
   pbuf vendor
   ```

   By default, this command reads the configuration from `pbuf.yaml`. The configuration provides details like each module's repository, branch or tag, path, and output directory.

---

### Configuration (`pbuf.yaml`)

A typical `pbuf.yaml` file contains the following:

```yaml
version: "1.0"
modules:
  - repository: [repository_url]
    path: [path_in_repository]
    branch: [branch_name]
    tag: [tag_name]
    out: [output_folder_on_local]
```

Replace placeholders with appropriate values:

- `[repository_url]`: The URL of the Git repository.
- `[path_in_repository]`: Path to the folder or file you want to vendor in the repository.
- `[branch_name]`: Specific branch name to clone (optional if tag is provided).
- `[tag_name]`: Specific tag to clone (optional if branch is provided).
- `[output_folder_on_local]`: Folder where the vendor content should be placed on your local machine.

#### Examples
```yaml
version: v1
modules:
  # will copy examples/addressbook.proto file to proto/addressbook.proto
  - repository: https://github.com/protocolbuffers/protobuf
    path: examples/addressbook.proto
    branch: main
    out: proto
   # will copy examples folder to examples folder
  - repository: https://github.com/protocolbuffers/protobuf
    path: examples
    tag: v24.4
```
---

### Private repositories and authentication

To authenticate with repositories that require authentication, `pbuf` uses `.netrc` file. Ensure your `.netrc` file is properly configured in your home directory with credentials.

---

### Contribution

If you'd like to contribute to `pbuf`, feel free to fork the repository and send us a pull request!

---

We hope you find `pbuf` useful for your projects!
