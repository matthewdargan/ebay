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

const findingURL = "https://svcs.ebay.com/services/search/FindingService/v1?REST-PAYLOAD"

// NewFindingClient creates a new FindingClient with the given HTTP client and valid eBay application ID.
func NewFindingClient(client *http.Client, appID string) *FindingClient {
	return &FindingClient{Client: client, AppID: appID, URL: findingURL}
}

// APIError represents an eBay Finding API call error.
type APIError struct {
	// Err is the error that occurred during the call.
	Err error

	// StatusCode is the HTTP status code indicating why the call was bad.
	StatusCode int
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ebay: %v", e.Err)
	}
	return "ebay: API error occurred"
}

// FindItemsByCategories searches for items on eBay using specific eBay category ID numbers.
// See [Searching and Browsing By Category] for searching by category.
//
// The category IDs narrow down the search results. The provided parameters
// contain additional query parameters for the search. If the FindingClient is configured with an invalid
// AppID, the search call will fail to authenticate.
//
// An error of type [*APIError] is returned if the category IDs and/or additional parameters were not valid,
// the request could not be created, the request or response could not be completed, or the response could not
// be parsed into type [FindItemsByCategoriesResponse].
//
// If the returned error is nil, the [FindItemsByCategoriesResponse] will contain a non-nil ItemsResponse
// containing search results.
// See https://developer.ebay.com/devzone/finding/CallRef/findItemsByCategory.html.
//
// [Searching and Browsing By Category]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-browsing-by-category.html
func (c *FindingClient) FindItemsByCategories(ctx context.Context, params map[string]string) (FindItemsByCategoriesResponse, error) {
	var findItems FindItemsByCategoriesResponse
	err := c.findItems(ctx, params, &findItemsByCategoryParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsByKeywords searches for items on eBay by a keyword query.
// See [Searching by Keywords] for searching by keywords.
//
// The keywords narrow down the search results. The provided parameters contain additional query parameters
// for the search. If the FindingClient is configured with an invalid AppID, the search call will fail to authenticate.
//
// An error of type [*APIError] is returned if the keywords and/or additional parameters were not valid,
// the request could not be created, the request or response could not be completed, or the response could not
// be parsed into type [FindItemsByKeywordsResponse].
//
// If the returned error is nil, the [FindItemsByKeywordsResponse] will contain a non-nil ItemsResponse
// containing search results.
// See https://developer.ebay.com/devzone/finding/CallRef/findItemsByKeywords.html.
//
// [Searching by Keywords]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html
func (c *FindingClient) FindItemsByKeywords(ctx context.Context, params map[string]string) (FindItemsByKeywordsResponse, error) {
	var findItems FindItemsByKeywordsResponse
	err := c.findItems(ctx, params, &findItemsByKeywordsParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsAdvanced searches for items on eBay by category and/or keyword.
// See [Searching and Browsing By Category] for searching by category
// and [Searching by Keywords] for searching by keywords.
//
// The category IDs and keywords narrow down the search results. The provided parameters contain additional
// query parameters for the search. If the FindingClient is configured with an invalid AppID,
// the search call will fail to authenticate.
//
// An error of type [*APIError] is returned if the category IDs, keywords, and/or additional parameters were not valid,
// the request could not be created, the request or response could not be completed, or the response could not
// be parsed into type [FindItemsAdvancedResponse].
//
// If the returned error is nil, the [FindItemsAdvancedResponse] will contain a non-nil ItemsResponse
// containing search results.
// See https://developer.ebay.com/Devzone/finding/CallRef/findItemsAdvanced.html.
//
// [Searching and Browsing By Category]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-browsing-by-category.html
// [Searching by Keywords]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html
func (c *FindingClient) FindItemsAdvanced(ctx context.Context, params map[string]string) (FindItemsAdvancedResponse, error) {
	var findItems FindItemsAdvancedResponse
	err := c.findItems(ctx, params, &findItemsAdvancedParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsByProduct searches for items on eBay using specific eBay product values.
// See [Searching by Product] for searching by product.
//
// The product ID narrows down the search results. The provided parameters contain additional query parameters
// for the search. If the FindingClient is configured with an invalid AppID, the search call will fail to authenticate.
//
// An error of type [*APIError] is returned if the product ID and/or additional parameters were not valid,
// the request could not be created, the request or response could not be completed, or the response could not
// be parsed into type [FindItemsByProductResponse].
//
// If the returned error is nil, the [FindItemsByProductResponse] will contain a non-nil ItemsResponse
// containing search results.
// See https://developer.ebay.com/Devzone/finding/CallRef/findItemsByProduct.html.
//
// [Searching by Product]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-product.html
func (c *FindingClient) FindItemsByProduct(ctx context.Context, params map[string]string) (FindItemsByProductResponse, error) {
	var findItems FindItemsByProductResponse
	err := c.findItems(ctx, params, &findItemsByProductParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

// FindItemsInEBayStores searches for items in the eBay store inventories. The search can utilize a combination of
// store name, category IDs, and/or keywords. If a search includes keywords and/or category IDs but lacks a store name,
// it will search for items across all eBay stores.
// See [Searching and Browsing By Category] for searching by category
// and [Searching by Keywords] for searching by keywords.
//
// The store name, category IDs, and keywords narrow down the search results. The provided parameters contain
// additional query parameters for the search. If the FindingClient is configured with an invalid AppID,
// the search call will fail to authenticate.
//
// An error of type [*APIError] is returned if the store name, category IDs, keywords, and/or additional parameters
// were not valid, the request could not be created, the request or response could not be completed,
// or the response could not be parsed into type [FindItemsInEBayStoresResponse].
//
// If the returned error is nil, the [FindItemsInEBayStoresResponse] will contain a non-nil ItemsResponse
// containing search results.
// See https://developer.ebay.com/Devzone/finding/CallRef/findItemsIneBayStores.html.
//
// [Searching and Browsing By Category]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-browsing-by-category.html
// [Searching by Keywords]: https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html
func (c *FindingClient) FindItemsInEBayStores(ctx context.Context, params map[string]string) (FindItemsInEBayStoresResponse, error) {
	var findItems FindItemsInEBayStoresResponse
	err := c.findItems(ctx, params, &findItemsInEBayStoresParams{appID: c.AppID}, &findItems)
	if err != nil {
		return findItems, err
	}
	return findItems, nil
}

var (
	// ErrFailedRequest is returned when the eBay Finding API request fails.
	ErrFailedRequest = errors.New("failed to perform eBay Finding API request")

	// ErrInvalidStatus is returned when the eBay Finding API request returns an invalid status code.
	ErrInvalidStatus = errors.New("failed to perform eBay Finding API request with status code")

	// ErrDecodeAPIResponse is returned when there is an error decoding the eBay Finding API response body.
	ErrDecodeAPIResponse = errors.New("failed to decode eBay Finding API response body")
)

func (c *FindingClient) findItems(ctx context.Context, params map[string]string, v findParamsValidator, res ResultProvider) error {
	err := v.validate(params)
	if err != nil {
		return &APIError{Err: err, StatusCode: http.StatusBadRequest}
	}
	req, err := v.newRequest(ctx, c.URL)
	if err != nil {
		return &APIError{Err: err, StatusCode: http.StatusInternalServerError}
	}
	resp, err := c.Do(req)
	if err != nil {
		return &APIError{Err: fmt.Errorf("%w: %w", ErrFailedRequest, err), StatusCode: http.StatusInternalServerError}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &APIError{
			Err:        fmt.Errorf("%w %d", ErrInvalidStatus, resp.StatusCode),
			StatusCode: http.StatusInternalServerError,
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return &APIError{
			Err:        fmt.Errorf("%w: %w", ErrDecodeAPIResponse, err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	return nil
}
