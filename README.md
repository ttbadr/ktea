<p>
  <a href="https://github.com/jonas-grgt/ktea/releases"><img src="https://img.shields.io/github/release/jonas-grgt/ktea.svg" alt="Latest Release"></a>
  <a href="https://github.com/jonas-grgt/ktea/actions"><img src="https://github.com/jonas-grgt/ktea/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build Status"></a>
</p>

# 🫖 ktea - kafka terminal client

ktea is a tool designed to simplify and accelerate interactions with Kafka clusters.
![topics-page.png](topics-page.png)
![record-detail-page.png](record-detail-page.png)
![record-page.png](record-page.png)


## Installation

### Mac

```sh
brew tap jonas-grgt/ktea
brew install ktea
```

### Linux

Binaries available at the release page.

### Windows

Binaries available at the release page.

## Usage

### Config Sample
```yaml
clusters:
  - name: "my-kafka-cluster"
    color: "blue"
    active: true
    servers:
      - "localhost:9092"
    sasl:
      username: "your-sasl-username"
      password: "your-sasl-password"
      securityProtocol: "PLAIN_TEXT" # 或 "SASL_SSL"
    schema-registry:
      url: "http://localhost:8081"
      username: "your-sr-username"
      password: "your-sr-password"
    ssl-enabled: true
    tls-cert-file: "/path/to/your/tls.crt" # 可选，如果使用客户端证书
    tls-key-file: "/path/to/your/tls.key"   # 可选，如果使用客户端证书
    tls-ca-file: "/path/to/your/ca.crt"     # 可选，如果服务器证书是自签名的或非公共CA签发的
    tls-insecure-skip-verify: false # 如果为true，则跳过服务器证书验证，不推荐用于生产环境
    kafka-connect-clusters:
      - name: "my-connect-cluster-1"
        url: "http://localhost:8083"
        username: "your-connect-username" # 可选
        password: "your-connect-password" # 可选
      - name: "my-connect-cluster-2"
        url: "http://localhost:8084"

```

### Navigation

All tables can be navigated using vi like bindings:
- up: `j`
- down: `k`
- page down: `d`
- page up: `u`

### Configuration

All configuration is stored in `~/.config/ktea/config.conf`

### Certificate Conversion (JKS/PKCS12 to PEM)

`ktea` does not natively support Java Keystore (`.jks`) or PKCS12 (`.p12`) formats directly. You need to convert your certificates to PEM format using `keytool` and `openssl` before configuring them in `ktea`.

Here are the steps to convert your certificates:

1.  **Export Private Key and Certificate Chain from JKS/PKCS12 to PKCS12 format:**
    ```bash
    keytool -importkeystore -srckeystore your_keystore.jks -destkeystore temp.p12 -deststoretype PKCS12
    ```

2.  **Export Private Key to PEM format from PKCS12:**
    ```bash
    openssl pkcs12 -in temp.p12 -nocerts -nodes -out tls.key
    ```

3.  **Export Certificate Chain to PEM format from PKCS12:**
    ```bash
    openssl pkcs12 -in temp.p12 -nokeys -out tls.crt
    ```

4.  **Export CA Certificate to PEM format from JKS (Truststore):**
    ```bash
    keytool -exportcert -alias your_ca_alias -keystore your_truststore.jks -rfc -file ca.pem
    ```
    *Note: If your CA certificate is in your Keystore, use the Keystore path. If it's in a separate Truststore, use the Truststore path.*

After these steps, you will have `tls.key` (private key), `tls.crt` (certificate chain), and `ca.pem` (CA certificate) files, all in PEM format, which can be configured in the `ktea` UI.

### Cluster Management

Multiple clusters can be added.
Upon startup when no cluster is configured you will be prompted
to add one.

#### Supported Auth Methods

- No Auth
- SASL (SSL)
    - PLAIN

## Features

- *Multi-Cluster Support*: Seamlessly connect to multiple Kafka clusters and switch between them with ease.
- *Topic Management*: List, create, delete, and modify topics, including partition and offset details.
- *Record Consumption*: Consume records in text, JSON, and **Avro** formats, with powerful search capabilities.
- *Consumer Group Insights*: Monitor consumer groups, view their members, and track offsets.
- *Schema Registry Integration*: Browse, view, and register schemas effortlessly.

## Todo

- Add more authentication methods
- Add support for more message formats such as protobuf.
- Add ACL management.
- File based import/export of topics.
- Add ability to delete specific schema versions.
- Add consumption templating support.
- Many more, just file an issue requesting a feature!

## Development

### Dev cluster setup

A docker-compose setup is provided to quickly spin up a local Kafka cluster with pre-created topics, consumer groups,
commited offsets etc ...

```sh
cd docker
docker-compose up -d
```

### Generate data

After the local cluster is up and running, you can generate some data to work with, 
using`go run -tags prd ./cmd/generate`.

### Run `ktea`

Use `go run -tags dev cmd/ktea/main.go` to run `ktea` from the root of the repository.

> Note: running the tui with dev build tag will simulate an artificial slow network by sleeping for 2 seconds when doing network IO. This way the loaders and spinners can be visually asserted.
