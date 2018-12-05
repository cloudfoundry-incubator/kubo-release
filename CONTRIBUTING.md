# Contributing to CFCR

As a potential contributor, your changes and ideas are always welcome. Please do not hesitate to ask a question using GitHub issues or send a pull request to contribute changes.

## Contributor License Agreement
If you have not previously done so, please fill out and submit an [Individual Contributor License Agreement](https://www.cloudfoundry.org/governance/cff_individual_cla/) or a [Corporate Contributor License Agreement](https://www.cloudfoundry.org/governance/cff_corporate_cla/).

## Contributor Workflow
We encourage contributors to have discussion around design and implmentation with team members before making significant changes to the project through [GitHub Issues](https://github.com/cloudfoundry-incubator/kubo-release/issues). The product manager will prioritize where the feature will fit into the project's road map.

1. Fork the project on [GitHub](https://github.com/cloudfoundry-incubator/kubo-release)
1. Make your feature addition or bug fix. Please make sure there is appropriate [test coverage](#writing-the-tests).
1. [Run tests](#running-the-tests).
1. Make sure your fork is up to date with `develop`.
1. Send a pull request for the `develop` branch.
1. The team will triage the pull request and prioritise it in the team's backlog of ongoing work.
1. A team member will approve the pull request or make suggestions for changes.

## Writing the tests

### Unit test coverage

Unit tests code are collocated in the same package with source code.

If the changes are for BOSH templating logic then please consider adding unit tests to [BOSH templating tests](spec/).

## Running the Tests
### Pre-requisites

1. Install [Go](https://golang.org/doc/install)
1. Install [Ginkgo](https://onsi.github.io/ginkgo/)
1. Install [Ruby](https://www.ruby-lang.org/en/documentation/installation/)
1. Install Bundler: `gem install bundler`

### Running Kubo-Release Unit Tests

Execute command `./scripts/run_tests` to run unit tests for kubo-release.  These includes tests for the [BOSH templating tests](spec/).  Please run these tests before submitting a pull request.

### Integration tests

Integration tests are located in [kubo-ci](https://github.com/cloudfoundry-incubator/kubo-ci).

1. Install [Go](https://golang.org/doc/install)
1. Install [Ginkgo](https://onsi.github.io/ginkgo/)
1. Have a kubeconfig to your running CFCR cluster saved at `~/.kube/config`, or set `export KUBECONFIG=/path/to/your/kubeconfig`
1. Clone the [kubo-ci](https://github.com/cloudfoundry-incubator/kubo-ci) repository
1. Run
```
GOPATH="./kubo-ci" ginkgo -keepGoing -r -progress -flakeAttempts=2 -skipPackage "${skipped_packages}" "./kubo-ci/src/tests/integration-tests/"
```
Where `${skipped_packages}` is a comma-delimited string of zero, one or more of:
- `multiaz`
- `persistent_volume`
- `k8s_lbs`
- `cidrs`
- `addons`

Some tests will require additional ops-files to be applied when deploying your cluster. We’d recommend reading the [Continuous Integration configuration files](https://github.com/cloudfoundry-incubator/kubo-ci/tree/master/templates) and the integration test source code.

### KDRATs tests
KDRATs (kubo-disaster-recovery-acceptance-tests) test disaster recovery of CFCR using BOSH backup and restore (BBR).
Follow [these instructions](https://github.com/cloudfoundry-incubator/kubo-disaster-recovery-acceptance-tests/blob/master/README.md) from the kubo-disaster-recovery-acceptance-tests repository.

### Turbulence Tests
Turbulence tests are tests that introduce failure scenarios via the turbulence release to verify resilience.
1. Clone the kubo-ci repository
1. Your BOSH director must be deployed with turbulence.
  1. Please use the [turbulence ops-file](https://github.com/cloudfoundry/bosh-deployment/blob/master/turbulence.yml), and the [CFCR ops-file](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/manifests/ops-files/use-dev-turbulence.yml) which applies a development version of turbulence. This is unfortunately required until [this bug](https://github.com/cppforlife/turbulence-release/issues/22) is resolved.
  1. Upload the [turbulence runtime config](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/manifests/turbulence/runtime-config.yml) to your director, and name it turbulence. More details can be found in [this CI script](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/scripts/configure-bosh.sh).  This ensures a turbulence client is deployed on your CFCR cluster VMs.
1. Deploy CFCR
1. Have a kubeconfig to your running CFCR cluster saved at `~/.kube/config`, or set `export KUBECONFIG=/path/to/your/kubeconfig`
1. Create a [json test config file](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/scripts/generate-test-config.sh#L119-L132).  Refer to the [turbulence test script](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/scripts/run-k8s-turbulence-tests.sh) for further hints.
1. Run `CONFIG="/path/to/testconfig" ginkgo -failFast -progress -r "./kubo-ci/src/tests/turbulence-tests/"`

### Upgrade Tests
Upgrade tests determine whether BOSH can automatically upgrade from the previous released version of CFCR to the latest version.
1. Clone the kubo-ci repository
1. Deploy CFCR with the previously released version of CFCR
1. Create a bash script containing the `bosh deploy -n` instructions you just used to deploy CFCR
1. Upload the latest release to the BOSH director
1. Create a json testconfig file with [IaaS](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/scripts/generate-test-config.sh#L108), [BOSH](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/scripts/generate-test-config.sh#L112-L118), and [upgrade flags](https://github.com/cloudfoundry-incubator/kubo-ci/blob/master/scripts/generate-test-config.sh#L109-L111)
1. Run `BOSH_DEPLOY_COMMAND="path/to/your/bosh-deploy-script.sh" CONFIG="/path/to/your/testconfig" ginkgo -r -v -progress "./kubo-ci/src/tests/upgrade-tests/"`