## Output:

```bash
>> go run main.go

Waiting for caches to sync...
Found 12 pods
Pod: kube-system/civo-csi-controller-0 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-cri2l

Pod: kube-system/civo-csi-node-8l8n5 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited

Pod: kube-system/coredns-6799fbcd5-57sz7 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-cri2l

Pod: kube-system/metrics-server-67c658944b-bgp7g 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-cri2l

Pod: default/httpd7 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited

Pod: default/install-traefik2-nodeport-cl-pdrpf 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited

Pod: default/nginx-deployment2-69947777ff-2lc4h 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited

Pod: kube-system/civo-csi-node-dfq9f 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-cri2l

Pod: kube-system/traefik-2mk2h 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-cri2l

Pod: kube-system/traefik-92nzx 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited

Pod: default/httpd8 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited

Pod: kube-system/civo-ccm-5474f5869d-s7fk4 
Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-cri2l


=== With Namespace Index ===
Pods in default namespace: 4
Pod details from namespace index:
  Name: httpd7, Namespace: default, Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited
  Name: httpd8, Namespace: default, Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited
  Name: install-traefik2-nodeport-cl-pdrpf, Namespace: default, Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited
  Name: nginx-deployment2-69947777ff-2lc4h, Namespace: default, Node: k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited

=== With Node Index ===
Pods on node k3s-cloudterms-k8s-1486-8a8686-node-pool-c68e-kited: 6
  Name: traefik-92nzx, Namespace: kube-system
  Name: httpd7, Namespace: default
  Name: httpd8, Namespace: default
  Name: install-traefik2-nodeport-cl-pdrpf, Namespace: default
  Name: nginx-deployment2-69947777ff-2lc4h, Namespace: default
  Name: civo-csi-node-8l8n5, Namespace: kube-system
```