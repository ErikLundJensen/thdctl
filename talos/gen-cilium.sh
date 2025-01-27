helm template                                                   \
    cilium                                                      \
    cilium/cilium                                               \
    -f cilium-values.yaml                                       \
    --version 1.16.3                                            \
    --namespace kube-system                                     \
    --set ipam.mode=kubernetes                                  \
    --set kubeProxyReplacement=true                             \
    --set securityContext.capabilities.ciliumAgent="{CHOWN,KILL,NET_ADMIN,NET_RAW,IPC_LOCK,SYS_ADMIN,SYS_RESOURCE,DAC_OVERRIDE,FOWNER,SETGID,SETUID}" \
    --set securityContext.capabilities.cleanCiliumState="{NET_ADMIN,SYS_ADMIN,SYS_RESOURCE}" \
    --set cgroup.autoMount.enabled=false                        \
    --set cgroup.hostRoot=/sys/fs/cgroup                        \
    --set k8sServiceHost="${KUBERNETES_API_SERVER_ADDRESS}"     \
    --set k8sServicePort="${KUBERNETES_API_SERVER_PORT}"        > gen/cilium.yaml


# Optimizations
#    --set routingMode=native                                    \
#    --set bpf.datapathMode=netkit                               \
#    --set bandwidthManager.enabled=true                         \
#    --set bandwidthManager.bbr=true                             \

    #--set encryption.enabled=true                               \
    #--set hostFirewall.enabled=true                             \
    #--set encryption.type=wireguard                             \
