# configmgr

### Introduction

`configmgr` is a utility for applying defined configuration parameters against resources on your system. It takes the path to a YAML-based configuration file as its sole argument, and outputs a YAML-formatted report detailing its activities and their results.

### Resource Types

#### Files (`File`)

The utility manages files using the following fields:

- `Path`: A string containing the path of the target file on disk.
- `Owner`: The UID of the user who should own the file as an integer.
- `Group`: The GID of the group that the file should belong to as an integer.
- `Mode`: The umask of the permissions that should be applied to the file.
- `Create`: A boolean flag indicating whether to create the file if it does not exist.
- `Directory`: A boolean flag indicating whether the path in question is a directory, rather than a file.
- `Content`: An array of byte literals comprising the desired contents of the file. 


#### Debian Packages (`Deb`)

The utility manages Debian packages using the following fields:

- `Name`: A string containing the name of the package to be targeted.
- `Install`: A boolean indicating whether to install the package if it is not present (incompatible with `Remove`).
- `Remove`: A boolean indicating whether to remove the package if it is present (incompatible with `Install` and `Upgrade`).
- `Upgrade`: A boolean indicating whether to upgrade the package to a newer version if one is available (incompatible with `Remove`).

#### System Services (`Service`)

The utility manages system services using the following fields:

- `Name`: A string containing the name of the service to be targeted.
- `Running`: A boolean indicating whether the service should be started. If set to `true`, the utility will always attempt to start the targeted service. If set to `false`, it will always attempt to stop the targeted service.
- `Restart`: (Requires `Running` to be `true`) A boolean indicating whether to attempt to restart the targeted service each time the utility runs.

### Config File Structure

As mentioned above, all configuration is composed in YAML. An example file can be found at example.yaml

