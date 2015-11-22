package main
import (
	"flag"
	"net/http"
	"time"
	"github.com/wrong-kendall/tank_utility"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)


var (
	Version = "0.0.0.dev"

	listenAddress = flag.String("web.listen-address", ":9494", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	pollRate   = flag.Int("poll-rate", 15, "How frequently to poll for tank utility updates (in minutes)")

	insecure = flag.Bool("insecure", true, "Whether to skip certificate checks.")
	token_file = flag.String("token_file", "", "Path to read the API token from (or write to).")
	tank_utility_endpoint = flag.String("tank_utility_endpoint", "https://data.tankutility.com/api", "API endpoint for Tank Utility")
)

var (
	percentage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "propane_percentage",
			Help: "The percentage remaining reported during the last poll.",
		},
		[]string{"name"},
	)
	temperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "temperature",
			Help: "The temperature reported during the last poll.",
		},
		[]string{"name"},
	)
	capacity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "capacity",
			Help: "The capacity of the tank.",
		},
		[]string{"name"},
	)
	timestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "last_update",
			Help: "The timestamp of the last reading.",
		},
		[]string{"name"},
	)
)

func readTanks() {
	var token_response tank_utility.TokenResponse
	if *token_file != "" {
		token_response = tank_utility.ReadTokenFromFile(*token_file)
	} else {
		log.Error("Could not get token.")
	}
	token := token_response.Token
	// TODO(kendall): Handle invalid or no token_file set.
	device_list := tank_utility.GetDeviceList(token, *tank_utility_endpoint, *insecure).Devices
	for i := 0; i < len(device_list); i++ {
		var device_info tank_utility.DeviceInfo
		device_info = tank_utility.GetDeviceInfo(device_list[i], token, *tank_utility_endpoint, *insecure)
		timestamp.WithLabelValues(device_info.Device.Name).Set(float64(device_info.Device.LastReading.Time))
		capacity.WithLabelValues(device_info.Device.Name).Set(float64(device_info.Device.Capacity))
		percentage.WithLabelValues(device_info.Device.Name).Set(device_info.Device.LastReading.Tank)
		temperature.WithLabelValues(device_info.Device.Name).Set(float64(device_info.Device.LastReading.Temperature))
	}

}

func startPolling() {
	t := time.NewTicker(time.Duration(*pollRate) * time.Minute)
	log.Info("Starting polling of tanks.")
	for {
		readTanks()
		<-t.C
	}
}

func main() {
	flag.Parse()

	handler := prometheus.Handler()
	prometheus.MustRegister(percentage)
	prometheus.MustRegister(temperature)
	prometheus.MustRegister(timestamp)
	prometheus.MustRegister(capacity)

	http.Handle(*metricsPath, handler)
	go startPolling()

	log.Infof("Starting tank_utility_exporter v%s at %s", Version, *listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
