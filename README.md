# http-source-controller

Defines an `Http` source for the OpenFluxCD ecosystem.

This source is capable of producing an [Artifact](https://github.com/openfluxcd/artifact) object that various other
controllers can understand and consume.

## Basic Scenario

This controller consumes a simple object for which the definition is such:

```yaml
apiVersion: openfluxcd.openfluxcd/v1alpha1
kind: Http
metadata:
  labels:
    app.kubernetes.io/name: http-source-controller
    app.kubernetes.io/managed-by: kustomize
  name: http-sample
  namespace: http-source-controller-system
spec:
  url: "https://raw.githubusercontent.com/openfluxcd/controller-manager/main/README.md"
```

It will fetch whatever the URL is pointing to and create an Artifact that then can be used to get the content.
The Artifact will be provided by a file server for which the URL will be located in the status such as:

```yaml
Name:         http-http-source-controller-system-http-sample
Namespace:    http-source-controller-system
Labels:       <none>
Annotations:  <none>
API Version:  openfluxcd.mandelsoft.org/v1alpha1
Kind:         Artifact
Metadata:
  Creation Timestamp:  2024-07-26T09:33:53Z
  Generation:          2
  Resource Version:    1242
  UID:                 4e7d8da1-27a8-4cbd-8878-eed00dd7bba9
Spec:
  Digest:            sha256:ec3f9489eb75419e0f9973c1667d949e1617961e5d17b62803467bceef137cac
  Last Update Time:  2024-07-26T09:34:59Z
  Revision:          18ad60b599b0c85dad0ceb4f50c7873c19bd050da832e22042b71c944aeaa315
  Size:              144
  URL:               http://http-source-controller.http-source-controller.svc.cluster.local./http/http-source-controller-system/http-sample/18ad60b599b0c85dad0ceb4f50c7873c19bd050da832e22042b71c944aeaa315.tar.gz
Events:              <none>
```

Further controllers that can understand and watch these Artifact can then fetch the content and perform some action on
it.

The most important part here is the `Revision`. This identifies the content as unique. This could be a version, a digest,
a commit SHA, etc. `Revision` is updated whenever there is new content. This new content triggers an update to the existing
Artifact, generating a new file and an updated URL, Digest, Last Update Time and Size.

Let's see some scenarios using two controllers that understand the fetched content.

## Kustomize based scenario

[Kustomize Controller](https://github.com/openfluxcd/kustomize-controller) is one of these controllers.

For an example scenario, we are going to use this repository: [Podinfo Kustomize](https://github.com/openfluxcd/podinfo-kustomize). It contains files and an archive
for kustomization. The resulting Artifact then can be used as a source for our Kustomization object.

First, create an Http object for the tar gz in the release section:

```yaml
apiVersion: openfluxcd.openfluxcd/v1alpha1
kind: Http
metadata:
  name: http-podinfo-kustomize
  namespace: http-source-controller-system
spec:
  url: "https://github.com/Skarlso/podinfo-kustomize/releases/download/v0.1.0/podinfo.tar.gz"
```

Reconciling this object will result in an Artifact like this:

```yaml
API Version:  openfluxcd.mandelsoft.org/v1alpha1
Kind:         Artifact
Metadata:
  Creation Timestamp:  2024-08-01T07:26:22Z
  Generation:          1
  Resource Version:    1027
  UID:                 9141ec2a-db70-4419-a8cd-f568ce669e74
Spec:
  Digest:            sha256:f579566539936c1ffa7f3d1ed037de84d3afd3a79083c255aa49356fefc597b2
  Last Update Time:  2024-08-01T07:26:22Z
  Revision:          897eb933ed697314bc128e102dd8d99c7f811534b61e3d402c9bdc876dee5132
  Size:              2973
  URL:               http://http-source-controller.http-source-controller-system.svc.cluster.local./http/http-source-controller-system/http-podinfo-kustomize/897eb933ed697314bc128e102dd8d99c7f811534b61e3d402c9bdc876dee5132.tar.gz
```

Once we have this artifact, we can set that as a SourceRef for our Kustomization object like this:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo-kustomize
  namespace: http-source-controller-system
spec:
  interval: 10m
  targetNamespace: default
  sourceRef:
    kind: Artifact
    name: http-http-source-controller-system-http-podinfo-kustomize
    apiVersion: openfluxcd.mandelsoft.org/v1alpha1
  path: "."
  prune: true
  timeout: 1m
```

Now, kustomize-controller can fetch the generated artifact from the http-source-controller provided
file server. If all went well, it should apply the fetched kustomize files and deploy Podinfo in
the default namespace:

```
k get pods
NAME                       READY   STATUS    RESTARTS   AGE
podinfo-8476886b4c-9br9s   1/1     Running   0          14s
```

Next, let's take a look on the same scenario but using Helm instead.

## Helm based scenario
