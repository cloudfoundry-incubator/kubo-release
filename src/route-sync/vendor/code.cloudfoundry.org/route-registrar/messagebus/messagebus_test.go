package messagebus_test

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/route-registrar/config"
	"code.cloudfoundry.org/route-registrar/messagebus"
	"github.com/nats-io/nats"
)

var _ = Describe("Messagebus test Suite", func() {
	var (
		natsCmd      *exec.Cmd
		natsHost     string
		natsUsername string
		natsPassword string

		testSpyClient *nats.Conn

		logger            lager.Logger
		messageBusServers []config.MessageBusServer
		messageBus        messagebus.MessageBus
	)

	BeforeEach(func() {
		natsUsername = "nats-user"
		natsPassword = "nats-pw"
		natsHost = "127.0.0.1"

		natsCmd = startNats(natsHost, natsPort, natsUsername, natsPassword)

		logger = lagertest.NewTestLogger("nats-test")
		var err error
		servers := []string{
			fmt.Sprintf(
				"nats://%s:%s@%s:%d",
				natsUsername,
				natsPassword,
				natsHost,
				natsPort,
			),
		}

		opts := nats.DefaultOptions
		opts.Servers = servers

		testSpyClient, err = opts.Connect()
		Expect(err).ToNot(HaveOccurred())

		// Ensure nats server is listening before tests
		Eventually(func() string {
			connStatus := testSpyClient.Status()
			return fmt.Sprintf("%v", connStatus)
		}, 5*time.Second).Should(Equal("1"))

		Expect(err).ShouldNot(HaveOccurred())

		messageBusServer := config.MessageBusServer{
			fmt.Sprintf("%s:%d", natsHost, natsPort),
			natsUsername,
			natsPassword,
		}

		messageBusServers = []config.MessageBusServer{messageBusServer}

		messageBus = messagebus.NewMessageBus(logger)
	})

	AfterEach(func() {
		testSpyClient.Close()

		err := natsCmd.Process.Kill()
		Expect(err).NotTo(HaveOccurred())
		_, err = natsCmd.Process.Wait()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Connect", func() {
		It("connects without error", func() {
			err := messageBus.Connect(messageBusServers)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("when no servers are provided", func() {
			BeforeEach(func() {
				messageBusServers = []config.MessageBusServer{}
			})

			It("returns error", func() {
				err := messageBus.Connect(messageBusServers)
				Expect(err).Should(HaveOccurred())
			})
		})

		Context("when nats connection is successful", func() {
			BeforeEach(func() {
				err := messageBus.Connect(messageBusServers)
				Expect(err).ShouldNot(HaveOccurred())
			})
			It("logs a message", func() {
				Eventually(logger).Should(gbytes.Say(`nats-connection-successful`))
				Eventually(logger).Should(gbytes.Say(fmt.Sprintf("%s", natsHost)))
			})
		})

		Context("when nats connection closes", func() {
			BeforeEach(func() {
				err := messageBus.Connect(messageBusServers)
				Expect(err).ShouldNot(HaveOccurred())
				messageBus.Close()
			})

			It("logs a message", func() {
				Eventually(logger).Should(gbytes.Say(`nats-connection-disconnected`))
				Eventually(logger).Should(gbytes.Say(fmt.Sprintf("%s", natsHost)))
				Eventually(logger).Should(gbytes.Say(`nats-connection-closed`))
				Eventually(logger).Should(gbytes.Say(fmt.Sprintf("%s", natsHost)))
			})
		})
	})

	Describe("SendMessage", func() {
		const (
			topic             = "router.registrar"
			host              = "some_host"
			privateInstanceId = "some_id"
		)

		var (
			route config.Route
		)

		BeforeEach(func() {
			err := messageBus.Connect(messageBusServers)
			Expect(err).ShouldNot(HaveOccurred())

			route = config.Route{
				Name:            "some_name",
				Port:            12345,
				URIs:            []string{"uri1", "uri2"},
				RouteServiceUrl: "https://rs.example.com",
				Tags:            map[string]string{"tag1": "val1", "tag2": "val2"},
			}
		})

		It("send messages", func() {
			registered := make(chan string)
			testSpyClient.Subscribe(topic, func(msg *nats.Msg) {
				registered <- string(msg.Data)
			})

			// Wait for the nats library to register our callback.
			// We use a sleep because there's no way to know that the callback was
			// registered successfully (e.g. they don't provide a channel)
			time.Sleep(20 * time.Millisecond)

			err := messageBus.SendMessage(topic, host, route, privateInstanceId)
			Expect(err).ShouldNot(HaveOccurred())

			// Assert that we got the right message
			var receivedMessage string
			Eventually(registered, 2).Should(Receive(&receivedMessage))

			expectedRegistryMessage := messagebus.Message{
				URIs:            route.URIs,
				Host:            host,
				Port:            route.Port,
				RouteServiceUrl: route.RouteServiceUrl,
				Tags:            route.Tags,
			}

			var registryMessage messagebus.Message
			err = json.Unmarshal([]byte(receivedMessage), &registryMessage)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(registryMessage.URIs).To(Equal(expectedRegistryMessage.URIs))
			Expect(registryMessage.Port).To(Equal(expectedRegistryMessage.Port))
			Expect(registryMessage.RouteServiceUrl).To(Equal(expectedRegistryMessage.RouteServiceUrl))
			Expect(registryMessage.Tags).To(Equal(expectedRegistryMessage.Tags))
		})

		Context("when the connection is already closed", func() {
			BeforeEach(func() {
				err := messageBus.Connect(messageBusServers)
				Expect(err).ShouldNot(HaveOccurred())

				messageBus.Close()
			})

			It("returns error", func() {
				err := messageBus.SendMessage(topic, host, route, privateInstanceId)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
})

func startNats(host string, port int, username, password string) *exec.Cmd {
	fmt.Fprintf(GinkgoWriter, "Starting gnatsd on port %d\n", port)

	cmd := exec.Command(
		"gnatsd",
		"-p", strconv.Itoa(port),
		"--user", username,
		"--pass", password)

	err := cmd.Start()
	if err != nil {
		fmt.Printf("gnatsd failed to start: %v\n", err)
	}

	natsTimeout := 10 * time.Second
	natsPollingInterval := 20 * time.Millisecond
	Eventually(func() error {
		_, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
		return err
	}, natsTimeout, natsPollingInterval).Should(Succeed())

	fmt.Fprintf(GinkgoWriter, "gnatsd running on port %d\n", port)
	return cmd
}
