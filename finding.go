// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

package ebay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	findingURL        = "https://svcs.ebay.com/services/search/FindingService/v1"
	operationAdvanced = "findItemsAdvanced"
	operationCategory = "findItemsByCategory"
	operationKeywords = "findItemsByKeywords"
	operationProduct  = "findItemsByProduct"
	operationStores   = "findItemsIneBayStores"
	serviceVersion    = "1.0.0"
	responseFormat    = "JSON"
	restPayload       = ""
)

// A FindingClient is a client that interacts with the eBay Finding API.
type FindingClient struct {
	// Client is the HTTP client used to make requests to the eBay Finding API.
	*http.Client

	// AppID is the eBay application ID.
	//
	// AppID must be a valid application ID requested from eBay. If the AppID is not valid,
	// authentication to the eBay Finding API will fail.
	// See https://developer.ebay.com/api-docs/static/gs_create-the-ebay-api-keysets.html.
	AppID string

	// URL specifies the eBay Finding API endpoint.
	//
	// URL defaults to the eBay Production API Gateway URI, but can be changed to
	// the eBay Sandbox endpoint or localhost for testing purposes.
	// See https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-making-a-call.html#Endpoints.
	URL string
}

// NewFindingClient creates a new FindingClient with the given HTTP client and valid eBay application ID.
func NewFindingClient(client *http.Client, appID string) *FindingClient {
	return &FindingClient{Client: client, AppID: appID, URL: findingURL}
}

var (
	// ErrNewRequest is returned when creating an HTTP request fails.
	ErrNewRequest = errors.New("ebay: failed to create HTTP request")

	// ErrFailedRequest is returned when the eBay Finding API request fails.
	ErrFailedRequest = errors.New("ebay: failed to perform eBay Finding API request")

	// ErrInvalidStatus is returned when the eBay Finding API request returns an invalid status code.
	ErrInvalidStatus = errors.New("ebay: failed to perform eBay Finding API request with status code")

	// ErrDecodeAPIResponse is returned when there is an error decoding the eBay Finding API response body.
	ErrDecodeAPIResponse = errors.New("ebay: failed to decode eBay Finding API response body")
)

// FindItemsAdvanced searches for items on eBay by category and/or keyword.
// See [Searching and Browsing By Category] for searching by category
// and [Searching by Keywords] for searching by keywords.
//
// [Searching and Browsing By Category]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-browsing-by-category.html
// [Searching by Keywords]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html
func (c *FindingClient) FindItemsAdvanced(ctx context.Context, params map[string]string) (*FindItemsAdvancedResponse, error) {
	req, err := c.newRequest(ctx, operationAdvanced, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewRequest, err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedRequest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStatus, resp.StatusCode)
	}
	var res FindItemsAdvancedResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}
	return &res, nil
}

// FindItemsByCategory searches for items on eBay using specific eBay category ID numbers.
// See [Searching and Browsing By Category] for searching by category.
//
// [Searching and Browsing By Category]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-browsing-by-category.html
func (c *FindingClient) FindItemsByCategory(ctx context.Context, params map[string]string) (*FindItemsByCategoryResponse, error) {
	req, err := c.newRequest(ctx, operationCategory, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewRequest, err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedRequest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStatus, resp.StatusCode)
	}
	var res FindItemsByCategoryResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}
	return &res, nil
}

// FindItemsByKeywords searches for items on eBay by a keyword query.
// See [Searching by Keywords] for searching by keywords.
//
// [Searching by Keywords]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html
func (c *FindingClient) FindItemsByKeywords(ctx context.Context, params map[string]string) (*FindItemsByKeywordsResponse, error) {
	req, err := c.newRequest(ctx, operationKeywords, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewRequest, err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedRequest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStatus, resp.StatusCode)
	}
	var res FindItemsByKeywordsResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}
	return &res, nil
}

// FindItemsByProduct searches for items on eBay using specific eBay product values.
// See [Searching by Product] for searching by product.
//
// [Searching by Product]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-product.html
func (c *FindingClient) FindItemsByProduct(ctx context.Context, params map[string]string) (*FindItemsByProductResponse, error) {
	req, err := c.newRequest(ctx, operationProduct, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewRequest, err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedRequest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStatus, resp.StatusCode)
	}
	var res FindItemsByProductResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}
	return &res, nil
}

// FindItemsInEBayStores searches for items in the eBay store inventories. The search can utilize a combination of
// store name, category IDs, and/or keywords. If a search includes keywords and/or category IDs but lacks a store name,
// it will search for items across all eBay stores.
// See [Searching and Browsing By Category] for searching by category
// and [Searching by Keywords] for searching by keywords.
//
// [Searching and Browsing By Category]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-browsing-by-category.html
// [Searching by Keywords]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html
func (c *FindingClient) FindItemsInEBayStores(ctx context.Context, params map[string]string) (*FindItemsInEBayStoresResponse, error) {
	req, err := c.newRequest(ctx, operationStores, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewRequest, err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedRequest, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrInvalidStatus, resp.StatusCode)
	}
	var res FindItemsInEBayStoresResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err)
	}
	return &res, nil
}

func (c *FindingClient) newRequest(ctx context.Context, op string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return nil, err
	}
	qry := req.URL.Query()
	qry.Set("Operation-Name", op)
	qry.Set("Service-Version", serviceVersion)
	qry.Set("Security-AppName", c.AppID)
	qry.Set("Response-Data-Format", responseFormat)
	qry.Set("REST-Payload", restPayload)
	for k, v := range params {
		if v != "" {
			qry.Set(k, v)
		}
	}
	req.URL.RawQuery = qry.Encode()
	return req, nil
}
