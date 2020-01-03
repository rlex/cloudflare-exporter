package main

import (
	"context"

	"github.com/machinebox/graphql"
)

type respDataStruct struct {
	Viewer Viewer `json:"viewer"`
}
type Dimensions struct {
	CacheStatus string `json:"cacheStatus"`
}
type Sum struct {
	EdgeResponseBytes int `json:"edgeResponseBytes"`
}
type HTTPRequestsCacheGroups struct {
	Dimensions Dimensions `json:"dimensions"`
	Sum        Sum        `json:"sum"`
}
type Zones struct {
	HTTPRequestsCacheGroups []HTTPRequestsCacheGroups `json:"httpRequestsCacheGroups"`
}
type Viewer struct {
	Zones []Zones `json:"zones"`
}

func buildGraphQLQuery(date string, zoneID string) *graphql.Request {

	queryForCache := graphql.NewRequest(`
	{
		viewer {
		  zones(filter: { zoneTag: $zoneTag }) {
			httpRequestsCacheGroups(
			  limit: 10000
			  filter: { datetime_gt: $lastSuccessfulScrape }
			) {
			  dimensions {
				cacheStatus
			  }
			  sum {
				edgeResponseBytes
			  }
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
