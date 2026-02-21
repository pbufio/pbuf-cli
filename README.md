## README.md for PowerBuf CLI

---

### PowerBuf CLI

The `pbuf` is a Command Line Interface (CLI) tool for PowerBuf, allowing you to easily manage, and vendor modules.

We recommend using `pbuf` with the [PowerBuf Registry](https://github.com/pbufio/pbuf-registry).

---

### Installation

#### Quick Install (Linux/macOS)

Install the latest version with a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/pbufio/pbuf-cli/main/install.sh | sh
```

Or using wget:

```bash
wget -qO- https://raw.githubusercontent.com/pbufio/pbuf-cli/main/install.sh | sh
```

The script will:
- Auto-detect your OS and architecture
- Download the latest release
- Install to `/usr/local/bin` (may require sudo)

To install to a custom directory:

```bash
INSTALL_DIR=$HOME/.local/bin curl -fsSL https://raw.githubusercontent.com/pbufio/pbuf-cli/main/install.sh | sh
```

#### Windows

**Option 1: Manual Download**

1. Download the latest Windows release from: https://github.com/pbufio/pbuf-cli/releases
2. Extract the `pbuf.exe` binary from the zip archive
3. Add the binary to your PATH:
   - Right-click "This PC" → Properties → Advanced system settings → Environment Variables
   - Under "System variables", find and edit "Path"
   - Add the directory containing `pbuf.exe`
4. Verify installation by opening a new terminal and running:
   ```cmd
   pbuf --help
   ```

**Option 2: PowerShell Script**

```powershell
# Download and extract (run as Administrator if installing to Program Files)
$version = (Invoke-RestMethod "https://api.github.com/repos/pbufio/pbuf-cli/releases/latest").tag_name
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$url = "https://github.com/pbufio/pbuf-cli/releases/download/$version/pbuf_${version}_windows_${arch}.zip"
$output = "$env:TEMP\pbuf.zip"

Invoke-WebRequest -Uri $url -OutFile $output
Expand-Archive -Path $output -DestinationPath "$env:TEMP\pbuf" -Force
Move-Item -Path "$env:TEMP\pbuf\pbuf.exe" -Destination "C:\Windows\System32\" -Force
Remove-Item -Path $output, "$env:TEMP\pbuf" -Recurse -Force

Write-Host "pbuf installed successfully!"
pbuf --help
```

#### From Source

If you have Go installed, you can build from source:

```bash
go install github.com/pbufio/pbuf-cli@latest
```

Or for a specific version:

```bash
go install github.com/pbufio/pbuf-cli@<tag>
```

#### Manual Binary Download

Download pre-built binaries for all platforms from the releases page: https://github.com/pbufio/pbuf-cli/releases

---

### Self-Hosted Quickstart

Get started with a self-hosted PowerBuf Registry in under 5 minutes:

#### 1. Start the Registry

```bash
# Using Docker
docker run -d -p 8080:8080 pbufio/pbuf-registry:latest

# Or using Docker Compose
# See https://github.com/pbufio/pbuf-registry for examples
```

#### 2. Initialize Your Project

```bash
# Initialize with your registry URL
pbuf init my-module http://localhost:8080

# Or for HTTPS with valid cert
pbuf init my-module https://registry.mycompany.com
```

#### 3. Authenticate (if required)

```bash
# Set your authentication token
pbuf auth your-token-here
```

#### 4. Create a Proto File

```bash
mkdir -p api/v1
cat > api/v1/service.proto << 'EOF'
syntax = "proto3";
package mycompany.mymodule.v1;

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
}
EOF
```

Update `pbuf.yaml` to export your proto files:
```yaml
export:
  paths:
    - api
```

#### 5. Push Your Module

```bash
pbuf modules push v0.0.1
```

#### 6. Vendor in Another Project

```bash
# In another project
pbuf init consumer-service http://localhost:8080
# Edit pbuf.yaml to add your module as dependency
pbuf vendor
```

Your protos are now vendored in `third_party/` and ready to use!

---

### Usage

#### General Structure

The general structure of commands is:

```bash
pbuf [command] [arguments...]
```

#### Available Commands

##### Init

The init command allows you to initialize a new `pbuf.yaml` file.

Interactively:
```bash
pbuf init
```

Non-interactively:
```bash
pbuf init [module_name] [registry_url]
```

Replace `[module_name]` with the name of the module you want to register. Replace `[registry_url]` with the URL of the PowerBuf Registry (by default it uses `pbuf.cloud`).

##### Auth (if applicable)

The auth command allows you to authenticate with the registry. It saves the token in the `.netrc` file.

```bash
pbuf auth [token]
```

##### Vendor

The vendor command allows you to vendor modules from provided configuration.

```bash
pbuf vendor
```

By default, this command reads the configuration from `pbuf.yaml`. The configuration provides details like the repository, branch or tag, path, and output directory for each module.

##### Register Module

The register command allows you to register a module to the registry.
    
 ```bash
 pbuf modules register
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
pbuf modules get [module_name]
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

