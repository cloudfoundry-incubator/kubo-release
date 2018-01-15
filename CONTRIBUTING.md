# Contributing to Kubo

## Contributor License Agreement
If you have not previously done so, please fill out and submit an [Individual
Contributor License
Agreement](https://www.cloudfoundry.org/governance/cff_individual_cla/) or a
[Corporate Contributor License
Agreement](https://www.cloudfoundry.org/governance/cff_corporate_cla/).

## Developer Workflow
Before making significant changes it's best to communicate with the maintainers
of the project through [GitHub
Issues](https://github.com/cloudfoundry-incubator/kubo-release/issues).

1. Fork the project on [GitHub](https://github.com/cloudfoundry-incubator/kubo-release)
1. Make your feature addition or bug fix. Please make sure there is appropriate test coverage.
1. [Run tests](#running-the-tests).
1. Make sure your fork is up to date with `master`.
1. Send a pull request.

## Running the Tests
### Running Release Tests
Running `./scripts/run_tests` will run the unit tests for kubo-release.  These
includes tests for [route-sync](src/route-sync) and the [bosh templating
tests](spec/).  Please run these tests before submitting a pull request.

### Running Integration Tests
Before submitting pull request please deploy
[kubo-deployment](https://github.com/cloudfoundry-incubator/kubo-deployment),
and run the integration tests.
