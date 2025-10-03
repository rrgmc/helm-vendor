# helm-vendor

`helm-vendor` is a command-line tool to manage vendoring of Kubernetes Helm Charts.

Features:
- config-based list of charts to manage.
- commands to download and upgrade charts automatically.
- check for newer versions of installed charts.
- when upgrading charts, the chart of the current version is downloaded, and only the files contained in it are deleted
  locally, before unpacking the new version. This ensures any new file added manually will be kept.
- Optionally a diff can be made of any local changes in relation to the original chart, and the patch applied on 
  the new version during upgrade. 

## Install

Get an executable from the [releases](https://github.com/rrgmc/helm-vendor/releases) or if you have a Go 1.25 compiler
available:

```shell
go get github.com/rrgmc/helm-vendor
```

## Example

**helm-vendor.yaml**:

```yaml
charts:
  - path: opentelemetry-collector
    repository:
      url: https://open-telemetry.github.io/opentelemetry-helm-charts
    name: opentelemetry-collector
    files:
      ignore:
        - "ci/**"
        - "examples/**"
  - path: datadog
    repository:
      url: https://helm.datadoghq.com
    name: datadog
  - path: argo-cd
    repository:
      url: https://argoproj.github.io/argo-helm
    name: argo-cd
    files:
      ignore:
        - "charts/redis-ha/**"
  - path: temporal-cluster
    repository:
      url: https://go.temporal.io/helm-charts/
    name: temporal
```

```shell
$ helm-vendor fetch-all
Downloading "opentelemetry-collector"...
Downloading "datadog"...
Downloading "argo-cd"...
Downloading "temporal-cluster"...

$ helm-vendor check
- opentelemetry-collector: [local:0.136.1] [latest:0.136.1]
- datadog: [local:3.135.4] [latest:3.135.4]
- argo-cd: [local:8.5.8] [latest:8.5.8]
- temporal-cluster: [local:0.67.0] [latest:0.67.0]

$ helm-vendor check opentelemetry-collector
opentelemetry-collector:
- description: OpenTelemetry Collector Helm chart for Kubernetes
- local: 0.136.1
- latest: 0.136.1
- versions:
	- 0.136.1 [2025-09-26]
	- 0.136.0 [2025-09-26]
	- 0.135.1 [2025-09-25]
	- 0.135.0 [2025-09-25]
	- 0.134.1 [2025-09-23]
	- 0.134.0 [2025-09-15]
	- 0.133.1 [2025-09-15]
	- 0.133.0 [2025-09-08]
	- 0.132.0 [2025-08-27]
	- 0.131.0 [2025-08-20]
	
$ helm-vendor upgrade opentelemetry-collector
# Would upgrade opentelemetry-collector to the latest version.
```

#### Upgrade process

Upgrading the `opentelemetry-collector` version from the local one `0.133.1` to latest `0.136.1`:

- download the `0.133.1` version.
- create a `diff` of the files contained in this chart version with the current local files, and write it to the root folder.
- delete all files from the local folder which are contained in this chart version. This ensures any custom file is kept.
- download the `0.136.1` version.
- copy all files contained in this chart version to the output folder, respecting the `ignore` configuration.
- if `apply-patch=true` is set, the `diff` generated above is applied to the new chart version.

## Author

Rangel Reale (rangelreale@gmail.com)
