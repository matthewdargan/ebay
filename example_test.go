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

func ExampleFindingClient_FindItemsByCategories() {
	params := map[string]string{
		"categoryId":            "9355",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "500.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	client := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	fc := ebay.NewFindingClient(client, appID)
	fc.URL = "http://127.0.0.1"
	resp, err := fc.FindItemsByCategories(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request: Get "http://127.0.0.1?OPERATION-NAME=findItemsByCategory&RESPONSE-DATA-FORMAT=JSON&SECURITY-APPNAME=your_app_id&SERVICE-VERSION=1.0.0&categoryId%280%29=9355&itemFilter%280%29.name=MaxPrice&itemFilter%280%29.paramName=Currency&itemFilter%280%29.paramValue=EUR&itemFilter%280%29.value%280%29=500.0": dial tcp 127.0.0.1:80: connect: connection refused
}

func ExampleFindingClient_FindItemsByKeywords() {
	params := map[string]string{
		"keywords":              "iphone",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "500.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	client := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	fc := ebay.NewFindingClient(client, appID)
	fc.URL = "http://127.0.0.1"
	resp, err := fc.FindItemsByKeywords(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request: Get "http://127.0.0.1?OPERATION-NAME=findItemsByKeywords&RESPONSE-DATA-FORMAT=JSON&SECURITY-APPNAME=your_app_id&SERVICE-VERSION=1.0.0&itemFilter%280%29.name=MaxPrice&itemFilter%280%29.paramName=Currency&itemFilter%280%29.paramValue=EUR&itemFilter%280%29.value%280%29=500.0&keywords=iphone": dial tcp 127.0.0.1:80: connect: connection refused
}

func ExampleFindingClient_FindItemsAdvanced() {
	params := map[string]string{
		"categoryId":            "9355",
		"keywords":              "iphone",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "500.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	client := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	fc := ebay.NewFindingClient(client, appID)
	fc.URL = "http://127.0.0.1"
	resp, err := fc.FindItemsAdvanced(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request: Get "http://127.0.0.1?OPERATION-NAME=findItemsAdvanced&RESPONSE-DATA-FORMAT=JSON&SECURITY-APPNAME=your_app_id&SERVICE-VERSION=1.0.0&categoryId%280%29=9355&itemFilter%280%29.name=MaxPrice&itemFilter%280%29.paramName=Currency&itemFilter%280%29.paramValue=EUR&itemFilter%280%29.value%280%29=500.0&keywords=iphone": dial tcp 127.0.0.1:80: connect: connection refused
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
	client := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	fc := ebay.NewFindingClient(client, appID)
	fc.URL = "http://127.0.0.1"
	resp, err := fc.FindItemsByProduct(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request: Get "http://127.0.0.1?OPERATION-NAME=findItemsByProduct&RESPONSE-DATA-FORMAT=JSON&SECURITY-APPNAME=your_app_id&SERVICE-VERSION=1.0.0&itemFilter%280%29.name=MaxPrice&itemFilter%280%29.paramName=Currency&itemFilter%280%29.paramValue=EUR&itemFilter%280%29.value%280%29=50.0&productId=9780131101630&productId.%40type=ISBN": dial tcp 127.0.0.1:80: connect: connection refused
}

func ExampleFindingClient_FindItemsInEBayStores() {
	params := map[string]string{
		"storeName":             "Supplytronics",
		"itemFilter.name":       "MaxPrice",
		"itemFilter.value":      "50.0",
		"itemFilter.paramName":  "Currency",
		"itemFilter.paramValue": "EUR",
	}
	client := &http.Client{Timeout: time.Second * 5}
	appID := "your_app_id"
	fc := ebay.NewFindingClient(client, appID)
	fc.URL = "http://127.0.0.1"
	resp, err := fc.FindItemsInEBayStores(context.Background(), params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp)
	}
	// Output:
	// ebay: failed to perform eBay Finding API request: Get "http://127.0.0.1?OPERATION-NAME=findItemsIneBayStores&RESPONSE-DATA-FORMAT=JSON&SECURITY-APPNAME=your_app_id&SERVICE-VERSION=1.0.0&itemFilter%280%29.name=MaxPrice&itemFilter%280%29.paramName=Currency&itemFilter%280%29.paramValue=EUR&itemFilter%280%29.value%280%29=50.0&storeName=Supplytronics": dial tcp 127.0.0.1:80: connect: connection refused
}
