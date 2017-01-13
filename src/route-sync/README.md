# route-sync

Synchronizes routes from a Kubernetes cluster to external L3/L7 routers.

## Dependencies

- golang 1.7.4 

## Development

This repo should be imported as `route-sync` meaning your GOPATH should be set to `~/kubo-release`

## Running tests

``
# ensure ginkgo is installed
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

ginkgo -r
``` 
