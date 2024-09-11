# grafana

This directory contains an example configuration for Grafana 9 to be used with the in cluster monitor metrics.

There is currently only one variant "ocp-oauth" which configures Grafana to use the OpenShift OAuth service for authentication.

## ocp-oauth Overlay

Following are the configuration values that need to be set based on the cluster:

- `CLUSTER_ROUTES_BASE` in `patch_deployment_oauth.yaml`, based on the DNS name of the cluster

For example, your cluster domain (the part after `.apps.` in routes) is `my-cluster.devcluster.openshift.com` then the following configuration values would be correct:

```plain
CLUSTER_ROUTES_BASE=my-cluster.devcluster.openshift.com
```

Once these configuration values have been set in the files, the configuration can be applied using:

```bash
oc apply -k overlays/ocp-oauth/
```

The configuration creates a `grafana` deployment and route in the `openshift-monitoring` namespace. Your Grafana instance will be reachable using `https://grafana-openshift-monitoring.apps.${CLUSTER_ROUTES_BASE}/`.
