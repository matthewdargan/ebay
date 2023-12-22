// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

package ebay_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/matthewdargan/ebay"
)

func ExampleFindingClient_FindItemsAdvanced() {
	params := map[string]string{
		"categoryId":            "9355",
		"keywords":              "iphone",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "500.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	c := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	client := ebay.NewFindingClient(c, appID)
	resp, err := client.FindItemsAdvanced(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request with status code: 500
}

func ExampleFindingClient_FindItemsByCategory() {
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
	resp, err := client.FindItemsByCategory(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request with status code: 500
}

func ExampleFindingClient_FindItemsByKeywords() {
	params := map[string]string{
		"keywords":              "iphone",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "500.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	c := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	client := ebay.NewFindingClient(c, appID)
	resp, err := client.FindItemsByKeywords(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request with status code: 500
}

func ExampleFindingClient_FindItemsByProduct() {
	params := map[string]string{
		"productId.@type":       "ISBN",
		"productId":             "9780131101630",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "50.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	c := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	client := ebay.NewFindingClient(c, appID)
	resp, err := client.FindItemsByProduct(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request with status code: 500
}

func ExampleFindingClient_FindItemsInEBayStores() {
	params := map[string]string{
		"storeName":             "Supplytronics",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "50.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	c := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	client := ebay.NewFindingClient(c, appID)
	resp, err := client.FindItemsInEBayStores(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request with status code: 500
}
