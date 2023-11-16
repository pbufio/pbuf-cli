## README.md for PowerBuf CLI

---

### PowerBuf CLI

The `pbuf-cli` is a Command Line Interface (CLI) tool for PowerBuf, allowing you to easily manage, and vendor modules.

We recommend using `pbuf-cli` with the [PowerBuf Registry](https://github.com/pbufiio/pbuf-registry).

---

### Installation

#### Binary

Grab latest binaries from the releases page: https://github.com/pbufio/pbuf-cli/releases.

#### From sources

First, you must ensure that you have `Go` installed on your system. 

After that, install application from source using the following commands:

```bash
go install github.com/pbufio/pbuf-cli@latest
```

or for specific version:

```bash
go install github.com/pbufio/pbuf-cli@<tag>
```

---

### Usage

#### General Structure

The general structure of commands is:

```bash
pbuf-cli [command] [arguments...]
```

#### Available Commands

##### Vendor

The vendor command allows you to vendor modules from provided configuration.

```bash
pbuf-cli vendor
```

By default, this command reads the configuration from `pbuf.yaml`. The configuration provides details like the repository, branch or tag, path, and output directory for each module.

##### Register Module

The register command allows you to register a module to the registry.
    
 ```bash
 pbuf-cli modules register
 ```

You can register a module by providing the following details in `pbuf.yaml`:

```yaml
version: v1
name: [module_name]
...
```

Replace `[module_name]` with the name of the module you want to register.

##### Get Module

The get command allows you to get the information about a module from the registry.

```bash
pbuf-cli modules get [module_name]
```

Response example:
```json
{
  "id": "b9f9898a-8510-4017-9618-5244176eb1b8",
  "name": "pbufio/pbuf-registry",
  "tags": [
    "v0.0.0",
    "v0.1.0"
  ]
}
```

##### List Modules

The list command allows you to list all modules from the registry.

```bash
pbuf-cli modules list
```

Response example:
```json
[
  {
    "id": "b9f9898a-8510-4017-9618-5244176eb1b8",
    "name": "pbufio/pbuf-registry"
  },
  {
    "id": "b9f9898a-8510-4017-9618-5244176eb1b8",
    "name": "pbufio/pbuf-cli"
  }
]
```

##### Push Module

The push command allows you to push `.proto` files to the registry with a specific tag.

```bash
pbuf-cli modules push [tag]
```

Replace `[tag]` with the tag you want to push.

#### Delete Tag

The delete command allows you to delete a tag from the registry.

```bash
pbuf-cli modules delete-tag [tag]
```

The command deletes all the proto files associated with the tag.

#### Delete Module

The delete command allows you to delete a module from the registry.

```bash
pbuf-cli modules delete [module_name]
```

The command deletes all the tags and proto files associated with the module.

---

### Configuration (`pbuf.yaml`)

A typical `pbuf.yaml` file contains the following:

```yaml
version: v1
name: [module_name]
registry:
  addr: [registry_url]
  insecure: true # no tls support at the moment, but it will be added soon
export:
  paths:
    - [proto_files_path]
modules:
  # use the registry to vendor .proto files
  - name: [dependency_module_name]
    path: [path_in_registry]
    tag: [dependency_module_tag]
    out: [output_folder_on_local]
    gen_out: [gen_output_folder_on_local] # optional, if provided then patchers will be applied
  # use a git repository to vendor .proto files
  - repository: [repository_url]
    path: [path_in_repository]
    branch: [branch_name]
    tag: [tag_name]
    out: [output_folder_on_local]
    gen_out: [gen_output_folder_on_local] # optional, if provided then patchers will be applied
```

Replace main placeholders with appropriate values:
- `[module_name]`: The module name you want to register.
- `[registry_url]`: The URL of the pbuf-registry.
- `[proto_files_path]: One or several paths that contain `.proto` files.

Replace placeholders in the registry modules with appropriate values:
- `[dependency_module_name]`: The module name you want to vendor.
- `[path_in_registry]`: Path to the folder or file in the registry you want to vendor.
- `[tag_name]`: Specific tag to vendor.
- `[output_folder_on_local]`: Folder where the vendor content should be placed on your local machine.
- `[gen_output_folder_on_local]`: Folder where the generated content should be placed on your local machine. Used to patch `go_package` option

Replace placeholders in modules placed in git with appropriate values:
- `[repository_url]`: The URL of the Git repository.
- `[path_in_repository]`: Path to the folder or file in the repository you want to vendor.
- `[branch_name]`: Specific branch name to clone (optional if tag is provided).
- `[tag_name]`: Specific tag to clone (optional if branch is provided).
- `[output_folder_on_local]`: Folder where the vendor content should be placed on your local machine.
- `[gen_output_folder_on_local]`: Folder where the generated content should be placed on your local machine. Used to patch `go_package` option

#### Examples

#### Push Module
```yaml
version: v1
name: pbufio/pbuf-registry
registry:
  addr: pbuf.cloud:8081
  insecure: true
# all `.proto` files from `api` and `entities` folders
# will be exported as the module proto files
export:
  paths:
    - api
    - entities
modules: []
```

#### Vendor Modules
```yaml
version: v1
name: pbuf-cli
registry:
   addr: pbuf.cloud:8081
   insecure: true
modules:
  # will copy api/v1/*.proto file to third_party/api/v1/*.proto
  - name: pbufio/pbuf-registry
    tag: v0.0.1
    out: third_party
  # will copy api/v1/*.proto file to third_party/api/v1/*.proto
  # and add or change `go_package` option to `<go_mod_name>/gen/pbuf-registry/api/v1`
  - name: pbufio/pbuf-registry
    tag: v0.0.1
    out: third_party
    gen_out: gen
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

To authenticate with repositories that require authentication, `pbuf-cli` uses `.netrc` file. Ensure your `.netrc` file is properly configured in your home directory with credentials.

---

### Contribution

If you'd like to contribute to `pbuf`, feel free to fork the repository and send us a pull request!

---

We hope you find `pbuf` useful for your projects!
