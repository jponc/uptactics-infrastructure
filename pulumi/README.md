# Infrastructure

This Pulumi application spins up an EKS Cluster using Fargate nodes in a new VPC

# CoreDNS issue

CoreDNS by default only runs on ec2 nodes. To run it in Fargate instance we need to remove the annotation from the deployment by running:

Make sure your current-context is pointing to the correct AWS cluster

```
kubectl patch deployment coredns \
-n kube-system \
--type json \
-p='[{"op": "remove", "path": "/spec/template/metadata/annotations/eks.amazonaws.com~1compute-type"}]'
```
