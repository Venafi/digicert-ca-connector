[![Venafi](https://raw.githubusercontent.com/Venafi/.github/master/images/Venafi_logo.png)](https://www.venafi.com/)
[![MPL 2.0 License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](https://opensource.org/licenses/MPL-2.0)

# digicert-ca-connector
Sample TLSPC Certificate Authority connector

# Dependencies
Below are the minimal dependency version that are required to build ....
- GNU Make 3.81
- jq - commandline JSON processor [version 1.6]
- go version go1.20
- Docker version 24.0.7
- golangci-lint has version 1.52.2

# Setting up environment variables
To build an image that will be run within a Venafi Satellite for certificate issuance and/or import operations you will need to define a CONTAINER_REGISTRY environment variable.

```bash
export CONTAINER_REGISTRY=company.jfrog.io/ca-connectors/digicert
```

> **_NOTE:_** The image, push and manifests make targets will fail if no CONTAINER_REGISTRY value is set.

> **_NOTE:_** Venafi developer documentation is available at ?

# Venafi Satellite

A Venafi Satellite is a customer hosted infrastructure component designed to provide certain functionality within the
customers network such as private network discovery and integrations among other services. The Venafi Satellite provides
private services within a customer's network by making communications from the customer network to Venafi as a Service,
requiring only outbound communication and connectivity to managed endpoints.

To support integrations with other systems, such as a DigiCert Certificate Authority, developers can create a
Certificate Authority (CA) connector to perform predefined functions. A CA connector is a plugin that acts as a
middleware to communicate between the Venafi Platform to any 3rd party Certificate Authority. The CA connector is
responsible for authenticating, pulling available product (issuance) and import options, requesting and retrieving
certificates from the client's chosen CA to Venafi.

In the Venafi world, every CA connector is a REST-based web service with certain predefined APIs implemented to perform
a task. Venafi provides a CA connector framework to help build more CA connectors easily to support the clients'
ecosystem. A CA connector is composed of three parts.

- An executable that is run within a container. The executable uses a web framework to receive incoming requests from a
  service within the Venafi Satellite. The request is processed and the response is returned to the internal service
  which then sends the result to VaaS.
- A manifest that defines a series of data structures that are used during different operations.
  - The manifest MUST define the properties required for connecting to the Certificate Authority
  - In case of developing a ca connector for certificate issuance, the manifest MUST define the properties required for
    submitting a certificate request, as well as retrieving issued certificate and chain.
  - In case of developing a ca connector for certificate import in TLS Protect Cloud, the manifest MUST define properties specific
    to the import configuration and option that will be needed for the retrieval of available certificates.
- A container image that is compatible with the executable. It is strongly recommended that the image also contain the
  manifest.json so that if a change to the manifest is made, but not to the executable code, then the container image
  SHA256 digest will also be changed.

Additional resources for developing a CA connector are available at [Venafi Dev Central](https://developer.venafi.com/tlsprotectcloud/docs/libraries-and-sdks-ca-connector-framework)

# Manifest

The manifest.json file contains the definitions for connection, issuance and import operations. These definitions are
also used in the Venafi as a Service UI for using the CA connector.
As data is exchanged with the CA connector that data is validated against the manifest.json file. The only field names
and values permitted are those defined within the file.

The manifest is a JSON document with a defined structure and required nodes.  The top level node must the following fields:
- ___name___: the required name of the CA connector such as "DigiCert CA Connector".  This value is shown in the TLS Protect Cloud UI.
- ___pluginType___: a required field for a CA connector and the value must be "CA"
- ___workTypes___: a required collection of strings indicating the capabilities of the CA connector. Supported values
  are ["CERTIFICATE_IMPORT"], ["ISSUANCE"], and ["CERTIFICATE_IMPORT", "ISSUANCE"]. Work types value actually controls
  supported operations for the connector - issuance, import or issuance and import.

Additionally the top level node should contain the following:
- ___deployment___: a required node that contains the image location that will be used by the Venafi Satellite to pull the container.
  - ___executionTarget___: a required field the value "vsat".
  - ___image___: the required container registry and image path used to pull the container image.
- ___domainSchema___: a required node that contains definitions for connection, issuance and/or import operations.

## The connection property
The data required to test connectivity with Certificate Authority can be defined in the connection node of the domainSchema node.

- ___connection___: a node within the domainSchema node that defines the properties needed to perform a connection test against the Certificate Authority.
  - _type_: a required field which has a value of "object".
  - _properties_: a collection of nodes defining the property values needed to perform a connection to the Certificate Authority.
    - _configuration_: a node within the domainSchema node that defines the configuration properties needed to perform a connection test against the Certificate Authority.
      - _type_: a required field which has a value of "object".
      - _properties_: a collection of nodes defining the configuration property values needed to perform a connection to the Certificate Authority. For example, "Server URL", "Port", etc.
      - _required_: an optional collection of property names indicating that a value is required for the corresponding configuration property.
    - _credentials_: a node within the domainSchema node that defines the credentials properties needed to perform a connection test against the Certificate Authority.
      - _type_: a required field which has a value of "object". 
      - _properties_: a collection of nodes defining the credentials property values needed to perform a connection to the Certificate Authority. For example, "API Key", "Username", "Password", "Client Certificate", etc.
      - _required_: an optional collection of property names indicating that a value is required for the corresponding credentials property.
    - _required_: an optional collection of property names indicating that a value is required for the corresponding property.

In this sample machine connector the connection definition is:

```json
{
  "connection": {
    "type": "object",
    "properties": {
      "configuration": {
        "type": "object",
        "properties": {
          "serverUrl": {
            "type": "string",
            "x-labelLocalizationKey": "serverUrl.label",
            "x-rank": 0
          }
        },
        "required": [
          "serverUrl"
        ]
      },
      "credentials": {
        "type": "object",
        "properties": {
          "apiKey": {
            "type": "string",
            "description": "apiKey.description",
            "x-encrypted": true,
            "x-labelLocalizationKey": "apiKey.label",
            "x-rank": 1,
            "x-controlOptions": {
              "password": "true",
              "hidePasswordLabel": "password.hidePassword",
              "showPasswordLabel": "password.showPassword"
            }
          }
        },
        "required": [
          "apiKey"
        ]
      }
    },
    "required": [
      "configuration",
      "credentials"
    ]
  }
}
  ```

## The productOption property
The data required to specific product(issuance) options with Certificate Authority can be defined in the productOption node of the domainSchema node.

- ___productOption___: a node within the domainSchema node that defines the specific product(issuance) properties needed to submit a certificate request to the Certificate Authority.
  - _type_: a required field which has a value of "object".
  - _properties_: a collection of nodes defining the property values needed to request a certificate from the Certificate Authority.
    - _name_: a required field displaying the product option name.
    - _types_: a required field displaying the product option certificate types. Supported values are ["SSL"], ["CODE_SIGN"], and ["SSL", "CODE_SIGN"].
    - _productDetails_: an optional properties describing specific/available product details for that product option that will be displayed in the UI.
  - _required_: an optional collection of property names indicating that a value is required for the corresponding property. 

The following is an example of this section of the manifest:

```json
 {
  "productOption": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string"
      },
      "types": {
        "type": "array",
        "items": {
          "type": "string",
          "enum": [
            "SSL",
            "CODE_SIGN"
          ],
          "default": "SSL"
        }
      },
      "productDetails": {
        "type": "objects",
        "properties": {
          "hashAlgorithms": {
            "type": "array",
            "items": {
              "type": "string"
            }
          },
          "nameId": {
            "type": "string"
          },
          "certificateType": {
            "type": "string"
          },
          "allowAutoRenew": {
            "type": "boolean"
          },
          "organizationIds": {
            "type": "array",
            "items": {
              "type": "integer"
            }
          }
        }
      }
    },
    "required": [
      "name",
      "types"
    ]
  }
}
 ```

## The product property
The data required to specific product(issuance) properties needed for certificate issuance, can be defined in the product node of the domainSchema node.

- ___product___: a node within the domainSchema node that defines the specific product(issuance) properties needed to submit a certificate request to the Certificate Authority.
  - _type_: a required field which has a value of "object".
  - _properties_: a collection of nodes defining the property values needed to request a certificate from the Certificate Authority. Using x-dynamic-values per specific property and referencing specific productDetails domain schema, will describe the available values for selection in the UI
  - _required_: an optional collection of property names indicating that a value is required for the corresponding property.

The following is an example of this section of the manifest:

```json
 {
  "product": {
    "type": "object",
    "properties": {
      "organizationId": {
        "type": "integer",
        "x-labelLocalizationKey": "organizationId.label",
        "x-rank": 0
      },
      "hashAlgorithm": {
        "type": "string",
        "x-dynamic-values": "$.hashAlgorithms",
        "x-labelLocalizationKey": "hashAlgorithm.label",
        "x-rank": 1
      },
      "nameId": {
        "type": "string",
        "x-dynamic-values": "$.nameId",
        "x-labelLocalizationKey": "nameId.label",
        "x-rank": 2
      }
    }
  }
}
 ```

## The localizationResources property

The TLS Protect Cloud user interface for a CA connector is dynamically rendered using the definitions within the
manifest. The property definitions in the domainSchema node are evaluated as the user interface is rendered. Property
labels and descriptions are mapped from the property definition using the x-labelLocalizationKey field where the value
contains a dotted path to within a language in the localizationResources node.

- ___localizationResources___: a required top level node containing the text shown in the TLS Protect Cloud UI.
  - ___en___: a required node containing the English language localization values. The definitions contained within this
    node are mapped to the x-labelLocalizationKey fields defined on properties in the manifest.
    - ___address___: a node containing the localization fields for properties having an x-labelLocalizationKey values
      beginning with address.
      - ___label___: the value for properties having an x-labelLocalizationKey value of "address.label".
      - ___description___: the description for the specific property and displays helper text (hints) in the UI

The following is an example of this section of the manifest:

```json
{
  "localizationResources": {
    "en": {
      "serverUrl": {
        "label": "Server URL"
      },
      "apiKey": {
        "label": "API Key",
        "description": "API key used to authenticate against the Certificate Authority"
      },
      "organizationId": {
        "label": "Organizations"
      },
      "hashAlgorithm": {
        "label": "Signature Hash"
      },
      "nameId": {
        "label": "Name ID"
      },
      "autoRenew": {
        "label": "Allow Auto Renew"
      }
    }
  }
}
```

# Routes

All CA connectors must support at least one of issuance or import operations or both of them. These operations include
testing access to the Certificate Auhtority(e.g. testConnection), retrieving product(issuance) and import options(e.g. getOptions), validating specific
product attributes against Certificate Authority(e.g. validateProduct), requesting certificate(e.g. requestCertificate), checking order details for submitted
certificate request(e.g. checkOrder), downloading issued certificate once available and searching for certificates matching import
configuration that will be available for import in TLS Protect Cloud server and configuring usage of an installed
certificate (e.g. configureInstallationEndpoint).

- ___hooks___: a required top level node containing the mapping and requestConverters nodes.
  - ___mapping___: a required node defining the operations and the corresponding REST URL path.  Each of the required sub-nodes _MUST HAVE_ a path definition containing the REST URL path to be used to execute the corresponding operation.

    -
      | Hook node          | Description                                                                                                              | Required/Optional For Issuance | Required/Optional For Import |
      |:-------------------|:-------------------------------------------------------------------------------------------------------------------------|:-------------------------------|:-----------------------------|
      | testConnection     | testing connection to Certificate Authority                                                                              | Required                       | Required                     |
      | getOptions         | retrieving product(issuance) and/or import options based on supported operations                                         | Required(product options)      | Required(import options)     |
      | validateProduct    | validating specific product(issuance) attributes                                                                         | Required                       | N/A                          |
      | requestCertificate | submitting certificate request to Certificate Authority                                                                  | Required                       | N/A                          |
      | checkOrder         | checking order details for submitted certificate request                                                                 | Optional                       | N/A                          |
      | checkCertificate   | checking certificate details for submitted certificate request                                                           | Optional                       | N/A                          |
      | importCertificates | retrieving certificates matching specific import configuration, that will be available for import in TLS Protect Cloud.  | N/A                            | Required                     |

      **Table reference information**
          - **Required** - must be implemented as part of connector implementation
          - **Optional** - depends on the connector implementation(for example, if issued certificate is returned as part of requestCertificate response, checkOrder and checkCertificate implementation are not needed. Otherwise, some of them, or both of them can be needed in order to download issued certificate from the CA.
          - **N/A** - not applicable for this work type(for example, some hooks are needed for issuance work type, so if connector for import only is being developed they don't need to be implemented)
  - ___requestConverters___: a required array of named converters.  If any manifest property has an x-encrypted field with a value of true then the collection must contain the value of "arguments-decrypter".


# Issuing Certificates Connector Basics

## Manifest

## Testing Access

# Importing Certificates Connector Basics

# Code Structure

# Building

## Binary

## Container Image

# Testing
