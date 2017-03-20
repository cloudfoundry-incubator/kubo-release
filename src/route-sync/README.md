# route-sync

Synchronizes routes from a Kubernetes cluster to external L3/L7 routers.

## Dependencies

- golang 1.8 

## Development

This repo should be imported as `route-sync` meaning your GOPATH should be set to the `kubo-release` directory.

### Running the application

```
cp .envrc.example .envrc

# modify the config for your enviornment
$EDITOR .envrc

direnv allow

go build && ./route-sync
```

### Running tests

```
# ensure ginkgo is installed
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

ginkgo -r
``` 
