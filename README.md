# Talos Hetzner Dedicated Control CLI - thdctl

## Overview

`thdctl` is a command-line tool to manage Hetzner dedicated servers with Talos. It provides various commands to initialize, configure, and manage your servers.

## Docker based

To build and run`thdctl` use the provided Dockerfile:

```sh
make docker-build

docker run --rm -v $(pwd):/root thdctl:latest /app/thdctl --help
```

## Build and run without Docker

```sh
make build
```

## Usage

Use `thdctl --help` to get a list of available commands and arguments.  
Username and password for the Hetzner Robot API must be set using environment variables:
```
export HETZNER_USERNAME='myAPIuser'
export HETZNER_PASSWORD='password'
```

There are two ways of installing Talos using this CLI:  

* init
* reconcile

The init command install Talos at a clean server.  
The reconcile command uses a server specification and reconcile the given specification.  

The later command is intended for a crossplane provider, however, it can be used from command line as well.  

### Commands

#### `init`

Initialize Hetzner dedicated server by using a Hetzner server number.

```sh
thdctl init <serverNumber>
```

Example:

```sh
thdctl init 123456
```

### `reconsile`

Example using the reconcile command: 

```sh
thdctl reconcile -f talos/serverSpec.yaml
```


### Flags & Defaults

- `--help`: Show help information for `thdctl` commands.
- `--version`: Show the version of `thdctl`.

The environment variable "HETZNET_SSH_PASSWORD" can be used if Hetzner Rescue API no longer returns the password. For example, when activating the rescue mode then the password is only available until the server reboots.


## Example Workflow

1. Initialize the server:

    ```sh
    thdctl init 123456
    ```

The remaning steps is regular Talos initialization. Below is just an overall description.  

2. Wait for the API server to be ready, then apply the configuration:

    ```sh
    cd talos
    . ./init-env-sh
    ./generate-config.sh
    ```

    Apply talos config:

    ```sh
    talosctl -n ${NODE_01_IP} -e ${NODE_01_IP}  apply-config -f gen/c1.yaml --insecure
    ```

3. Wait for "waiting for bootstrap" and then bootstrap Talos:

    ```sh
    talosctl bootstrap
    ```

4. Get Kubernetes configuration
    ```sh
    talosctl kubeconfig -f ./gen/kubeconfig
    export KUBECONFIG=$(pwd)/gen/kubeconfig
    ```

5. Apply the Cilium configuration:

    ```sh
    ./gen-cilium.sh
    kubectl apply -f gen/cilium.yaml
    ```

6. Reboot the servers:

    ```sh
    talosctl reboot
    ```

7. Wait for the nodes to be ready and open the Talos dashboard:

    ```sh
    talosctl dashboard
    ```

8. Watch the pods get healthy:

    ```sh
    kubectl get pods -A
    ```

