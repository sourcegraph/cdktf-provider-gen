# cdktf-provider-gen

CDK-Terrain provider generator for Go. CDK-Terrain is the open source successor to CDK for Terraform (CDKTF).

## Background

There are several supported ways of consuming Terraform providers or modules in CDK-Terrain:

> This project is explicitly not tracking the Terraform google Provider version 1:1. In fact, it always tracks latest of ~> 4.0 with every release. If there are scenarios where you explicitly have to pin your provider version, you can do so by generating the provider constructs manually.

[Pre-built providers] works out-of-the-box, but they lack the ability to track specific upstream provider version to meet your organization need. 
Terraform providers are usually not strictly following semver and could introduce breaking changes in minor or patch versions.

On the other hand, you can use the `cdktn-cli` and run `cdktn get` to generate the provider constructs.
This command lacks the ability to cache the generated constructs and you have to re-generate all providers all the time. Having many providers in the project could significantly increase the time it takes to generate the constructs. 

In Go, it does not provide an easy way to make individual provider a separate Go module. This is problematic when you would like to avoid comitting generated codes to your main repository and host them in a separate centralized Git repository for ease of consumption. Then, you would have to make each provider a separate module because providers such as `google` and `aws` are too big and exceed the limit of Go modules proxy.

Additionally, the upstream generator suffers from the problem of packaging duplicated [jsii modules](https://www.npmjs.com/package/jsii-pacmak) in the Go packages under the same project. For example, a project includes `aws` and `google` providers. The generated `aws` Go package will contain the jsii modules for both `aws` and `google` provider. It has the following problems:

- Bloated binary. The produced Go binary includes duplicate (`N^2` where `N` is the number of providers in the CDKTF project) jsii modules.
- Slow compliation time.
- Slow startup time. The Go program needs to load all the jsii modules during `init` time.

It's not really sustainable to replicate Hashicorp's own infra to publish every [pre-build providers] using [cdktf/cdktf-provider-project] when we just want something simple.

Therefore, this generator creates one standalone Go module for each provider or Terraform module while using CDK-Terrain's provider generator.

## Installation

We also require `node` and `npm` to be installed.

```sh
go install github.com/sourcegraph/cmd/cdktf-provider-go@main 
```

## Usage

In a GitHub repository `your-org/cdktf-providers`:

```sh
go mod init github.com/your-org/cdktf-providers
mkdir gen
```

Create the config file:

```sh
touch google.yaml
```

```yaml
name: google
provider:
  source: registry.terraform.io/hashicorp/google
  version: "4.69.1"
language: go
target:
  language: go
  moduleName: github.com/your-org/cdktf-providers/gen
  packageName: google
output: gen
```

Run the generator:

```sh
# Optionally use -cdktn-version to specify the version of CDK-Terrain to use
cdktf-provider-gen -config google.yml
```

Finally, you will have a Go module created at `gen/google`. Once you push your changes to remote, you can import it with:

```sh
go get github.com/your-org/cdktf-providers/gen/google
```

## Troubleshooting

### Broken code generation error from `node`

> [!NOTE]
> This generator packages CDK-Terrain output as standalone Go modules, so upstream generator changes can require corresponding updates here.

Use the `-keep` flag to retain the intermediate assets:

```sh
export SRC_LOG_LEVEL=debug
cdktf-provider-gen -config google.yml -keep
```

Locate the `tmpDir` from output logs, and `cd` into the directory. You will find the `node` project we use to generate provider code from. 

Then, you can manually run relevant commands to debug the issue:

```sh
npm run compile
npm run pkg:go
```

[pre-built providers]: https://github.com/open-constructs/cdk-terrain
[cdktf/cdktf-provider-google]: https://github.com/cdktf/cdktf-provider-google
[cdktf/cdktf-provider-google-go]: https://github.com/cdktf/cdktf-provider-google-go
[cdktf/cdktf-provider-project]: https://github.com/cdktf/cdktf-provider-project
