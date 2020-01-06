package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli"
)

func recordMetrics(conf *config) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if err := parseFlags(conf, c); err != nil {
			cli.ShowAppHelp(c)
			log.Fatal(err)
			return err

		}
		go func() {
			for {

				var date = time.Now().Add(time.Duration(-4) * time.Minute).Format(time.RFC3339)

				// Construct a new API object
				api, err := cloudflare.New(conf.apiKey, conf.apiEmail)
				if err != nil {
					log.Println(err)
				}
				zones, err := api.ListZones()
				if err != nil {
					log.Println("Listing zone errored: ", err)
				}
				for _, zone := range zones {
					if zone.Plan.ZonePlanCommon.Name == "Enterprise Website" {
						log.Println(zone.Name)
						resp, err := getCloudflareCacheMetrics(buildGraphQLQuery(date, zone.ID), conf.apiEmail, conf.apiKey)

						if err == nil {
							for _, node := range resp.Viewer.Zones[0].Caching{
								requestBytes.With(prometheus.Labels{"cacheStatus": node.Dimensions.CacheStatus, "zoneName": zone.Name}).Set(float64(node.SumEdgeResponseBytes.EdgeResponseBytes))
							}
							for _, node := range resp.Viewer.Zones[0].ResponseCodes{
								requestResponseCodes.With(prometheus.Labels{"responseCode": strconv.Itoa(node.SumResponseStatus.ResponseStatusMap[0].EdgeResponseStatus), "zoneName": zone.Name}).Set(float64(node.SumResponseStatus.ResponseStatusMap[0].Requests))
							}
							log.Println("Fetch done at:", date)
							fetchDone.Inc()
						} else {
							log.Println("Fetch failed :", err)
							fetchFailed.Inc()
						}
					}

				}
				time.Sleep(240 * time.Second)
			}
		}()
		return nil
	}
}

var (
	fetchFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cloudflare_failed_fetches",
		Help: "The total number of failed fetches",
	})
	fetchDone = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cloudflare_done_fetches",
		Help: "The total number of done fetches",
	})
	requestBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudflare_processed_bytes",
		Help: "The total number of processed bytes, labelled per cache status",
	},
		[]string{"cacheStatus", "zoneName"},
	)
	requestResponseCodes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudflare_requests_per_response_code",
		Help: "The total number of request, labelled per HTTP response codes",
	},
		[]string{"responseCode", "zoneName"},
	)
)

func main() {
	log.SetPrefix("[cloudflare-exporter] ")
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stderr)

	app := cli.NewApp()
	app.Name = "cloudflare-exporter"
	app.Usage = "export Cloudflare metrics to prometheus"
	app.Flags = flags

	conf := &config{}
	app.Action = recordMetrics(conf)

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
	}
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		log.Fatal(err)
	}
}

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "api-key",
		Usage: "Your Cloudflare API token",
	},
	cli.StringFlag{
		Name:  "api-email",
		Usage: "The email address associated with your Cloudflare API token and account",
	},
}

type config struct {
	apiKey   string
	apiEmail string
}

func parseFlags(conf *config, c *cli.Context) error {
	conf.apiKey = c.String("api-key")
	conf.apiEmail = c.String("api-email")

	return conf.Validate()
}

func (conf *config) Validate() error {

	if conf.apiKey == "" || conf.apiEmail == "" {
		return errors.New("Must provide both api-key and api-email")
	}

	return nil
}
