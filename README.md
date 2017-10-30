# Azure DNS sync

Because currently Azure's DNS servers are limited maybe there is need
for a sync of some hostnames from an internal DNS server into an public
Azure DNS zone.


## Setup

Create azure config secret first:
```
kubectl create secret generic azure-config-file --from-file=/etc/kubernetes/azure.json

```

Create deployment:
```
apiVersion: v1
kind: ConfigMap
metadata:
   Name: azure-dns-sync-config
data:
  config.yml: |
    ---
    
    default:
      resourceGroup: your-dns-resource-group
      zone: your-azure.zone
      ttl: 60
    
    entries:
      - name: foo.example.com
        azure:
          name: example-barfoo
      - name: foo.example.com
        azure:
          name: example-foobar
          resourceGroup: other-dns-resource-group
          zone: other-azure.zone
          ttl: 120
        dns:
          - 8.8.4.4
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: azure-dns-sync
spec:
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: azure-dns-sync
    spec:
      containers:
      - name: azure-dns-sync
        image: mblaschke/azure-dns-sync
        volumeMounts:
        - name: azure-config-file
          mountPath: /etc/kubernetes
          readOnly: true
        - name: azure-dns-sync-config-volume
          mountPath: /etc/azure-dns-sync
          readOnly: true
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file
      - name: azure-dns-sync-config-volume
        configMap:
          name: azure-dns-sync-config
```

Run deployment
```
kubectl apply -f deployment.yml
```
