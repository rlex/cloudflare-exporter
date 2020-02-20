package main

import (
	"context"

	"github.com/machinebox/graphql"
)

type respDataStruct struct {
	Viewer Viewer `json:"viewer"`
}

type Viewer struct {
	Zones []Zone `json:"zones"`
}

type Zone struct {
	LoadBalancerEvents []LoadBalancerEvent `json:"LoadBalancerEvents"`
	CacheRequests      []CacheRequest      `json:"cacheRequests"`
	Caching            []Caching           `json:"caching"`
    FirewallEvents     []FirewallEvent     `json:"firewallEvents"`
	RequestsByColo     []RequestsByColo    `json:"requestsByColo"`
	ResponseCodes      []ResponseCode      `json:"responseCodes"`
}

type CacheRequest struct {
	SumResponseStatus CacheRequestSumResponseStatus `json:"SumResponseStatus"`
}

type CacheRequestSumResponseStatus struct {
	CachedRequests int64 `json:"cachedRequests"`
	Requests       int64 `json:"requests"`
}

type Caching struct {
    SumEdgeResponseBytes SumEdgeResponseBytes `json:"SumEdgeResponseBytes"`
	Count                int64                `json:"count"`
	Dimensions           CachingDimensions    `json:"dimensions"`
}

type CachingDimensions struct {
	CacheStatus string `json:"cacheStatus"`
}

type SumEdgeResponseBytes struct {
	EdgeResponseBytes int64 `json:"edgeResponseBytes"`
}

type FirewallEvent struct {
	Count      int64                   `json:"count"`
	Dimensions FirewallEventDimensions `json:"dimensions"`
}

type FirewallEventDimensions struct {
	Action            string `json:"action"`
	ClientCountryName string `json:"clientCountryName"`
	ClientIP          string `json:"clientIP"`
	RuleID            string `json:"ruleId"`
}

type LoadBalancerEvent struct {
	Count      int64                       `json:"count"`
	Dimensions LoadBalancerEventDimensions `json:"dimensions"`
}

type LoadBalancerEventDimensions struct {
	ColoCode           string `json:"coloCode"`
	LBName             string `json:"lbName"`
	SelectedOriginName string `json:"selectedOriginName"`
}

type RequestsByColo struct {
	Dimensions RequestsByColoDimensions `json:"dimensions"`
	Sum        RequestsByColoSum        `json:"sum"`
}

type RequestsByColoDimensions struct {
	ColoCode string `json:"coloCode"`
}

type RequestsByColoSum struct {
	Requests int64 `json:"requests"`
}

type ResponseStatusMap struct {
	EdgeResponseStatus int `json:"edgeResponseStatus"`
	Requests           int `json:"requests"`
}

type SumResponseStatus struct {
	ResponseStatusMap []ResponseStatusMap `json:"responseStatusMap"`
}

type ResponseCode struct {
	SumResponseStatus SumResponseStatus `json:"sumResponseStatus"`
}


func buildGraphQLQuery(date string, zoneID string) *graphql.Request {

	queryForCache := graphql.NewRequest(`
	{
		viewer {
			zones(filter: {zoneTag: $zoneTag}) {
				caching: httpRequestsCacheGroups(limit: 10000, filter: { datetime_gt: $lastSuccessfulScrape}) {
					dimensions {
						cacheStatus
					}
					SumEdgeResponseBytes: sum {
						edgeResponseBytes
					}
					count
				}
				responseCodes: httpRequests1mGroups(limit: 10000, filter: { datetime_gt: $lastSuccessfulScrape}) {
          SumResponseStatus: sum {
						responseStatusMap {
							edgeResponseStatus
							requests
						}
					}
				}
				requestsByColo: httpRequests1mByColoGroups(limit: 10000, filter: { datetime_gt: $lastSuccessfulScrape}) {
					dimensions {
						coloCode
					}
					sum {
						requests
					}
				}
				firewallEvents: firewallEventsAdaptiveGroups(limit: 10000, filter: { datetime_gt: $lastSuccessfulScrape}) {
					dimensions {
						action
						clientCountryName
						clientIP
                        ruleId
					}
					count
				}
				LoadBalancerEvents: loadBalancingRequestsGroups(limit: 10000, filter: { datetime_gt: $lastSuccessfulScrape}) {
					dimensions {
						coloCode
						selectedOriginName
						lbName
					}
					count
				}
			}
		}
	}
	  `)

	// set any variables
	queryForCache.Var("zoneTag", zoneID)
	queryForCache.Var("lastSuccessfulScrape", date)

	return queryForCache
}

// Get cloudflare metrics from GraphQL using the provided api-email and api-key parameters and returns a marshalled JSON struct and an error if something went wrong during the fetch
func getCloudflareCacheMetrics(query *graphql.Request, apiEmail string, apiKey string) (respData respDataStruct, err error) {
	client := graphql.NewClient("https://api.cloudflare.com/client/v4/graphql")

	req := query

	// set header fields -> token and email!
	req.Header.Set("x-auth-key", apiKey)
	req.Header.Set("x-auth-email", apiEmail)
	// define a Context for the request
	ctx := context.Background()

	// run it and capture the response

	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, err
	}
	return respData, nil
}