> If `module_name` is not provided, the `name` from `pbuf.yaml` is used.

##### List Modules

The list command allows you to list all modules from the registry.

```bash
pbuf modules list
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
pbuf modules push [tag] [--draft]
```

Replace `[tag]` with the tag you want to push. Use the `--draft` flag to push a draft tag.

> Draft tags are temporary tags that are automatically deleted in a week.

##### Update Modules Tags

The update command allows you to update the modules' tags to the latest in the registry. The command saves the latest tags in the `pbuf.yaml` file for each module.

```bash
pbuf modules update
```

##### Delete Tag

The delete command allows you to delete a tag from the registry.

```bash
pbuf modules delete-tag [tag]
```

The command deletes all the proto files associated with the tag.

##### Delete Module

The delete command allows you to delete a module from the registry.

```bash
pbuf modules delete [module_name]
```

The command deletes all the tags and proto files associated with the module.

##### Get Metadata

The metadata command allows you to get parsed metadata (packages) for a module tag.

```bash
pbuf metadata get [module_name] [tag]
```

Replace `[module_name]` with the name of the module. Replace `[tag]` with the tag to get metadata for.

> If `module_name` is not provided, the `name` from `pbuf.yaml` is used.

#### Users / Bots

The `users` command group allows you to manage users, bots, and permissions.

##### Create User or Bot

```bash
pbuf users create [name] --type user|bot
```

Replace `[name]` with the user or bot name. The `--type` flag defaults to `user`.

##### List Users

```bash
pbuf users list [--page-size 50] [--page 0]
```

##### Get User

```bash
pbuf users get [id]
```

##### Update User

```bash
pbuf users update [id] [--name new_name] [--active] [--inactive]
```

Only one of `--active` or `--inactive` can be set at a time.

##### Delete User

```bash
pbuf users delete [id]
```

##### Regenerate Token

```bash
pbuf users regenerate-token [id]
```

##### Grant Permission

```bash
pbuf users grant-permission [user_id] [module_name] --permission read|write|admin
```

##### Revoke Permission

```bash
pbuf users revoke-permission [user_id] [module_name]
```

##### List Permissions

```bash
pbuf users list-permissions [user_id]
```

#### Drift Detection

The `drift` command group allows you to manage drift detection events.

##### List Drift Events

```bash
pbuf drift list [--unacknowledged-only]
```

The `--unacknowledged-only` flag defaults to `true` and filters to only unacknowledged events.

##### Get Module Drift Events

```bash
pbuf drift module [module_name] [--tag tag_name]
```

Replace `[module_name]` with the name of the module. Use the optional `--tag` flag to filter by tag name.

##### Get Module Dependency Drift Status

```bash
pbuf drift dependencies [module_name] [--tag tag_name]
```

Replace `[module_name]` with the name of the module. Use the optional `--tag` flag to evaluate dependency drift for a specific module tag.

---

### Configuration (`pbuf.yaml`)

A typical `pbuf.yaml` file contains the following:

```yaml
version: v1
name: [module_name]
registry:
  addr: [registry_url]
  insecure: false # set to true only for local dev with HTTP; use false (default) for HTTPS in production
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
- `[proto_files_path]`: One or several paths that contain `.proto` files.

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
  addr: pbuf.cloud
  insecure: false # pbuf.cloud uses HTTPS; set to true only for local HTTP registries
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
   addr: pbuf.cloud
   insecure: false # use HTTPS for production registries
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

### Security & Authentication

#### .netrc Authentication

`pbuf` uses the `.netrc` file for authentication with registries and private Git repositories. Ensure your `.netrc` file is properly configured in your home directory (`~/.netrc`).

For PowerBuf Registry with static token authentication:
```
machine <registry_host>
token <token>
```

Example:
```
machine pbuf.cloud
token 1234567890
```

For private Git repositories:
```
machine github.com
login <username>
password <personal_access_token>
```

#### CI/CD Integration

**GitHub Actions:**
```yaml
- name: Setup .netrc
  run: |
    echo "machine pbuf.cloud" > ~/.netrc
    echo "token ${{ secrets.PBUF_TOKEN }}" >> ~/.netrc
    chmod 600 ~/.netrc

- name: Vendor protos
  run: pbuf vendor
```

**GitLab CI:**
```yaml
before_script:
  - echo "machine pbuf.cloud" > ~/.netrc
  - echo "token ${PBUF_TOKEN}" >> ~/.netrc
  - chmod 600 ~/.netrc

vendor:
  script:
    - pbuf vendor
```

#### Security Best Practices

- **Never commit** `.netrc` files to version control
- Use **environment variables** or **secrets management** in CI/CD
- Set proper file permissions: `chmod 600 ~/.netrc`
- For production registries, always use **HTTPS** (`insecure: false`)
- Rotate tokens regularly

### Contribution

If you'd like to contribute to `pbuf`, feel free to fork the repository and send us a pull request!

---

We hope you find `pbuf` useful for your projects!
