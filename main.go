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
                            log.Printf("%+v\n", resp)
							for _, node := range resp.Viewer.Zones[0].Caching{
								requestBytes.With(prometheus.Labels{
									"zone_name": zone.Name, "cache_status": node.Dimensions.CacheStatus}).Set(float64(node.SumEdgeResponseBytes.EdgeResponseBytes))
							}
							for _, node := range resp.Viewer.Zones[0].ResponseCodes{
								requestResponseCodes.With(prometheus.Labels{
									"zone_name": zone.Name, "response_code": strconv.Itoa(node.SumResponseStatus.ResponseStatusMap[0].EdgeResponseStatus)}).Set(float64(node.SumResponseStatus.ResponseStatusMap[0].Requests))
							}
							for _, node := range resp.Viewer.Zones[0].FirewallEvents {
								requestFirewallEvents.With(prometheus.Labels{
									"zone_name": zone.Name, "action": node.Dimensions.Action, "client_country_name": node.Dimensions.ClientCountryName, "client_ip": node.Dimensions.ClientIP, "rule_id": node.Dimensions.RuleID}).Set(float64(node.Count))
							}
							for _, node := range resp.Viewer.Zones[0].LoadBalancerEvents{
								requestLoadBalancerEvents.With(prometheus.Labels{
									"zone_name": zone.Name, "colo_code": node.Dimensions.ColoCode, "selected_origin_name": node.Dimensions.SelectedOriginName, "lb_name": node.Dimensions.LBName}).Set(float64(node.Count))
							}
							for _, node := range resp.Viewer.Zones[0].RequestsByColo{
								requestRequestsByColo.With(prometheus.Labels{
									"zone_name": zone.Name, "colo_code": node.Dimensions.ColoCode}).Set(float64(node.Sum.Requests))
							}
//							for _, node := range resp.Viewer.Zones[0].CacheRequests {
//								requestCacheRequests_c.With(prometheus.Labels{
//									"zone_name": zone.Name}).Set(float64(node.SumResponseStatus.CachedRequests))
//							}
//							for _, node := range resp.Viewer.Zones[0].CacheRequests{
//								requestCacheRequests_u.With(prometheus.Labels{
//									"zone_name": zone.Name}).Set(float64(node.SumResponseStatus.Requests))
//							}
							log.Println("Fetch done at:", date)
							fetchDone.Inc()
						} else {
							log.Println("Fetch failed :", err)
							fetchFailed.Inc()
						}
					}

				}
				time.Sleep(30 * time.Second)
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
		[]string{"cache_status", "zone_name"},
	)
	requestResponseCodes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudflare_requests_per_response_code",
		Help: "The total number of request, labelled per HTTP response codes",
	},
		[]string{"response_code", "zone_name"},
	)
	requestFirewallEvents = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudflare_firewall_events",
		Help: "WAF events",
	},
		[]string{"action", "zone_name", "client_country_name", "client_ip", "rule_id"},
	)
	requestLoadBalancerEvents = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudflare_loadbalancer_events",
		Help: "WAF events",
	},
		[]string{"colo_code", "zone_name", "selected_origin_name", "lb_name"},
	)
	requestRequestsByColo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudflare_requests_by_colo",
		Help: "The total number of processed requests, labelled per point of presence",
	},
		[]string{"colo_code", "zone_name"},
	)
//	requestCacheRequests_c = promauto.NewGaugeVec(prometheus.GaugeOpts{
//		Name: "cloudflare_requests_cached",
//		Help: "The total number of processed requests, labelled per point of presence",
//	},
//		[]string{"zonename"},
//	)
//	requestCacheRequests_u = promauto.NewGaugeVec(prometheus.GaugeOpts{
//		Name: "cloudflare_requests_not_cached",
//		Help: "The total number of processed requests, labelled per point of presence",
//	},
//		[]string{"zonename"},
//	)
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
	&cli.StringFlag{
		Name:  "api-key",
		Usage: "Your Cloudflare API token",
	},
	&cli.StringFlag{
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
