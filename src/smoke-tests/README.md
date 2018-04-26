# CFCR Smoke Tests

Tests that run against a remote kubernetes cluster


## How To Run

1. Local from the command line given `kubectl` works

```
ginkgo -r
```

2. As a bosh-errand

```
bosh -d cfcr run-errand smoke-tests
```

3. Remote using a binary (e.g. jumphost)

```
GOARCH=amd64 GOOS=linux go test  -c -v -o run-tests
scp run-tests jumphost:~/
ssh jumphost -c ~/run-tests
```
