package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
	"golang.org/x/net/html"
)

var wavefrontToken = getConfig()

func main() {

	// Tags we'll add to the metric
	tags := map[string]string{
		"env": "dev",
	}

	directCfg := &senders.DirectConfiguration{
		Server:               "https://vmware.wavefront.com",
		Token:                wavefrontToken,
		BatchSize:            10000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: 1,
	}

	sender, err := senders.NewDirectSender(directCfg)
	if err != nil {
		panic(err)
	}

	reporter := reporting.NewReporter(
		sender,
		application.New("compute-test-app", "compute-test-service"),
		reporting.Source("compute-test"),
		reporting.Prefix("compute.test"),
		reporting.LogErrors(true),
	)

	counter := metrics.NewCounter()
	err = reporter.RegisterMetric("counter", counter, tags)
	guardError(err)

	histogram := reporting.NewHistogram()
	err = reporter.RegisterMetric("hist1", histogram, tags)
	guardError(err)

	deltaCounter := metrics.NewCounter()
	err = reporter.RegisterMetric(reporting.DeltaCounterName("delta.metric"), deltaCounter, tags)
	guardError(err)

	fmt.Println("Search wavefront: ts(\"compute.test.counter\")")
	fmt.Println("Entering loop to simulate metrics flushing. Hit ctrl+c to cancel")

	resp, err := http.Get("https://www.worldometers.info/coronavirus/coronavirus-cases/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(body)

	z := html.NewTokenizer(reader)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			log.Println("End of Parsing")
			log.Fatal(z.Token())
		}
		fmt.Println(z.Token())
	}
	//for {
	//	value := rand.Int63n(16)
	//	fmt.Printf("Incrementing by random value: %v\n", value)
	//
	//	counter.Inc(value)
	//	deltaCounter.Inc(value)
	//
	//	time.Sleep(time.Second * 10)
	//}
}

func guardError(err error) {
	if err != nil {
		log.Println("Error Registering Metric:")
		log.Println(err)
	}
}

func getConfig() string {
	devlog("Reading API key from file")
	file, err := ioutil.ReadFile("config/wavefront.token")

	fmt.Println(string(file))

	if err != nil {
		log.Println(err)
		log.Fatal("Could not Read Wavefront Token: Please create `src/config/wavefront.token` file with Auth Token")
	}

	return string(file)
}

func devlog(data string) () {
	f, err := os.OpenFile("local.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(time.Now().String() + "\t" + data + "\n")
	if err != nil {
		log.Fatal(err)
	}
}
