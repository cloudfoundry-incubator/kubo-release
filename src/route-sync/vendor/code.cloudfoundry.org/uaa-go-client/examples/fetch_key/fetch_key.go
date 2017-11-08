package fetch_key

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	client "code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/config"
)

func main() {
	var (
		err       error
		uaaClient client.Client
		key       string
	)

	if len(os.Args) < 3 {
		fmt.Printf("Usage: <uaa-url> <skip-verification>\n\n")
		fmt.Printf("For example: https://uaa.service.cf.internal:8443 true\n")
		return
	}

	skip, err := strconv.ParseBool(os.Args[2])
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	cfg := &config.Config{
		UaaEndpoint:      os.Args[1],
		SkipVerification: skip,
	}

	logger := lager.NewLogger("test")
	clock := clock.NewClock()

	uaaClient, err = client.NewClient(logger, cfg, clock)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Printf("Connecting to: %s ...\n", cfg.UaaEndpoint)

	key, err = uaaClient.FetchKey()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Printf("Response:\n%s\n", key)

}
