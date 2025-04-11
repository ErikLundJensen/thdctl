# Talos Hetzner Dedicated Control CLI - thdctl

## Overview

`thdctl` is a command-line tool to install Sidero Labs Talos at Hetzner Dedicated servers.

## Build and run

```sh
make build
```

Run the CLI to get commands and arguments.
```
./thdctl --help
```


## Docker based

```sh
make docker-build
```

```
docker run --rm -v $(pwd):/root thdctl:latest /app/thdctl --help
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

The `init` command install Talos at a clean server.  
The `reconcile` command uses a server specification and reconcile the given specification. 

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

By default the disk /dev/sda is used for installation. Use the `--disk` option to specify another disk. 
(Previous version defaulted to a nvme disk, however, nvme disk are often used for user data and the slower SATA disk used for OS)

```sh
thdctl init 123456 --disk sda
```
The available disks are listed if the given disk is not found. Thereby it should be easier to select the correct disk in the second attempt. 


#### `reconcile`

Example using the reconcile command: 

```sh
thdctl reconcile -f talos/serverSpec.yaml
```


#### Flags & Defaults

```sh
Usage:
  thdctl [command]

Available Commands:
  completion        Generate the autocompletion script for the specified shell
  getServer         Get server details
  help              Help about any command
  init              Initialize the application
  listFirewallRules List all firewall rules for a server
  listServers       List all servers
  reconcile         Reconcile server configuration from file

Flags:
      --debug        enable debug logging
  -h, --help         help for thdctl
      --log string   set log format (txt|json) (default "txt")
```

The environment variable "HETZNET_SSH_PASSWORD" can be used if Hetzner Rescue API no longer returns the password. For example, when activating the rescue mode then the password is only available until the server reboots. If the CLI stops while the server is rebooting then the password must be set as environment variable.

## Example Workflow

1. Initialize the server:

    ```sh
    thdctl init 123456
    ```

The remaning steps are regular Talos initialization.  

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

