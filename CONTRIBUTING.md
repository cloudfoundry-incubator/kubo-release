# Contributing to CFCR

As a potential contributor, your changes and ideas are always welcome. Please do not hesitate to ask a question using GitHub issues or send a pull request to contribute changes.

## Contributor License Agreement
If you have not previously done so, please fill out and submit an [Individual Contributor License Agreement](https://www.cloudfoundry.org/governance/cff_individual_cla/) or a [Corporate Contributor License Agreement](https://www.cloudfoundry.org/governance/cff_corporate_cla/).

## Contributor Workflow
We encourage contributors to have discussion around design and implmentation with team members before making significant changes to the project through [GitHub Issues](https://github.com/cloudfoundry-incubator/kubo-release/issues). The project manager will prioritize where the feature will fit into the project's road map.

1. Fork the project on [GitHub](https://github.com/cloudfoundry-incubator/kubo-release)
1. Make your feature addition or bug fix. Please make sure there is appropriate [test coverage](#writing-the-tests).
1. [Run tests](#running-the-tests).
1. Make sure your fork is up to date with `develop`.
1. Send a pull request for a `develop` branch.
1. The team will triage the pull request and assign it to a team member.
1. A team member will approve the pull request or make suggestions for changes.

## Writing the tests

Unit Testing is the responsibility of all contributors.

### Unit test coverage

These confirm that a particular function behaves as intended. If the change is meant route sync package then contributor must consider adding unit tests using ginkgo testing framework and there are tests in the package to guide as examples. Unit tests code are collocated in the same package with source code.

If the changes are for BOSH templating logic then please consider adding unit tests to [BOSH templating tests](spec/).

## Running the Tests
### Pre-requisites

1. Install [Go](https://golang.org/doc/install)
1. Install [Ginkgo](https://onsi.github.io/ginkgo/)
1. Install [Ruby](https://www.ruby-lang.org/en/documentation/installation/)
1. Install Bundler

	```
	gem install bundler
	```

### Running Release Unit Tests

Execute command `./scripts/run_tests` to run unit tests for kubo-release.  These includes tests for the [BOSH templating tests](spec/).  Please run these tests before submitting a pull request.

### Integration tests

Integration tests are located in [kubo-ci](https://github.com/cloudfoundry-incubator/kubo-ci).

1. Install [Go](https://golang.org/doc/install)
1. Install [Ginkgo](https://onsi.github.io/ginkgo/)
1. Clone kubo-ci repository
1. Point kubeconfig to the cluster under test.
1. Go inside kubo-ci repository.
1. Run integration tests:

  ```
  GOPATH=$PWD ginkgo -skipPackage addons -keepGoing -r src/tests/integration-tests
  ```
  
**Note** Tests will create load balancers and persistent disks. Some tests require direct access to the kubelet nodes, but most of them just require access to Kubernetes API.

## Optional tools to deploy

### Kubo-deployment

[Kubo-deployment](https://github.com/cloudfoundry-incubator/kubo-deployment) helps users to deploy CFCR using a set of helper scripts and manifest generation tools. If your change includes deployment level changes and update to the BOSH manifest, please also update `kubo-deployment` repository. Follow the steps mentioned in project's [`CONTRIBUTING.md`](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/CONTRIBUTING.md) doc as well.

It is advisable for the Contributor to run [integration tests](https://github.com/cloudfoundry-incubator/kubo-deployment/blob/master/CONTRIBUTING.md#running-integration-tests) before submitting.
