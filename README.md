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

For example, let's consider a basic controller's [install.yaml](https://github.com/Skarlso/crd-bootstrap/releases/download/v0.8.0/install.yaml) as an example.

## Helm based scenario
