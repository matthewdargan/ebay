# eBay Go API Client

[![GoDoc](https://godoc.org/github.com/matthewdargan/ebay?status.svg)](https://godoc.org/github.com/matthewdargan/ebay)
[![Go Report Card](https://goreportcard.com/badge/github.com/matthewdargan/ebay)](https://goreportcard.com/report/github.com/matthewdargan/ebay)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Package ebay provides an eBay API client and endpoint wrappers
that streamline the process of performing parameter validation,
making API requests, and handling responses.

To interact with the eBay Finding API, create a `FindingClient`:

```go
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
	// handle error
}
```

For more details on the available methods and their usage,
see the examples in the Go documentation.

## Installation

Run the following to import the `ebay` package:

```sh
go get -u github.com/matthewdargan/ebay
```
