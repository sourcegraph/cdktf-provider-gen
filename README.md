# cdktf-provider-gen

Terraform CDK providers generator for Go.

## Background

There are several supported ways of consuming Terraform providers or moduels in CDKTF:

> This project is explicitly not tracking the Terraform google Provider version 1:1. In fact, it always tracks latest of ~> 4.0 with every release. If there are scenarios where you explicitly have to pin your provider version, you can do so by generating the provider constructs manually.

[Pre-built providers] works out-of-the-box, but they lack the ability to track specific upstream provider version to meet your organization need. 
Terraform providers are usually not strictly following semver and could introduce breaking changes in minor or patch versions.

On the other hand, you can use the `cdktf-cli` and run `cdktf get` to generate the provider constructs.
This command lacks the ability to cache the generated constructs and you have to re-generate all providers all the time. Having many providers in the project could significantly increase the time it takes to generate the constructs. 

In Go, it does not provide an easy way to make individual provider a separate Go module. This is problematic when you would like to avoid comitting generated codes to your main repository and host them in a separate centralized Git repository for ease of consumption. Then, you would have to make each provider a separate module because providers such as `google` and `aws` are too big and exceed the limit of Go modules proxy.

It's not really sustainable to replicate Hashicorp's own infra to publish every [pre-build providers] using [cdktf/cdktf-provider-project] when we just want something simple.

Therefore, we reverse engineer how [cdktf/cdktf-provider-google] generates the standalone Go module [cdktf/cdktf-provider-google-go] and created this project for our use case.

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
# Optionally use -cdktf-version to specify the version of cdktf to use
cdktf-provider-gen -config google.yml
```

Finally, you will have a Go module created at `gen/google`. Once you push your changes to remote, you can import it with:

```sh
go get github.com/your-org/cdktf-providers/gen/google
```

[pre-built providers]: https://developer.hashicorp.com/terraform/cdktf/concepts/providers#install-pre-built-providerss
[cdktf/cdktf-provider-google]: https://github.com/cdktf/cdktf-provider-google
[cdktf/cdktf-provider-google-go]: https://github.com/cdktf/cdktf-provider-google-go
[cdktf/cdktf-provider-project]: https://github.com/cdktf/cdktf-provider-project