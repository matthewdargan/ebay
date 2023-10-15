// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

/*
Package ebay provides an eBay API client and endpoint wrappers
that streamline the process of performing parameter validation,
making API requests, and handling responses.

To interact with the eBay Finding API, create a [FindingClient]:

	params := map[string]string{
		"categoryId":            "9355",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "500.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	c := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	client := ebay.NewFindingClient(c, appID)
	resp, err := client.FindItemsByCategories(context.Background(), params)
	if err != nil {
		// handle error
	}

For more details on the available methods and their usage,
see the examples under [FindingClient].
*/
package ebay
