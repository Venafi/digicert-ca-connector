[![Venafi](https://raw.githubusercontent.com/Venafi/.github/master/images/Venafi_logo.png)](https://www.venafi.com/)
[![MPL 2.0 License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](https://opensource.org/licenses/MPL-2.0)

# digicert-ca-connector

## Overview

Venafi provides the [DigiCert](https://www.digicert.com/) CA Connector as a reference sample to accelerate development of other CA connectors and assist developers with understanding the CA Connector Framework for [TLS Protect Cloud](https://venafi.com/tls-protect/).  DigiCert makes for a particularly good sample because of its integration complexity that requires all aspects of the CA Connector Framework.

## Code

Before you start reviewing the source code for this sample connector, please familiarize yourself with the [CA Connector Framework](https://developer.venafi.com/tlsprotectcloud/docs/libraries-and-sdks-ca-connector-framework) guide on Venafi Dev Central.  It contains a wealth of information that will provide you the background you need to be successful in developing a CA connector including a step-by-step walk through of the development process, details on how connector logic is invoked (webhooks), and a thorough "manifest" (schema) reference. 

### Prerequisites

The following software and minimum versions are required to build this sample connector and container image:
- GNU Make version 3.81
- `jq` version 1.6
- go version 1.20
- Docker version 24.0.7
- `golangci-lint` version 1.52.2

### Structure

This sample CA connector uses the following directory structure and it is strongly recommended that all CA connectors use it as their model.

```
├── build
├── cmd
│   └── digicert-ca-connector
│       └── app
└── internal
    ├── app
    │   ├── digicert-ca-connector
    │   │   └── mocks
    │   ├── domain
    │   └── service
    └── handler
        └── web
```

The source code for "main" application is located in `cmd/digicert-ca-connector` directory.  The `internal/handler` directory contains the source code for the webhook service and the `internal/app` directory contains the source code for logic invoked by the webhooks including interactions with the DigiCert CA. 

> [!NOTE] 
> The DigiCert API methods invoked by this sample connector are documented [here](https://dev.digicert.com/en/certcentral-apis/services-api.html).

## Building

To build a container image for running your connector within a Venafi Satellite, you will need to set the `CONTAINER_REGISTRY` environment variable:

```bash
export CONTAINER_REGISTRY=company.jfrog.io/tlspc
```

This sample includes a Makefile with targets for testing and building the connector.  The container image you build can be stored in your container registry and used to generate the final manifest file for deploying your connector. 

Some of the Makefile targets are:
- **help**: show available make targets
- **build**: create an executable binary that can be executed in a container running within a VSatellite.  The target operating system is Linux and the architecture will be AMD64.
- **test**: run the tests defined within the machine connector source code.
- **image**: use the `build/Dockerfile` to create a container image and stage it for the `CONTAINER_REGISTRY`.
- **push**: use the `build/Dockerfile` to create a container image and push it to the `CONTAINER_REGISTRY`.
- **manifests**: use the `manifest.json` file to create:
  - **manifest.create.json**: is an updated `manifest.json` file that includes the container registry image path.  The file content can be used to create a new machine connector for a tenant in TLS Protect Cloud.
  - **manifest.update.json**: is an updated `manifest.json` file that includes the container registry image path.  The file content can be used to update an existing machine connector for a tenant in TLS Protect Cloud.

> [!TIP]
> You can use the `TAG` environment variable to override the default container image tag value of 'latest'.

You can chain the targets together to clean, build, and push in a single command, `make clean image push`:

```
go mod download
go generate github.com/venafi/digicert-ca-connector/...
mkdir -p output/bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o output/bin/digicert-ca-connector ./cmd/digicert-ca-connector/main.go
docker --context=default buildx build --output type=image,name=company.jfrog.io/tlspc/tls-protect-digicert-ca-connector:latest,push=true --metadata-file=buildx-digest.json \
        --target image \
        --file build/Dockerfile \
         \
        --platform=linux/amd64 \
        --builder default .
[+] Building 1.7s (9/9) FINISHED                                                     docker:default
 => [internal] load .dockerignore                                                              0.0s
 => => transferring context: 2B                                                                0.0s
 => [internal] load build definition from Dockerfile                                           0.0s
 => => transferring dockerfile: 319B                                                           0.0s
 => [internal] load metadata for gcr.io/distroless/static-debian11@sha256:8ad6f3ec70da         0.3s
 => CACHED [1/3] FROM gcr.io/distroless/static-debian11@sha256:8ad6f3ec70dad966479b9fb         0.0s
 => [internal] load build context                                                              0.1s
 => => transferring context: 9.84MB                                                            0.1s
 => [2/3] COPY output/bin/digicert-ca-connector /bin                                           0.0s
 => [3/3] COPY manifest.json /bin                                                              0.0s
 => exporting to image                                                                         0.0s
 => => exporting layers                                                                        0.0s
 => => writing image sha256:47d4b42dd31585956c98febddf2e8edd65d9404aa8c5e0d16111c18a1e         0.0s
 => => naming to company.jfrog.io/tlspc/tls-protect-digicert-ca-connector:latest               0.0s
 => ERROR pushing company.jfrog.io/tlspc/tls-protect-digicert-ca-connector:latest with docker  1.3s
 => => pushing layer 19b41e526448                                                              0.8s
 => => pushing layer 8a7f5e8c4176                                                              0.8s
 => => pushing layer 5b1fa8e3e100                                                              0.8s
 ```

## Deploying

When you have completed your connector, you can deploy it exclusively in your TLS Protect Cloud production environment. With a tenant-specific connector, tenants can develop their own personal connectors (that are inaccessible by other tenants). This gives you the confidence to ensure your connectors work properly in a production environment before you release them to your customers. For details, see the [managing connectors for your tenant](https://developer.venafi.com/tlsprotectcloud/docs/integrate-connector-into-tenant-environment) guide on Venafi Dev Central.

To generate the final manifests for deployment to TLS Protect Cloud you can use ___make manifests___.  The manifests target will use the ___build___ and ___image___ targets to build the executable and the image.
