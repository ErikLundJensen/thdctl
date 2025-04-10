#!/bin/bash

set -e

# TODO: change node name and output file name
# TODO: error handling...

mkdir -p gen

# TODO: Only during bootstrapping of cluster:
talosctl gen secrets --output-file gen/secrets.yaml
talosctl gen config --with-secrets gen/secrets.yaml --output-types talosconfig -o gen/talosconfig ${CLUSTER_NAME} https://${KUBERNETES_API_SERVER_ADDRESS}:${KUBERNETES_API_SERVER_PORT}

cp all-nodes.yaml gen/all-nodes.yaml

yq -i '.machine.install.disk = "/dev/sda" | .machine.network.hostname = "node1"' gen/all-nodes.yaml

talosctl gen config \
        --output gen/c1.yaml                                      \
        --output-types controlplane                               \
        --with-cluster-discovery=false                            \
        --with-secrets gen/secrets.yaml                           \
        --config-patch @cluster.yaml                              \
        --config-patch @gen/all-nodes.yaml                        \
        --kubernetes-version $KUBERNETES_VERSION                  \
        $CLUSTER_NAME                                             \
        https://${KUBERNETES_API_SERVER_ADDRESS}:${KUBERNETES_API_SERVER_PORT}

# TODO: loop through all nodes
yq -i '.machine.install.disk = "/dev/sda" | .machine.network.hostname = "node2"' gen/all-nodes.yaml

talosctl gen config \
        --output gen/c2.yaml                                      \
        --output-types controlplane                               \
        --with-cluster-discovery=false                            \
        --with-secrets gen/secrets.yaml                           \
        --config-patch @cluster.yaml                              \
        --config-patch @gen/all-nodes.yaml                        \
        --kubernetes-version $KUBERNETES_VERSION                  \
        $CLUSTER_NAME                                             \
        https://${KUBERNETES_API_SERVER_ADDRESS}:${KUBERNETES_API_SERVER_PORT}

#talosctl -n ${NODE_01_IP} -e ${NODE_01_IP}  apply-config -f gen/c1.yaml --insecure
#talosctl -n  ${NODE_02_IP} -e ${NODE_02_IP} apply-config -f gen/c2.yaml --insecure

#TODO: only at first node
#./gen-cilium.sh
#kubectl apply -f gen/cilium.yaml 