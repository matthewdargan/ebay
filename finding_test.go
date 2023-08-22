// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

package ebay_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/matthewdargan/ebay"
)

type findItemsTestCase struct {
	Name   string
	Params map[string]string
	Err    error
}

var (
	appID     = "super secret ID"
	itemsResp = []ebay.FindItemsResponse{
		{
			Ack:       []string{"Success"},
			Version:   []string{"1.0"},
			Timestamp: []time.Time{time.Date(2023, 6, 24, 0, 0, 0, 0, time.UTC)},
			SearchResult: []ebay.SearchResult{
				{
					Count: "1",
					Item: []ebay.SearchItem{
						{
							ItemID:   []string{"1234567890"},
							Title:    []string{"Sample Item"},
							GlobalID: []string{"global-id-123"},
							Subtitle: []string{"Sample Item Subtitle"},
							PrimaryCategory: []ebay.Category{
								{
									CategoryID:   []string{"category-id-123"},
									CategoryName: []string{"Sample Category"},
								},
							},
							GalleryURL:  []string{"https://example.com/sample-item.jpg"},
							ViewItemURL: []string{"https://example.com/sample-item"},
							ProductID: []ebay.ProductID{
								{
									Type:  "product-type-123",
									Value: "product-value-123",
								},
							},
							AutoPay:    []string{"true"},
							PostalCode: []string{"12345"},
							Location:   []string{"Sample Location"},
							Country:    []string{"Sample Country"},
							ShippingInfo: []ebay.ShippingInfo{
								{
									ShippingServiceCost: []ebay.Price{
										{
											CurrencyID: "USD",
											Value:      "5.99",
										},
									},
									ShippingType:            []string{"Standard"},
									ShipToLocations:         []string{"US"},
									ExpeditedShipping:       []string{"false"},
									OneDayShippingAvailable: []string{"false"},
									HandlingTime:            []string{"1"},
								},
							},
							SellingStatus: []ebay.SellingStatus{
								{
									CurrentPrice: []ebay.Price{
										{
											CurrencyID: "USD",
											Value:      "19.99",
										},
									},
									ConvertedCurrentPrice: []ebay.Price{
										{
											CurrencyID: "USD",
											Value:      "19.99",
										},
									},
									SellingState: []string{"Active"},
									TimeLeft:     []string{"P1D"},
								},
							},
							ListingInfo: []ebay.ListingInfo{
								{
									BestOfferEnabled:  []string{"true"},
									BuyItNowAvailable: []string{"false"},
									StartTime:         []time.Time{time.Date(2023, 6, 24, 0, 0, 0, 0, time.UTC)},
									EndTime:           []time.Time{time.Date(2023, 7, 24, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)},
									ListingType:       []string{"Auction"},
									Gift:              []string{"false"},
									WatchCount:        []string{"10"},
								},
							},
							ReturnsAccepted: []string{"true"},
							Condition: []ebay.Condition{
								{
									ConditionID:          []string{"1000"},
									ConditionDisplayName: []string{"New"},
								},
							},
							IsMultiVariationListing: []string{"false"},
							TopRatedListing:         []string{"true"},
							DiscountPriceInfo: []ebay.DiscountPriceInfo{
								{
									OriginalRetailPrice: []ebay.Price{
										{
											CurrencyID: "USD",
											Value:      "29.99",
										},
									},
									PricingTreatment: []string{"STP"},
									SoldOnEbay:       []string{"true"},
									SoldOffEbay:      []string{"false"},
								},
							},
						},
					},
				},
			},
			PaginationOutput: []ebay.PaginationOutput{
				{
					PageNumber:     []string{"1"},
					EntriesPerPage: []string{"10"},
					TotalPages:     []string{"1"},
					TotalEntries:   []string{"1"},
				},
			},
			ItemSearchURL: []string{"https://example.com/search?q=sample"},
		},
	}
	findItemsByCategoriesResp = ebay.FindItemsByCategoriesResponse{
		ItemsResponse: itemsResp,
	}
	findItemsByKeywordsResp = ebay.FindItemsByKeywordsResponse{
		ItemsResponse: itemsResp,
	}
	findItemsAdvancedResp = ebay.FindItemsAdvancedResponse{
		ItemsResponse: itemsResp,
	}
	findItemsByProductResp = ebay.FindItemsByProductResponse{
		ItemsResponse: itemsResp,
	}
	findItemsInEBayStoresResp = ebay.FindItemsInEBayStoresResponse{
		ItemsResponse: itemsResp,
	}

	findItemsByCategories = "FindItemsByCategories"
	findItemsByKeywords   = "FindItemsByKeywords"
	findItemsAdvanced     = "FindItemsAdvanced"
	findItemsByProduct    = "FindItemsByProduct"
	findItemsInEBayStores = "FindItemsInEBayStores"

	categoryIDTCs = []findItemsTestCase{
		{
			Name:   "can find items if params contains categoryId of length 1",
			Params: map[string]string{"categoryId": "1"},
		},
		{
			Name:   "can find items if params contains categoryId of length 5",
			Params: map[string]string{"categoryId": "1234567890"},
		},
		{
			Name:   "can find items if params contains categoryId of length 10",
			Params: map[string]string{"categoryId": "1234567890"},
		},
		{
			Name:   "returns error if params contains empty categoryId",
			Params: map[string]string{"categoryId": ""},
			Err:    fmt.Errorf("%w: %s: %w", ebay.ErrInvalidCategoryID, `strconv.Atoi: parsing ""`, strconv.ErrSyntax),
		},
		{
			Name:   "returns error if params contains categoryId of length 11",
			Params: map[string]string{"categoryId": "12345678901"},
			Err:    ebay.ErrInvalidCategoryIDLength,
		},
		{
			Name:   "returns error if params contains non-numbered, invalid categoryId",
			Params: map[string]string{"categoryId": "a"},
			Err:    fmt.Errorf("%w: %s: %w", ebay.ErrInvalidCategoryID, `strconv.Atoi: parsing "a"`, strconv.ErrSyntax),
		},
		{
			// categoryId(1) will be ignored because indexing does not start at 0.
			Name:   "can find items if params contains categoryId, categoryId(1)",
			Params: map[string]string{"categoryId": "1", "categoryId(1)": "2"},
		},
		{
			Name:   "returns error if params contain numbered and non-numbered categoryId syntax types",
			Params: map[string]string{"categoryId": "1", "categoryId(0)": "2"},
			Err:    ebay.ErrInvalidIndexSyntax,
		},
		{
			Name:   "can find items by numbered categoryId",
			Params: map[string]string{"categoryId(0)": "1"},
		},
		{
			Name:   "can find items if params contains 2 categoryIds of length 1",
			Params: map[string]string{"categoryId(0)": "1", "categoryId(1)": "2"},
		},
		{
			Name: "can find items if params contains 2 categoryIds of length 10",
			Params: map[string]string{
				"categoryId(0)": "1234567890",
				"categoryId(1)": "9876543210",
			},
		},
		{
			Name: "returns error if params contains 1 categoryId of length 1, 1 categoryId of length 11",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "12345678901",
			},
			Err: ebay.ErrInvalidCategoryIDLength,
		},
		{
			Name: "returns error if params contains 1 categoryId of length 11, 1 categoryId of length 1",
			Params: map[string]string{
				"categoryId(0)": "12345678901",
				"categoryId(1)": "1",
			},
			Err: ebay.ErrInvalidCategoryIDLength,
		},
		{
			Name: "can find items if params contains 3 categoryIds of length 1",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "2",
				"categoryId(2)": "3",
			},
		},
		{
			Name: "can find items if params contains 3 categoryIds of length 10",
			Params: map[string]string{
				"categoryId(0)": "1234567890",
				"categoryId(1)": "9876543210",
				"categoryId(2)": "8976543210",
			},
		},
		{
			Name: "returns error if params contains 1 categoryId of length 11, 2 categoryIds of length 1",
			Params: map[string]string{
				"categoryId(0)": "12345678901",
				"categoryId(1)": "1",
				"categoryId(2)": "2",
			},
			Err: ebay.ErrInvalidCategoryIDLength,
		},
		{
			Name: "returns error if params contains 2 categoryIds of length 1, 1 middle categoryId of length 11",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "12345678901",
				"categoryId(2)": "2",
			},
			Err: ebay.ErrInvalidCategoryIDLength,
		},
		{
			Name: "returns error if params contains 2 categoryIds of length 1, 1 categoryId of length 11",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "2",
				"categoryId(2)": "12345678901",
			},
			Err: ebay.ErrInvalidCategoryIDLength,
		},
		{
			Name:   "returns error if params contains numbered, invalid categoryId",
			Params: map[string]string{"categoryId(0)": "a"},
			Err:    fmt.Errorf("%w: %s: %w", ebay.ErrInvalidCategoryID, `strconv.Atoi: parsing "a"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains 1 valid, 1 invalid categoryId",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "a",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidCategoryID, `strconv.Atoi: parsing "a"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains 4 categoryIds",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "2",
				"categoryId(2)": "3",
				"categoryId(3)": "4",
			},
			Err: ebay.ErrMaxCategoryIDs,
		},
	}
	keywordsTCs = []findItemsTestCase{
		{
			Name:   "can find items if params contains keywords of length 2",
			Params: map[string]string{"keywords": generateStringWithLen(2, true)},
		},
		{
			Name:   "can find items if params contains keywords of length 12",
			Params: map[string]string{"keywords": generateStringWithLen(12, true)},
		},
		{
			Name:   "can find items if params contains keywords of length 350",
			Params: map[string]string{"keywords": generateStringWithLen(350, true)},
		},
		{
			Name:   "returns error if params contains empty keywords",
			Params: map[string]string{"keywords": ""},
			Err:    ebay.ErrInvalidKeywordsLength,
		},
		{
			Name:   "returns error if params contains keywords of length 1",
			Params: map[string]string{"keywords": generateStringWithLen(1, true)},
			Err:    ebay.ErrInvalidKeywordsLength,
		},
		{
			Name:   "returns error if params contains keywords of length 351",
			Params: map[string]string{"keywords": generateStringWithLen(351, true)},
			Err:    ebay.ErrInvalidKeywordsLength,
		},
		{
			Name:   "can find items if params contains 1 keyword of length 2",
			Params: map[string]string{"keywords": generateStringWithLen(2, false)},
		},
		{
			Name:   "can find items if params contains 1 keyword of length 98",
			Params: map[string]string{"keywords": generateStringWithLen(98, false)},
		},
		{
			Name:   "returns error if params contains 1 keyword of length 99",
			Params: map[string]string{"keywords": generateStringWithLen(99, false)},
			Err:    ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains 2 keywords of length 1",
			Params: map[string]string{
				"keywords": generateStringWithLen(1, false) + "," + generateStringWithLen(1, false),
			},
		},
		{
			Name: "can find items if params contains 2 keywords of length 98",
			Params: map[string]string{
				"keywords": generateStringWithLen(98, false) + "," + generateStringWithLen(98, false),
			},
		},
		{
			Name: "can find items if params contains keywords of length 1 and 98",
			Params: map[string]string{
				"keywords": generateStringWithLen(1, false) + "," + generateStringWithLen(98, false),
			},
		},
		{
			Name: "can find items if params contains keywords of length 98 and 1",
			Params: map[string]string{
				"keywords": generateStringWithLen(98, false) + "," + generateStringWithLen(1, false),
			},
		},
		{
			Name: "returns error if params contains 2 keywords of length 99",
			Params: map[string]string{
				"keywords": generateStringWithLen(99, false) + "," + generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "returns error if params contains keywords of length 1 and 99",
			Params: map[string]string{
				"keywords": generateStringWithLen(1, false) + "," + generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "returns error if params contains keywords of length 99 and 1",
			Params: map[string]string{
				"keywords": generateStringWithLen(99, false) + "," + generateStringWithLen(1, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=baseball card",
			Params: map[string]string{
				"keywords": "baseball card",
			},
		},
		{
			Name: "returns error if params contains space-separated keywords, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "baseball " + generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=baseball,card",
			Params: map[string]string{
				"keywords": "baseball,card",
			},
		},
		{
			Name: "returns error if params contains comma-separated keywords, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "baseball," + generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=(baseball,card)",
			Params: map[string]string{
				"keywords": "(baseball,card)",
			},
		},
		{
			Name: "returns error if params contains comma-separated, parenthesis keywords, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "(baseball," + generateStringWithLen(99, false) + ")",
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: `can find items if params contains keywords="baseball card"`,
			Params: map[string]string{
				"keywords": `"baseball card"`,
			},
		},
		{
			Name: "returns error if params contains double-quoted keywords, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": `"baseball ` + generateStringWithLen(99, false) + `"`,
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=baseball -autograph",
			Params: map[string]string{
				"keywords": "baseball -autograph",
			},
		},
		{
			Name: "returns error if params contains 1 keyword with minus sign before it, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "baseball -" + generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=baseball -(autograph,card,star)",
			Params: map[string]string{
				"keywords": "baseball -(autograph,card,star)",
			},
		},
		{
			Name: "returns error if params contains minus sign before group of keywords, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "baseball -(autograph,card," + generateStringWithLen(99, false) + ")",
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=baseball*",
			Params: map[string]string{
				"keywords": "baseball*",
			},
		},
		{
			Name: "returns error if params contains keyword of length 99 and asterisk",
			Params: map[string]string{
				"keywords": generateStringWithLen(99, false) + "*",
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=@1 baseball autograph card",
			Params: map[string]string{
				"keywords": "@1 baseball autograph card",
			},
		},
		{
			Name: "returns error if params contains @ with group of keywords, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "@1 baseball autograph " + generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=@1 baseball autograph card +star",
			Params: map[string]string{
				"keywords": "@1 baseball autograph card +star",
			},
		},
		{
			Name: "returns error if params contains @ with group of keywords, plus sign, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "@1 baseball autograph card +" + generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains keywords=ap* ip*",
			Params: map[string]string{
				"keywords": "ap* ip*",
			},
		},
		{
			Name: "returns error if params contains 2 asterisk keyword groups, 1 keyword of length 99",
			Params: map[string]string{
				"keywords": "ap* " + generateStringWithLen(99, false) + "*",
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
	}
	categoryIDKeywordsTCs = []findItemsTestCase{
		{
			Name: "can find items if params contains 1 categoryId of length 1, keywords of length 2",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(2, true),
			},
		},
		{
			Name: "can find items if params contains 2 categoryIds of length 1, keywords of length 2",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "2",
				"keywords":      generateStringWithLen(2, true),
			},
		},
		{
			Name: "returns error if params contains empty categoryId, keywords of length 2",
			Params: map[string]string{
				"categoryId": "",
				"keywords":   generateStringWithLen(2, true),
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidCategoryID, `strconv.Atoi: parsing ""`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains 4 categoryIds, keywords of length 2",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "2",
				"categoryId(2)": "3",
				"categoryId(3)": "4",
				"keywords":      generateStringWithLen(2, true),
			},
			Err: ebay.ErrMaxCategoryIDs,
		},
		{
			Name: "can find items if params contains 1 categoryId of length 1, 2 keywords of length 1",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(1, false) + "," + generateStringWithLen(1, false),
			},
		},
		{
			Name:   "returns error if params contains 1 categoryId of length 1, empty keywords",
			Params: map[string]string{"categoryId": "1", "keywords": ""},
			Err:    ebay.ErrInvalidKeywordsLength,
		},
		{
			Name: "returns error if params contains 1 categoryId of length 1, 1 keyword of length 99",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(99, false),
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
	}
	missingSearchParamTCs = []findItemsTestCase{
		{
			Name:   "returns error if params does not contain ",
			Params: map[string]string{},
		},
		{
			Name:   "returns error if params contains Global ID but not ",
			Params: map[string]string{"Global-ID": "EBAY-AT"},
		},
		{
			Name: "returns error if params contains non-numbered itemFilter but not ",
			Params: map[string]string{
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "returns error if params contains numbered itemFilter but not ",
			Params: map[string]string{
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
			},
		},
		{
			Name:   "returns error if params contains outputSelector but not ",
			Params: map[string]string{"outputSelector": "AspectHistogram"},
		},
		{
			Name: "returns error if params contains affiliate but not ",
			Params: map[string]string{
				"affiliate.customId":     "123",
				"affiliate.geoTargeting": "true",
				"affiliate.networkId":    "2",
				"affiliate.trackingId":   "123",
			},
		},
		{
			Name:   "returns error if params contains buyerPostalCode but not ",
			Params: map[string]string{"buyerPostalCode": "111"},
		},
		{
			Name: "returns error if params contains paginationInput but not ",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "1",
				"paginationInput.pageNumber":     "1",
			},
		},
		{
			Name:   "returns error if params contains sortOrder but not ",
			Params: map[string]string{"sortOrder": "BestMatch"},
		},
	}
	aspectFilterMissingSearchParamTCs = []findItemsTestCase{
		{
			Name: "returns error if params contains non-numbered aspectFilter but not ",
			Params: map[string]string{
				"aspectFilter.aspectName":      "Size",
				"aspectFilter.aspectValueName": "10",
			},
		},
		{
			Name: "returns error if params contains numbered aspectFilter but not ",
			Params: map[string]string{
				"aspectFilter(0).aspectName":      "Size",
				"aspectFilter(0).aspectValueName": "10",
			},
		},
	}
	easternTime = time.FixedZone("EasternTime", -5*60*60)
	testCases   = []findItemsTestCase{
		{
			Name:   "can find items if params contains Global ID EBAY-AT",
			Params: map[string]string{"Global-ID": "EBAY-AT"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-AU",
			Params: map[string]string{"Global-ID": "EBAY-AU"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-CH",
			Params: map[string]string{"Global-ID": "EBAY-CH"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-DE",
			Params: map[string]string{"Global-ID": "EBAY-DE"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-ENCA",
			Params: map[string]string{"Global-ID": "EBAY-ENCA"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-ES",
			Params: map[string]string{"Global-ID": "EBAY-ES"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-FR",
			Params: map[string]string{"Global-ID": "EBAY-FR"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-FRBE",
			Params: map[string]string{"Global-ID": "EBAY-FRBE"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-FRCA",
			Params: map[string]string{"Global-ID": "EBAY-FRCA"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-GB",
			Params: map[string]string{"Global-ID": "EBAY-GB"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-HK",
			Params: map[string]string{"Global-ID": "EBAY-HK"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-IE",
			Params: map[string]string{"Global-ID": "EBAY-IE"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-IN",
			Params: map[string]string{"Global-ID": "EBAY-IN"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-IT",
			Params: map[string]string{"Global-ID": "EBAY-IT"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-MOTOR",
			Params: map[string]string{"Global-ID": "EBAY-MOTOR"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-MY",
			Params: map[string]string{"Global-ID": "EBAY-MY"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-NL",
			Params: map[string]string{"Global-ID": "EBAY-NL"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-NLBE",
			Params: map[string]string{"Global-ID": "EBAY-NLBE"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-PH",
			Params: map[string]string{"Global-ID": "EBAY-PH"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-PL",
			Params: map[string]string{"Global-ID": "EBAY-PL"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-SG",
			Params: map[string]string{"Global-ID": "EBAY-SG"},
		},
		{
			Name:   "can find items if params contains Global ID EBAY-US",
			Params: map[string]string{"Global-ID": "EBAY-US"},
		},
		{
			Name:   "returns error if params contains Global ID EBAY-ZZZ",
			Params: map[string]string{"Global-ID": "EBAY-ZZZ"},
			Err:    fmt.Errorf("%w: %q", ebay.ErrInvalidGlobalID, "EBAY-ZZZ"),
		},
		{
			Name: "can find items by itemFilter.name, value",
			Params: map[string]string{
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items by itemFilter.name, value(0), value(1)",
			Params: map[string]string{
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value(0)": "1",
				"itemFilter.value(1)": "2",
			},
		},
		{
			Name: "can find items by itemFilter.name, value, paramName, paramValue",
			Params: map[string]string{
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "5.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "EUR",
			},
		},
		{
			Name:   "returns error if params contains itemFilter.name but not value",
			Params: map[string]string{"itemFilter.name": "BestOfferOnly"},
			Err:    fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "itemFilter.value"),
		},
		{
			// itemFilter.value(1) will be ignored because indexing does not start at 0.
			Name: "returns error if params contains itemFilter.name, value(1)",
			Params: map[string]string{
				"itemFilter.name":     "BestOfferOnly",
				"itemFilter.value(1)": "true",
			},
			Err: fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "itemFilter.value"),
		},
		{
			// itemFilter.value(1) will be ignored because indexing does not start at 0.
			// Therefore, only itemFilter.value is considered and this becomes a non-numbered itemFilter.
			Name: "can find items by itemFilter.name, value, value(1)",
			Params: map[string]string{
				"itemFilter.name":     "BestOfferOnly",
				"itemFilter.value":    "true",
				"itemFilter.value(1)": "true",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter.name param is found before other itemFilter params.
			Name:   "can find items if params contains itemFilter.value only",
			Params: map[string]string{"itemFilter.value": "true"},
		},
		{
			// The itemFilter will be ignored if no itemFilter.name param is found before other itemFilter params.
			Name:   "can find items if params contains itemFilter.paramName only",
			Params: map[string]string{"itemFilter.paramName": "Currency"},
		},
		{
			// The itemFilter will be ignored if no itemFilter.name param is found before other itemFilter params.
			Name:   "can find items if params contains itemFilter.paramValue only",
			Params: map[string]string{"itemFilter.paramValue": "EUR"},
		},
		{
			Name: "returns error if params contains itemFilter.paramName but not paramValue",
			Params: map[string]string{
				"itemFilter.name":      "MaxPrice",
				"itemFilter.value":     "5.0",
				"itemFilter.paramName": "Currency",
			},
			Err: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contains itemFilter.paramValue but not paramName",
			Params: map[string]string{
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "5.0",
				"itemFilter.paramValue": "EUR",
			},
			Err: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contain numbered and non-numbered itemFilter syntax types",
			Params: map[string]string{
				"itemFilter.name":     "BestOfferOnly",
				"itemFilter.value":    "true",
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "5.0",
			},
			Err: ebay.ErrInvalidIndexSyntax,
		},
		{
			Name: "returns error if params contain itemFilter.name, value, value(0)",
			Params: map[string]string{
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value":    "1",
				"itemFilter.value(0)": "2",
			},
			Err: ebay.ErrInvalidIndexSyntax,
		},
		{
			Name: "returns error if params contain itemFilter(0).name, value, value(0)",
			Params: map[string]string{
				"itemFilter(0).name":     "ExcludeCategory",
				"itemFilter(0).value":    "1",
				"itemFilter(0).value(0)": "2",
			},
			Err: ebay.ErrInvalidIndexSyntax,
		},
		{
			Name: "can find items by itemFilter(0).name, value",
			Params: map[string]string{
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
			},
		},
		{
			Name: "can find items by itemFilter(0).name, value(0), value(1)",
			Params: map[string]string{
				"itemFilter(0).name":     "ExcludeCategory",
				"itemFilter(0).value(0)": "1",
				"itemFilter(0).value(1)": "2",
			},
		},
		{
			Name: "can find items by itemFilter(0).name, value, paramName, and paramValue",
			Params: map[string]string{
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "EUR",
			},
		},
		{
			Name: "can find items by 2 basic, numbered itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "5.0",
			},
		},
		{
			Name: "can find items by 1st advanced, numbered and 2nd basic, numbered itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "BestOfferOnly",
				"itemFilter(1).value":      "true",
			},
		},
		{
			Name: "can find items by 1st basic, numbered and 2nd advanced, numbered itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":       "BestOfferOnly",
				"itemFilter(0).value":      "true",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "EUR",
			},
		},
		{
			Name: "can find items by 2 advanced, numbered itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "1.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "EUR",
			},
		},
		{
			Name:   "returns error if params contains itemFilter(0).name but not value",
			Params: map[string]string{"itemFilter(0).name": "BestOfferOnly"},
			Err:    fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "itemFilter(0).value"),
		},
		{
			// itemFilter(0).value(1) will be ignored because indexing does not start at 0.
			Name: "returns error if params contains itemFilter(0).name, value(1)",
			Params: map[string]string{
				"itemFilter(0).name":     "BestOfferOnly",
				"itemFilter(0).value(1)": "true",
			},
			Err: fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "itemFilter(0).value"),
		},
		{
			// itemFilter(0).value(1) will be ignored because indexing does not start at 0.
			// Therefore, only itemFilter(0).value is considered and this becomes a numbered itemFilter.
			Name: "can find items by itemFilter(0).name, value, value(1)",
			Params: map[string]string{
				"itemFilter(0).name":     "BestOfferOnly",
				"itemFilter(0).value":    "true",
				"itemFilter(0).value(1)": "true",
			},
		},
		{
			// The itemFilter will be ignored if no itemFilter(0).name param is found before other itemFilter params.
			Name:   "can find items if params contains itemFilter(0).value only",
			Params: map[string]string{"itemFilter(0).value": "true"},
		},
		{
			// The itemFilter will be ignored if no itemFilter(0).name param is found before other itemFilter params.
			Name:   "can find items if params contains itemFilter(0).paramName only",
			Params: map[string]string{"itemFilter(0).paramName": "Currency"},
		},
		{
			// The itemFilter will be ignored if no itemFilter(0).name param is found before other itemFilter params.
			Name:   "can find items if params contains itemFilter(0).paramValue only",
			Params: map[string]string{"itemFilter(0).paramValue": "EUR"},
		},
		{
			Name: "returns error if params contains itemFilter(0).paramName but not paramValue",
			Params: map[string]string{
				"itemFilter(0).name":      "MaxPrice",
				"itemFilter(0).value":     "5.0",
				"itemFilter(0).paramName": "Currency",
			},
			Err: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contains itemFilter(0).paramValue but not paramName",
			Params: map[string]string{
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramValue": "EUR",
			},
			Err: ebay.ErrIncompleteItemFilterParam,
		},
		{
			Name: "returns error if params contains non-numbered, unsupported itemFilter name",
			Params: map[string]string{
				"itemFilter.name":  "UnsupportedFilter",
				"itemFilter.value": "true",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrUnsupportedItemFilterType, "UnsupportedFilter"),
		},
		{
			Name: "returns error if params contains numbered, unsupported itemFilter name",
			Params: map[string]string{
				"itemFilter(0).name":  "UnsupportedFilter",
				"itemFilter(0).value": "true",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrUnsupportedItemFilterType, "UnsupportedFilter"),
		},
		{
			Name: "returns error if params contains numbered supported and unsupported itemFilter names",
			Params: map[string]string{
				"itemFilter(0).name":  "BestOfferOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "UnsupportedFilter",
				"itemFilter(1).value": "true",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrUnsupportedItemFilterType, "UnsupportedFilter"),
		},
		{
			Name: "can find items if params contains AuthorizedSellerOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "AuthorizedSellerOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains AuthorizedSellerOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "AuthorizedSellerOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains AuthorizedSellerOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "AuthorizedSellerOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains valid AvailableTo itemFilter",
			Params: map[string]string{
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "US",
			},
		},
		{
			Name: "returns error if params contains AvailableTo itemFilter with lowercase characters",
			Params: map[string]string{
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "us",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCountryCode, "us"),
		},
		{
			Name: "returns error if params contains AvailableTo itemFilter with 1 uppercase character",
			Params: map[string]string{
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "U",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCountryCode, "U"),
		},
		{
			Name: "returns error if params contains AvailableTo itemFilter with 3 uppercase character",
			Params: map[string]string{
				"itemFilter.name":  "AvailableTo",
				"itemFilter.value": "USA",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCountryCode, "USA"),
		},
		{
			Name: "can find items if params contains BestOfferOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains BestOfferOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains BestOfferOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "BestOfferOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains CharityOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "CharityOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains CharityOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "CharityOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains CharityOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "CharityOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition name",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "dirty",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 1000",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 1500",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1500",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 1750",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1750",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2000",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2010",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2010",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2020",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2020",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2030",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2030",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2500",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2500",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 2750",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "2750",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 3000",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "3000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 4000",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "4000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 5000",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "5000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 6000",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "6000",
			},
		},
		{
			Name: "can find items if params contains Condition itemFilter with condition ID 7000",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "7000",
			},
		},
		{
			Name: "returns error if params contains Condition itemFilter with condition ID 1",
			Params: map[string]string{
				"itemFilter.name":  "Condition",
				"itemFilter.value": "1",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCondition, "1"),
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID AUD",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "AUD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID CAD",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "CAD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID CHF",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "CHF",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID CNY",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "CNY",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID EUR",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "EUR",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID GBP",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "GBP",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID HKD",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "HKD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID INR",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "INR",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID MYR",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "MYR",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID PHP",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "PHP",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID PLN",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "PLN",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID SEK",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "SEK",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID SGD",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "SGD",
			},
		},
		{
			Name: "can find items if params contains Currency itemFilter with currency ID TWD",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "TWD",
			},
		},
		{
			Name: "returns error if params contains Currency itemFilter with currency ID ZZZ",
			Params: map[string]string{
				"itemFilter.name":  "Currency",
				"itemFilter.value": "ZZZ",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains EndTimeFrom itemFilter with future timestamp",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": time.Now().Add(5 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains EndTimeFrom itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": "not a timestamp",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains EndTimeFrom itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).In(easternTime).Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).In(easternTime).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains EndTimeFrom itemFilter with past timestamp",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeFrom",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains EndTimeTo itemFilter with future timestamp",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": time.Now().Add(5 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains EndTimeTo itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": "not a timestamp",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains EndTimeTo itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": time.Now().Add(1 * time.Second).In(easternTime).Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).In(easternTime).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains EndTimeTo itemFilter with past timestamp",
			Params: map[string]string{
				"itemFilter.name":  "EndTimeTo",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains ExcludeAutoPay itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeAutoPay",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains ExcludeAutoPay itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeAutoPay",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains ExcludeAutoPay itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeAutoPay",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains ExcludeCategory itemFilter with category ID 0",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains ExcludeCategory itemFilter with category ID 5",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains ExcludeCategory itemFilter with unparsable category ID",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "not a category ID",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "not a category ID", 0),
		},
		{
			Name: "returns error if params contains ExcludeCategory itemFilter with category ID -1",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeCategory",
				"itemFilter.value": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains ExcludeCategory itemFilter with category IDs 0 and 1",
			Params: map[string]string{
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "1",
			},
		},
		{
			Name: "returns error if params contains ExcludeCategory itemFilter with category IDs 0 and -1",
			Params: map[string]string{
				"itemFilter.name":     "ExcludeCategory",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name:   "can find items if params contains ExcludeCategory itemFilter with 25 category IDs",
			Params: generateFilterParams("ExcludeCategory", 25),
		},
		{
			Name:   "returns error if params contains ExcludeCategory itemFilter with 26 category IDs",
			Params: generateFilterParams("ExcludeCategory", 26),
			Err:    ebay.ErrMaxExcludeCategories,
		},
		{
			Name: "can find items if params contains ExcludeSeller itemFilter with seller ID 0",
			Params: map[string]string{
				"itemFilter.name":  "ExcludeSeller",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains ExcludeSeller itemFilter with seller IDs 0 and 1",
			Params: map[string]string{
				"itemFilter.name":     "ExcludeSeller",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "1",
			},
		},
		{
			Name: "returns error if params contains ExcludeSeller and Seller itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":  "ExcludeSeller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "Seller",
				"itemFilter(1).value": "0",
			},
			Err: ebay.ErrExcludeSellerCannotBeUsedWithSellers,
		},
		{
			Name: "returns error if params contains ExcludeSeller and TopRatedSellerOnly itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":  "ExcludeSeller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "TopRatedSellerOnly",
				"itemFilter(1).value": "true",
			},
			Err: ebay.ErrExcludeSellerCannotBeUsedWithSellers,
		},
		{
			Name:   "can find items if params contains ExcludeSeller itemFilter with 100 seller IDs",
			Params: generateFilterParams("ExcludeSeller", 100),
		},
		{
			Name:   "returns error if params contains ExcludeSeller itemFilter with 101 seller IDs",
			Params: generateFilterParams("ExcludeSeller", 101),
			Err:    ebay.ErrMaxExcludeSellers,
		},
		{
			Name: "can find items if params contains ExpeditedShippingType itemFilter.value=Expedited",
			Params: map[string]string{
				"itemFilter.name":  "ExpeditedShippingType",
				"itemFilter.value": "Expedited",
			},
		},
		{
			Name: "can find items if params contains ExpeditedShippingType itemFilter.value=OneDayShipping",
			Params: map[string]string{
				"itemFilter.name":  "ExpeditedShippingType",
				"itemFilter.value": "OneDayShipping",
			},
		},
		{
			Name: "returns error if params contains ExpeditedShippingType itemFilter with invalid shipping type",
			Params: map[string]string{
				"itemFilter.name":  "ExpeditedShippingType",
				"itemFilter.value": "InvalidShippingType",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidExpeditedShippingType, "InvalidShippingType"),
		},
		{
			Name: "can find items if params contains FeedbackScoreMax itemFilter with max 0",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMax itemFilter with max 5",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains FeedbackScoreMax itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMax itemFilter with max -1",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMax",
				"itemFilter.value": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains FeedbackScoreMin itemFilter with max 0",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin itemFilter with max 5",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains FeedbackScoreMin itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "not a minimum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin itemFilter with max -1",
			Params: map[string]string{
				"itemFilter.name":  "FeedbackScoreMin",
				"itemFilter.value": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 1 and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 0 and max 1",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 10 and min 5",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 5 and max 10",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "10",
			},
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 0 and unparsable min",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "not a minimum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with unparsable min and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 0 and min -1",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min -1 and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 0 and min 1",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "1",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 1 and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max 5 and min 10",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "10",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 10 and max 5",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "5",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with unparsable max and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 0 and unparsable max",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with max -1 and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMax",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "FeedbackScoreMin",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains FeedbackScoreMin/FeedbackScoreMax itemFilters with min 0 and max -1",
			Params: map[string]string{
				"itemFilter(0).name":  "FeedbackScoreMin",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "FeedbackScoreMax",
				"itemFilter(1).value": "-1",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "FeedbackScoreMax", "FeedbackScoreMin"),
		},
		{
			Name: "can find items if params contains FreeShippingOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "FreeShippingOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains FreeShippingOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "FreeShippingOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains FreeShippingOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "FreeShippingOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains HideDuplicateItems itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "HideDuplicateItems",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains HideDuplicateItems itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "HideDuplicateItems",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains HideDuplicateItems itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "HideDuplicateItems",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-AT",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-AT",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-AU",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-AU",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-CH",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-CH",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-DE",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-DE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-ENCA",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-ENCA",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-ES",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-ES",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-FR",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-FR",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-FRBE",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-FRBE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-FRCA",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-FRCA",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-GB",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-GB",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-HK",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-HK",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-IE",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-IE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-IN",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-IN",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-IT",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-IT",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-MOTOR",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-MOTOR",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-MY",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-MY",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-NL",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-NL",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-NLBE",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-NLBE",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-PH",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-PH",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-PL",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-PL",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-SG",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-SG",
			},
		},
		{
			Name: "can find items if params contains ListedIn itemFilter with Global ID EBAY-US",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-US",
			},
		},
		{
			Name: "returns error if params contains ListedIn itemFilter with Global ID EBAY-ZZZ",
			Params: map[string]string{
				"itemFilter.name":  "ListedIn",
				"itemFilter.value": "EBAY-ZZZ",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidGlobalID, "EBAY-ZZZ"),
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type Auction",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "Auction",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type AuctionWithBIN",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "AuctionWithBIN",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type Classified",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "Classified",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type FixedPrice",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "FixedPrice",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type StoreInventory",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "StoreInventory",
			},
		},
		{
			Name: "can find items if params contains ListingType itemFilter with listing type All",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "All",
			},
		},
		{
			Name: "returns error if params contains ListingType itemFilter with invalid listing type",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "not a listing type",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidListingType, "not a listing type"),
		},
		{
			Name: "returns error if params contains ListingType itemFilters with All and Auction listing types",
			Params: map[string]string{
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "All",
				"itemFilter.value(1)": "Auction",
			},
			Err: ebay.ErrInvalidAllListingType,
		},
		{
			Name: "returns error if params contains ListingType itemFilters with StoreInventory and All listing types",
			Params: map[string]string{
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "StoreInventory",
				"itemFilter.value(1)": "All",
			},
			Err: ebay.ErrInvalidAllListingType,
		},
		{
			Name: "returns error if params contains ListingType itemFilters with 2 Auction listing types",
			Params: map[string]string{
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "Auction",
				"itemFilter.value(1)": "Auction",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrDuplicateListingType, "Auction"),
		},
		{
			Name: "returns error if params contains ListingType itemFilters with 2 StoreInventory listing types",
			Params: map[string]string{
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "StoreInventory",
				"itemFilter.value(1)": "StoreInventory",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrDuplicateListingType, "StoreInventory"),
		},
		{
			Name: "returns error if params contains ListingType itemFilters with Auction and AuctionWithBIN listing types",
			Params: map[string]string{
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "Auction",
				"itemFilter.value(1)": "AuctionWithBIN",
			},
			Err: ebay.ErrInvalidAuctionListingTypes,
		},
		{
			Name: "returns error if params contains ListingType itemFilters with AuctionWithBIN and Auction listing types",
			Params: map[string]string{
				"itemFilter.name":     "ListingType",
				"itemFilter.value(0)": "AuctionWithBIN",
				"itemFilter.value(1)": "Auction",
			},
			Err: ebay.ErrInvalidAuctionListingTypes,
		},
		{
			Name: "can find items if params contains LocalPickupOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "LocalPickupOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains LocalPickupOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "LocalPickupOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains LocalPickupOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "LocalPickupOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains LocalSearchOnly itemFilter.value=true, buyerPostalCode, and MaxDistance",
			Params: map[string]string{
				"buyerPostalCode":     "123",
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains LocalSearchOnly itemFilter.value=false, buyerPostalCode, and MaxDistance",
			Params: map[string]string{
				"buyerPostalCode":     "123",
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "false",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains LocalSearchOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"buyerPostalCode":     "123",
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "123",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "returns error if params contains LocalSearchOnly itemFilter but no buyerPostalCode",
			Params: map[string]string{
				"itemFilter(0).name":  "LocalSearchOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "MaxDistance",
				"itemFilter(1).value": "5",
			},
			Err: ebay.ErrBuyerPostalCodeMissing,
		},
		{
			Name: "returns error if params contains LocalSearchOnly itemFilter but no MaxDistance itemFilter",
			Params: map[string]string{
				"buyerPostalCode":  "123",
				"itemFilter.name":  "LocalSearchOnly",
				"itemFilter.value": "true",
			},
			Err: ebay.ErrMaxDistanceMissing,
		},
		{
			Name: "can find items if params contains valid LocatedIn itemFilter",
			Params: map[string]string{
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "US",
			},
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with lowercase characters",
			Params: map[string]string{
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "us",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCountryCode, "us"),
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with 1 uppercase character",
			Params: map[string]string{
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "U",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCountryCode, "U"),
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with 3 uppercase character",
			Params: map[string]string{
				"itemFilter.name":  "LocatedIn",
				"itemFilter.value": "USA",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCountryCode, "USA"),
		},
		{
			Name: "can find items if params contains LocatedIn itemFilter with 25 country codes",
			Params: map[string]string{
				"itemFilter.name":      "LocatedIn",
				"itemFilter.value(0)":  "AA",
				"itemFilter.value(1)":  "AB",
				"itemFilter.value(2)":  "AC",
				"itemFilter.value(3)":  "AD",
				"itemFilter.value(4)":  "AE",
				"itemFilter.value(5)":  "AF",
				"itemFilter.value(6)":  "AG",
				"itemFilter.value(7)":  "AH",
				"itemFilter.value(8)":  "AI",
				"itemFilter.value(9)":  "AJ",
				"itemFilter.value(10)": "AK",
				"itemFilter.value(11)": "AL",
				"itemFilter.value(12)": "AM",
				"itemFilter.value(13)": "AN",
				"itemFilter.value(14)": "AO",
				"itemFilter.value(15)": "AP",
				"itemFilter.value(16)": "AQ",
				"itemFilter.value(17)": "AR",
				"itemFilter.value(18)": "AS",
				"itemFilter.value(19)": "AT",
				"itemFilter.value(20)": "AU",
				"itemFilter.value(21)": "AV",
				"itemFilter.value(22)": "AW",
				"itemFilter.value(23)": "AX",
				"itemFilter.value(24)": "AY",
			},
		},
		{
			Name: "returns error if params contains LocatedIn itemFilter with 26 country codes",
			Params: map[string]string{
				"itemFilter.name":      "LocatedIn",
				"itemFilter.value(0)":  "AA",
				"itemFilter.value(1)":  "AB",
				"itemFilter.value(2)":  "AC",
				"itemFilter.value(3)":  "AD",
				"itemFilter.value(4)":  "AE",
				"itemFilter.value(5)":  "AF",
				"itemFilter.value(6)":  "AG",
				"itemFilter.value(7)":  "AH",
				"itemFilter.value(8)":  "AI",
				"itemFilter.value(9)":  "AJ",
				"itemFilter.value(10)": "AK",
				"itemFilter.value(11)": "AL",
				"itemFilter.value(12)": "AM",
				"itemFilter.value(13)": "AN",
				"itemFilter.value(14)": "AO",
				"itemFilter.value(15)": "AP",
				"itemFilter.value(16)": "AQ",
				"itemFilter.value(17)": "AR",
				"itemFilter.value(18)": "AS",
				"itemFilter.value(19)": "AT",
				"itemFilter.value(20)": "AU",
				"itemFilter.value(21)": "AV",
				"itemFilter.value(22)": "AW",
				"itemFilter.value(23)": "AX",
				"itemFilter.value(24)": "AY",
				"itemFilter.value(25)": "AZ",
			},
			Err: ebay.ErrMaxLocatedIns,
		},
		{
			Name: "can find items if params contains LotsOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "LotsOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains LotsOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "LotsOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains LotsOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "LotsOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains MaxBids itemFilter with max 0",
			Params: map[string]string{
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains MaxBids itemFilter with max 5",
			Params: map[string]string{
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MaxBids itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MaxBids itemFilter with max -1",
			Params: map[string]string{
				"itemFilter.name":  "MaxBids",
				"itemFilter.value": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains MinBids itemFilter with max 0",
			Params: map[string]string{
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids itemFilter with max 5",
			Params: map[string]string{
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MinBids itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "not a minimum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids itemFilter with max -1",
			Params: map[string]string{
				"itemFilter.name":  "MinBids",
				"itemFilter.value": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with max 1 and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with min 0 and max 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with max and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with min and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with max 10 and min 5",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains MinBids/MaxBids itemFilters with min 5 and max 10",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "10",
			},
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 0 and unparsable min",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "not a minimum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with unparsable min and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 0 and min -1",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "-1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min -1 and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 0 and min 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "1",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 1 and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max 5 and min 10",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "10",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 10 and max 5",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "5",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with unparsable max and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 0 and unparsable max",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with max -1 and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxBids",
				"itemFilter(0).value": "-1",
				"itemFilter(1).name":  "MinBids",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "-1", 0),
		},
		{
			Name: "returns error if params contains MinBids/MaxBids itemFilters with min 0 and max -1",
			Params: map[string]string{
				"itemFilter(0).name":  "MinBids",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxBids",
				"itemFilter(1).value": "-1",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxBids", "MinBids"),
		},
		{
			Name: "can find items if params contains MaxDistance itemFilter with max 5 and buyerPostalCode",
			Params: map[string]string{
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "can find items if params contains MaxDistance itemFilter with max 6 and buyerPostalCode",
			Params: map[string]string{
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "6",
			},
		},
		{
			Name: "returns error if params contains MaxDistance itemFilter with unparsable max",
			Params: map[string]string{
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "not a maximum", 5),
		},
		{
			Name: "returns error if params contains MaxDistance itemFilter with max 4 and buyerPostalCode",
			Params: map[string]string{
				"buyerPostalCode":  "123",
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "4",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "4", 5),
		},
		{
			Name: "returns error if params contains MaxDistance itemFilter with max 5 but no buyerPostalCode",
			Params: map[string]string{
				"itemFilter.name":  "MaxDistance",
				"itemFilter.value": "5",
			},
			Err: ebay.ErrBuyerPostalCodeMissing,
		},
		{
			Name: "can find items if params contains MaxHandlingTime itemFilter with max 1",
			Params: map[string]string{
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "1",
			},
		},
		{
			Name: "can find items if params contains MaxHandlingTime itemFilter with max 5",
			Params: map[string]string{
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MaxHandlingTime itemFilter with max 0",
			Params: map[string]string{
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MaxHandlingTime itemFilter with unparsable max",
			Params: map[string]string{
				"itemFilter.name":  "MaxHandlingTime",
				"itemFilter.value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "not a maximum", 1),
		},
		{
			Name: "can find items if params contains MaxPrice itemFilter with max 0.0",
			Params: map[string]string{
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MaxPrice itemFilter with max 5.0",
			Params: map[string]string{
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "5.0",
			},
		},
		{
			Name: "can find items if params contains MaxPrice itemFilter with max 0.0, paramName Currency, and paramValue EUR",
			Params: map[string]string{
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with unparsable max",
			Params: map[string]string{
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w",
				ebay.ErrInvalidPrice, `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with max -1.0",
			Params: map[string]string{
				"itemFilter.name":  "MaxPrice",
				"itemFilter.value": "-1.0",
			},
			Err: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with max 0.0, paramName NotCurrency, and paramValue EUR",
			Params: map[string]string{
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "NotCurrency",
				"itemFilter.paramValue": "EUR",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidPriceParamName, "NotCurrency"),
		},
		{
			Name: "returns error if params contains MaxPrice itemFilter with max 0.0, paramName Currency, and paramValue ZZZ",
			Params: map[string]string{
				"itemFilter.name":       "MaxPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "ZZZ",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains MinPrice itemFilter with max 0.0",
			Params: map[string]string{
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice itemFilter with max 5.0",
			Params: map[string]string{
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "5.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice itemFilter with max 0.0, paramName Currency, and paramValue EUR",
			Params: map[string]string{
				"itemFilter.name":       "MinPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "EUR",
			},
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with unparsable max",
			Params: map[string]string{
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w",
				ebay.ErrInvalidPrice, `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max -1.0",
			Params: map[string]string{
				"itemFilter.name":  "MinPrice",
				"itemFilter.value": "-1.0",
			},
			Err: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 0.0, paramName NotCurrency, and paramValue EUR",
			Params: map[string]string{
				"itemFilter.name":       "MinPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "NotCurrency",
				"itemFilter.paramValue": "EUR",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidPriceParamName, "NotCurrency"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 0.0, paramName Currency, and paramValue ZZZ",
			Params: map[string]string{
				"itemFilter.name":       "MinPrice",
				"itemFilter.value":      "0.0",
				"itemFilter.paramName":  "Currency",
				"itemFilter.paramValue": "ZZZ",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with max 1.0 and min 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "1.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with min 0.0 and max 1.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "1.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with max and min 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with min and max 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with max 10.0 and min 5.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "10.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "5.0",
			},
		},
		{
			Name: "can find items if params contains MinPrice/MaxPrice itemFilters with min 5.0 and max 10.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "5.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "10.0",
			},
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 0.0 and unparsable min",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "not a minimum",
			},
			Err: fmt.Errorf("%w: %s: %w",
				ebay.ErrInvalidPrice, `strconv.ParseFloat: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with unparsable min and max 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
			Err: fmt.Errorf("%w: %s: %w",
				ebay.ErrInvalidPrice, `strconv.ParseFloat: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 0.0 and min -1.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "-1.0",
			},
			Err: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min -1.0 and max 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "-1.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
			Err: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 0.0 and min 1.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "1.0",
			},
			Err: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 1.0 and max 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "1.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "0.0",
			},
			Err: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 5.0 and min 10.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "5.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "10.0",
			},
			Err: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 10.0 and max 5.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "10.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "5.0",
			},
			Err: ebay.ErrInvalidMaxPrice,
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with unparsable max and min 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
			Err: fmt.Errorf("%w: %s: %w",
				ebay.ErrInvalidPrice, `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 0.0 and unparsable max",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w",
				ebay.ErrInvalidPrice, `strconv.ParseFloat: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max -1.0 and min 0.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxPrice",
				"itemFilter(0).value": "-1.0",
				"itemFilter(1).name":  "MinPrice",
				"itemFilter(1).value": "0.0",
			},
			Err: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 0.0 and max -1.0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinPrice",
				"itemFilter(0).value": "0.0",
				"itemFilter(1).name":  "MaxPrice",
				"itemFilter(1).value": "-1.0",
			},
			Err: fmt.Errorf("%w: %f (minimum value: %f)", ebay.ErrInvalidPrice, -1.0, 0.0),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 10.0 and min 5.0, paramName Invalid",
			Params: map[string]string{
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Invalid",
				"itemFilter(1).paramValue": "EUR",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 5.0, paramName Invalid and max 10.0",
			Params: map[string]string{
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Invalid",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 10.0 and min 5.0, paramValue ZZZ",
			Params: map[string]string{
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "ZZZ",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with min 5.0, paramValue ZZZ and max 10.0",
			Params: map[string]string{
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "ZZZ",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with max 10.0, paramName Invalid and min 5.0",
			Params: map[string]string{
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(0).paramName":  "Invalid",
				"itemFilter(0).paramValue": "EUR",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice/MaxPrice itemFilters with min 5.0 and max 10.0, paramName Invalid",
			Params: map[string]string{
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
				"itemFilter(1).paramName":  "Invalid",
				"itemFilter(1).paramValue": "EUR",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidPriceParamName, "Invalid"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with max 10.0, paramValue ZZZ and min 5.0",
			Params: map[string]string{
				"itemFilter(0).name":       "MaxPrice",
				"itemFilter(0).value":      "10.0",
				"itemFilter(0).paramName":  "Currency",
				"itemFilter(0).paramValue": "ZZZ",
				"itemFilter(1).name":       "MinPrice",
				"itemFilter(1).value":      "5.0",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "returns error if params contains MinPrice itemFilter with min 5.0 and max 10.0, paramValue ZZZ",
			Params: map[string]string{
				"itemFilter(0).name":       "MinPrice",
				"itemFilter(0).value":      "5.0",
				"itemFilter(1).name":       "MaxPrice",
				"itemFilter(1).value":      "10.0",
				"itemFilter(1).paramName":  "Currency",
				"itemFilter(1).paramValue": "ZZZ",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidCurrencyID, "ZZZ"),
		},
		{
			Name: "can find items if params contains MaxQuantity itemFilter with max 1",
			Params: map[string]string{
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "1",
			},
		},
		{
			Name: "can find items if params contains MaxQuantity itemFilter with max 5",
			Params: map[string]string{
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MaxQuantity itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MaxQuantity itemFilter with max 0",
			Params: map[string]string{
				"itemFilter.name":  "MaxQuantity",
				"itemFilter.value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "can find items if params contains MinQuantity itemFilter with max 1",
			Params: map[string]string{
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity itemFilter with max 5",
			Params: map[string]string{
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "5",
			},
		},
		{
			Name: "returns error if params contains MinQuantity itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "not a minimum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity itemFilter with max 0",
			Params: map[string]string{
				"itemFilter.name":  "MinQuantity",
				"itemFilter.value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with max 2 and min 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "2",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with min 1 and max 2",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "2",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with max and min 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with min and max 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with max 10 and min 5",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "5",
			},
		},
		{
			Name: "can find items if params contains MinQuantity/MaxQuantity itemFilters with min 5 and max 10",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "10",
			},
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 1 and unparsable min",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "not a minimum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with unparsable min and max 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "not a minimum",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a minimum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 1 and min 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 0 and max 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 1 and min 2",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "2",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 2 and max 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "2",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "1",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 5 and min 10",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "5",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "10",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 10 and max 5",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "10",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "5",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with unparsable max and min 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "not a maximum",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 1 and unparsable max",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "not a maximum",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidInteger, `strconv.Atoi: parsing "not a maximum"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with max 0 and min 1",
			Params: map[string]string{
				"itemFilter(0).name":  "MaxQuantity",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "MinQuantity",
				"itemFilter(1).value": "1",
			},
			Err: fmt.Errorf("%w: %q (minimum value: %d)", ebay.ErrInvalidInteger, "0", 1),
		},
		{
			Name: "returns error if params contains MinQuantity/MaxQuantity itemFilters with min 1 and max 0",
			Params: map[string]string{
				"itemFilter(0).name":  "MinQuantity",
				"itemFilter(0).value": "1",
				"itemFilter(1).name":  "MaxQuantity",
				"itemFilter(1).value": "0",
			},
			Err: fmt.Errorf("%w: %q must be greater than or equal to %q",
				ebay.ErrInvalidNumericFilter, "MaxQuantity", "MinQuantity"),
		},
		{
			Name: "can find items if params contains ModTimeFrom itemFilter with past timestamp",
			Params: map[string]string{
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains ModTimeFrom itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": "not a timestamp",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains ModTimeFrom itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).In(easternTime).Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).In(easternTime).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains ModTimeFrom itemFilter with future timestamp",
			Params: map[string]string{
				"itemFilter.name":  "ModTimeFrom",
				"itemFilter.value": time.Now().Add(5 * time.Second).UTC().Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(5*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains ReturnsAcceptedOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "ReturnsAcceptedOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains ReturnsAcceptedOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "ReturnsAcceptedOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains ReturnsAcceptedOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "ReturnsAcceptedOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains Seller itemFilter with seller ID 0",
			Params: map[string]string{
				"itemFilter.name":  "Seller",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "can find items if params contains Seller itemFilter with seller IDs 0 and 1",
			Params: map[string]string{
				"itemFilter.name":     "Seller",
				"itemFilter.value(0)": "0",
				"itemFilter.value(1)": "1",
			},
		},
		{
			Name: "returns error if params contains Seller and ExcludeSeller itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":  "Seller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "ExcludeSeller",
				"itemFilter(1).value": "0",
			},
			Err: ebay.ErrSellerCannotBeUsedWithOtherSellers,
		},
		{
			Name: "returns error if params contains Seller and TopRatedSellerOnly itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":  "Seller",
				"itemFilter(0).value": "0",
				"itemFilter(1).name":  "TopRatedSellerOnly",
				"itemFilter(1).value": "true",
			},
			Err: ebay.ErrSellerCannotBeUsedWithOtherSellers,
		},
		{
			Name:   "can find items if params contains Seller itemFilter with 100 seller IDs",
			Params: generateFilterParams("Seller", 100),
		},
		{
			Name:   "returns error if params contains Seller itemFilter with 101 seller IDs",
			Params: generateFilterParams("Seller", 101),
			Err:    ebay.ErrMaxSellers,
		},
		{
			Name: "can find items if params contains SellerBusinessType itemFilter with Business type",
			Params: map[string]string{
				"itemFilter.name":  "SellerBusinessType",
				"itemFilter.value": "Business",
			},
		},
		{
			Name: "can find items if params contains SellerBusinessType itemFilter with Private type",
			Params: map[string]string{
				"itemFilter.name":  "SellerBusinessType",
				"itemFilter.value": "Private",
			},
		},
		{
			Name: "returns error if params contains SellerBusinessType itemFilter with NotBusiness type",
			Params: map[string]string{
				"itemFilter.name":  "SellerBusinessType",
				"itemFilter.value": "NotBusiness",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidSellerBusinessType, "NotBusiness"),
		},
		{
			Name: "returns error if params contains SellerBusinessType itemFilter with Business and Private types",
			Params: map[string]string{
				"itemFilter.name":     "SellerBusinessType",
				"itemFilter.value(0)": "Business",
				"itemFilter.value(1)": "Private",
			},
			Err: ebay.ErrMultipleSellerBusinessTypes,
		},
		{
			Name: "can find items if params contains SoldItemsOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "SoldItemsOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains SoldItemsOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "SoldItemsOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains SoldItemsOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "SoldItemsOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contains StartTimeFrom itemFilter with future timestamp",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": time.Now().Add(5 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains StartTimeFrom itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": "not a timestamp",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains StartTimeFrom itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": time.Now().Add(1 * time.Second).In(easternTime).Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).In(easternTime).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains StartTimeFrom itemFilter with past timestamp",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeFrom",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains StartTimeTo itemFilter with future timestamp",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": time.Now().Add(5 * time.Second).UTC().Format(time.RFC3339),
			},
		},
		{
			Name: "returns error if params contains StartTimeTo itemFilter with unparsable value",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": "not a timestamp",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidDateTime, "not a timestamp"),
		},
		{
			Name: "returns error if params contains StartTimeTo itemFilter with non-UTC timestamp",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": time.Now().Add(1 * time.Second).In(easternTime).Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(1*time.Second).In(easternTime).Format(time.RFC3339)),
		},
		{
			Name: "returns error if params contains StartTimeTo itemFilter with past timestamp",
			Params: map[string]string{
				"itemFilter.name":  "StartTimeTo",
				"itemFilter.value": time.Now().Add(-1 * time.Second).UTC().Format(time.RFC3339),
			},
			Err: fmt.Errorf("%w: %q",
				ebay.ErrInvalidDateTime, time.Now().Add(-1*time.Second).UTC().Format(time.RFC3339)),
		},
		{
			Name: "can find items if params contains TopRatedSellerOnly itemFilter.value=true",
			Params: map[string]string{
				"itemFilter.name":  "TopRatedSellerOnly",
				"itemFilter.value": "true",
			},
		},
		{
			Name: "can find items if params contains TopRatedSellerOnly itemFilter.value=false",
			Params: map[string]string{
				"itemFilter.name":  "TopRatedSellerOnly",
				"itemFilter.value": "false",
			},
		},
		{
			Name: "returns error if params contains TopRatedSellerOnly itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "TopRatedSellerOnly",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "returns error if params contains TopRatedSellerOnly and Seller itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":  "TopRatedSellerOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "Seller",
				"itemFilter(1).value": "0",
			},
			Err: ebay.ErrTopRatedSellerCannotBeUsedWithSellers,
		},
		{
			Name: "returns error if params contains TopRatedSellerOnly and ExcludeSeller itemFilters",
			Params: map[string]string{
				"itemFilter(0).name":  "TopRatedSellerOnly",
				"itemFilter(0).value": "true",
				"itemFilter(1).name":  "ExcludeSeller",
				"itemFilter(1).value": "0",
			},
			Err: ebay.ErrTopRatedSellerCannotBeUsedWithSellers,
		},
		{
			Name: "can find items if params contains ValueBoxInventory itemFilter.value=1",
			Params: map[string]string{
				"itemFilter.name":  "ValueBoxInventory",
				"itemFilter.value": "1",
			},
		},
		{
			Name: "can find items if params contains ValueBoxInventory itemFilter.value=0",
			Params: map[string]string{
				"itemFilter.name":  "ValueBoxInventory",
				"itemFilter.value": "0",
			},
		},
		{
			Name: "returns error if params contains ValueBoxInventory itemFilter with non-boolean value",
			Params: map[string]string{
				"itemFilter.name":  "ValueBoxInventory",
				"itemFilter.value": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidValueBoxInventory, "123"),
		},
		{
			Name:   "can find items if params contains AspectHistogram outputSelector",
			Params: map[string]string{"outputSelector": "AspectHistogram"},
		},
		{
			Name:   "can find items if params contains CategoryHistogram outputSelector",
			Params: map[string]string{"outputSelector": "CategoryHistogram"},
		},
		{
			Name:   "can find items if params contains ConditionHistogram outputSelector",
			Params: map[string]string{"outputSelector": "ConditionHistogram"},
		},
		{
			Name:   "can find items if params contains GalleryInfo outputSelector",
			Params: map[string]string{"outputSelector": "GalleryInfo"},
		},
		{
			Name:   "can find items if params contains PictureURLLarge outputSelector",
			Params: map[string]string{"outputSelector": "PictureURLLarge"},
		},
		{
			Name:   "can find items if params contains PictureURLSuperSize outputSelector",
			Params: map[string]string{"outputSelector": "PictureURLSuperSize"},
		},
		{
			Name:   "can find items if params contains SellerInfo outputSelector",
			Params: map[string]string{"outputSelector": "SellerInfo"},
		},
		{
			Name:   "can find items if params contains StoreInfo outputSelector",
			Params: map[string]string{"outputSelector": "StoreInfo"},
		},
		{
			Name:   "can find items if params contains UnitPriceInfo outputSelector",
			Params: map[string]string{"outputSelector": "UnitPriceInfo"},
		},
		{
			Name:   "returns error if params contains non-numbered, unsupported outputSelector name",
			Params: map[string]string{"outputSelector": "UnsupportedOutputSelector"},
			Err:    ebay.ErrInvalidOutputSelector,
		},
		{
			// outputSelector(1) will be ignored because indexing does not start at 0.
			Name: "can find items if params contains outputSelector, outputSelector(1)",
			Params: map[string]string{
				"outputSelector":    "AspectHistogram",
				"outputSelector(1)": "CategoryHistogram",
			},
		},
		{
			Name: "returns error if params contain numbered and non-numbered outputSelector syntax types",
			Params: map[string]string{
				"outputSelector":    "AspectHistogram",
				"outputSelector(0)": "CategoryHistogram",
			},
			Err: ebay.ErrInvalidIndexSyntax,
		},
		{
			Name:   "can find items by numbered outputSelector",
			Params: map[string]string{"outputSelector(0)": "AspectHistogram"},
		},
		{
			Name: "can find items by 2 numbered outputSelector",
			Params: map[string]string{
				"outputSelector(0)": "AspectHistogram",
				"outputSelector(1)": "CategoryHistogram",
			},
		},
		{
			Name:   "returns error if params contains numbered, unsupported outputSelector name",
			Params: map[string]string{"outputSelector(0)": "UnsupportedOutputSelector"},
			Err:    ebay.ErrInvalidOutputSelector,
		},
		{
			Name: "returns error if params contains 1 supported, 1 unsupported outputSelector name",
			Params: map[string]string{
				"outputSelector(0)": "AspectHistogram",
				"outputSelector(1)": "UnsupportedOutputSelector",
			},
			Err: ebay.ErrInvalidOutputSelector,
		},
		{
			Name:   "can find items if params contains affiliate.customId=1",
			Params: map[string]string{"affiliate.customId": "1"},
		},
		{
			Name:   "can find items if params contains affiliate.customId of length 256",
			Params: map[string]string{"affiliate.customId": generateStringWithLen(256, false)},
		},
		{
			Name:   "returns error if params contains affiliate.customId of length 257",
			Params: map[string]string{"affiliate.customId": generateStringWithLen(257, false)},
			Err:    ebay.ErrInvalidCustomIDLength,
		},
		{
			Name:   "can find items if params contains affiliate.geoTargeting=true",
			Params: map[string]string{"affiliate.geoTargeting": "true"},
		},
		{
			Name:   "can find items if params contains affiliate.geoTargeting=false",
			Params: map[string]string{"affiliate.geoTargeting": "false"},
		},
		{
			Name:   "returns error if params contains affiliate.geoTargeting with non-boolean value",
			Params: map[string]string{"affiliate.geoTargeting": "123"},
			Err:    fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name: "can find items if params contain affiliate.networkId=2 and trackingId",
			Params: map[string]string{
				"affiliate.networkId":  "2",
				"affiliate.trackingId": "1",
			},
		},
		{
			Name: "can find items if params contain affiliate.networkId=9 and trackingId=1234567890",
			Params: map[string]string{
				"affiliate.networkId":  "9",
				"affiliate.trackingId": "1234567890",
			},
		},
		{
			Name: "can find items if params contain affiliate.networkId=5 and trackingId=veryunique",
			Params: map[string]string{
				"affiliate.networkId":  "5",
				"affiliate.trackingId": "veryunique",
			},
		},
		{
			Name:   "returns error if params contains affiliate.networkId but no trackingId",
			Params: map[string]string{"affiliate.networkId": "2"},
			Err:    ebay.ErrIncompleteAffiliateParams,
		},
		{
			Name:   "returns error if params contains affiliate.trackingId but no networkId",
			Params: map[string]string{"affiliate.trackingId": "1"},
			Err:    ebay.ErrIncompleteAffiliateParams,
		},
		{
			Name: "returns error if params contain affiliate.networkId=abc and trackingId",
			Params: map[string]string{
				"affiliate.networkId":  "abc",
				"affiliate.trackingId": "1",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidNetworkID, `strconv.Atoi: parsing "abc"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contain affiliate.networkId=1 and trackingId",
			Params: map[string]string{
				"affiliate.networkId":  "1",
				"affiliate.trackingId": "1",
			},
			Err: ebay.ErrInvalidNetworkIDRange,
		},
		{
			Name: "returns error if params contain affiliate.networkId=10 and trackingId",
			Params: map[string]string{
				"affiliate.networkId":  "10",
				"affiliate.trackingId": "1",
			},
			Err: ebay.ErrInvalidNetworkIDRange,
		},
		{
			Name: "returns error if params contain affiliate.networkId=9 and trackingId=abc",
			Params: map[string]string{
				"affiliate.networkId":  "9",
				"affiliate.trackingId": "abc",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidTrackingID, `strconv.Atoi: parsing "abc"`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contain affiliate.networkId=9 and trackingId=123456789",
			Params: map[string]string{
				"affiliate.networkId":  "9",
				"affiliate.trackingId": "123456789",
			},
			Err: ebay.ErrInvalidCampaignID,
		},
		{
			Name: "can find items if params contain affiliate.customId, geoTargeting, networkId, trackingId",
			Params: map[string]string{
				"affiliate.customId":     "abc123",
				"affiliate.geoTargeting": "true",
				"affiliate.networkId":    "2",
				"affiliate.trackingId":   "123abc",
			},
		},
		{
			Name:   "can find items if params contains buyerPostalCode=111",
			Params: map[string]string{"buyerPostalCode": "111"},
		},
		{
			Name:   "can find items if params contains buyerPostalCode=aaaaa",
			Params: map[string]string{"buyerPostalCode": "aaaaa"},
		},
		{
			Name:   "can find items if params contains buyerPostalCode=Postal Code Here",
			Params: map[string]string{"buyerPostalCode": "Postal Code Here"},
		},
		{
			Name:   "returns error if params contains buyerPostalCode=11",
			Params: map[string]string{"buyerPostalCode": "11"},
			Err:    ebay.ErrInvalidPostalCode,
		},
		{
			Name:   "can find items if params contains paginationInput.entriesPerPage=1",
			Params: map[string]string{"paginationInput.entriesPerPage": "1"},
		},
		{
			Name:   "can find items if params contains paginationInput.entriesPerPage=50",
			Params: map[string]string{"paginationInput.entriesPerPage": "50"},
		},
		{
			Name:   "can find items if params contains paginationInput.entriesPerPage=100",
			Params: map[string]string{"paginationInput.entriesPerPage": "100"},
		},
		{
			Name:   "returns error if params contains paginationInput.entriesPerPage=a",
			Params: map[string]string{"paginationInput.entriesPerPage": "a"},
			Err:    fmt.Errorf("%w: %s: %w", ebay.ErrInvalidEntriesPerPage, `strconv.Atoi: parsing "a"`, strconv.ErrSyntax),
		},
		{
			Name:   "returns error if params contains paginationInput.entriesPerPage=0",
			Params: map[string]string{"paginationInput.entriesPerPage": "0"},
			Err:    ebay.ErrInvalidEntriesPerPageRange,
		},
		{
			Name:   "returns error if params contains paginationInput.entriesPerPage=101",
			Params: map[string]string{"paginationInput.entriesPerPage": "101"},
			Err:    ebay.ErrInvalidEntriesPerPageRange,
		},
		{
			Name:   "can find items if params contains paginationInput.pageNumber=1",
			Params: map[string]string{"paginationInput.pageNumber": "1"},
		},
		{
			Name:   "can find items if params contains paginationInput.pageNumber=50",
			Params: map[string]string{"paginationInput.pageNumber": "50"},
		},
		{
			Name:   "can find items if params contains paginationInput.pageNumber=100",
			Params: map[string]string{"paginationInput.pageNumber": "100"},
		},
		{
			Name:   "returns error if params contains paginationInput.pageNumber=a",
			Params: map[string]string{"paginationInput.pageNumber": "a"},
			Err:    fmt.Errorf("%w: %s: %w", ebay.ErrInvalidPageNumber, `strconv.Atoi: parsing "a"`, strconv.ErrSyntax),
		},
		{
			Name:   "returns error if params contains paginationInput.pageNumber=0",
			Params: map[string]string{"paginationInput.pageNumber": "0"},
			Err:    ebay.ErrInvalidPageNumberRange,
		},
		{
			Name:   "returns error if params contains paginationInput.pageNumber=101",
			Params: map[string]string{"paginationInput.pageNumber": "101"},
			Err:    ebay.ErrInvalidPageNumberRange,
		},
		{
			Name: "can find items if params contains paginationInput.entriesPerPage=1, paginationInput.pageNumber=1",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "1",
				"paginationInput.pageNumber":     "1",
			},
		},
		{
			Name: "can find items if params contains paginationInput.entriesPerPage=100, paginationInput.pageNumber=100",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "100",
				"paginationInput.pageNumber":     "100",
			},
		},
		{
			Name: "returns if params contains paginationInput.entriesPerPage=0, paginationInput.pageNumber=1",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "0",
				"paginationInput.pageNumber":     "1",
			},
			Err: ebay.ErrInvalidEntriesPerPageRange,
		},
		{
			Name: "returns if params contains paginationInput.entriesPerPage=101, paginationInput.pageNumber=1",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "101",
				"paginationInput.pageNumber":     "1",
			},
			Err: ebay.ErrInvalidEntriesPerPageRange,
		},
		{
			Name: "returns if params contains paginationInput.entriesPerPage=1, paginationInput.pageNumber=0",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "1",
				"paginationInput.pageNumber":     "0",
			},
			Err: ebay.ErrInvalidPageNumberRange,
		},
		{
			Name: "returns if params contains paginationInput.entriesPerPage=1, paginationInput.pageNumber=101",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "1",
				"paginationInput.pageNumber":     "101",
			},
			Err: ebay.ErrInvalidPageNumberRange,
		},
		{
			Name: "returns if params contains paginationInput.entriesPerPage=0, paginationInput.pageNumber=0",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "0",
				"paginationInput.pageNumber":     "0",
			},
			Err: ebay.ErrInvalidEntriesPerPageRange,
		},
		{
			Name: "returns if params contains paginationInput.entriesPerPage=101, paginationInput.pageNumber=101",
			Params: map[string]string{
				"paginationInput.entriesPerPage": "101",
				"paginationInput.pageNumber":     "101",
			},
			Err: ebay.ErrInvalidEntriesPerPageRange,
		},
		{
			Name: "can find items if params contains BestMatch sortOrder",
			Params: map[string]string{
				"sortOrder": "BestMatch",
			},
		},
		{
			Name: "can find items if params contains BidCountFewest sortOrder and Auction listing type",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "Auction",
				"sortOrder":        "BidCountFewest",
			},
		},
		{
			Name:   "returns error if params contains BidCountFewest sortOrder but no Auction listing type",
			Params: map[string]string{"sortOrder": "BidCountFewest"},
			Err:    ebay.ErrAuctionListingMissing,
		},
		{
			Name: "can find items if params contains BidCountMost sortOrder and Auction listing type",
			Params: map[string]string{
				"itemFilter.name":  "ListingType",
				"itemFilter.value": "Auction",
				"sortOrder":        "BidCountMost",
			},
		},
		{
			Name:   "returns error if params contains BidCountMost sortOrder but no Auction listing type",
			Params: map[string]string{"sortOrder": "BidCountMost"},
			Err:    ebay.ErrAuctionListingMissing,
		},
		{
			Name:   "can find items if params contains CountryAscending sortOrder",
			Params: map[string]string{"sortOrder": "CountryAscending"},
		},
		{
			Name:   "can find items if params contains CountryDescending sortOrder",
			Params: map[string]string{"sortOrder": "CountryDescending"},
		},
		{
			Name:   "can find items if params contains CurrentPriceHighest sortOrder",
			Params: map[string]string{"sortOrder": "CurrentPriceHighest"},
		},
		{
			Name: "can find items if params contains DistanceNearest sortOrder and buyerPostalCode",
			Params: map[string]string{
				"buyerPostalCode": "111",
				"sortOrder":       "DistanceNearest",
			},
		},
		{
			Name:   "returns error if params contains DistanceNearest sortOrder but no buyerPostalCode",
			Params: map[string]string{"sortOrder": "DistanceNearest"},
			Err:    ebay.ErrBuyerPostalCodeMissing,
		},
		{
			Name:   "can find items if params contains EndTimeSoonest sortOrder",
			Params: map[string]string{"sortOrder": "EndTimeSoonest"},
		},
		{
			Name:   "can find items if params contains PricePlusShippingHighest sortOrder",
			Params: map[string]string{"sortOrder": "PricePlusShippingHighest"},
		},
		{
			Name:   "can find items if params contains PricePlusShippingLowest sortOrder",
			Params: map[string]string{"sortOrder": "PricePlusShippingLowest"},
		},
		{
			Name:   "can find items if params contains StartTimeNewest sortOrder",
			Params: map[string]string{"sortOrder": "StartTimeNewest"},
		},
		{
			Name:   "can find items if params contains WatchCountDecreaseSort sortOrder",
			Params: map[string]string{"sortOrder": "WatchCountDecreaseSort"},
		},
		{
			Name:   "returns error if params contains unsupported sortOrder name",
			Params: map[string]string{"sortOrder": "UnsupportedSortOrder"},
			Err:    ebay.ErrUnsupportedSortOrderType,
		},
	}
	aspectFilterTestCases = []findItemsTestCase{
		{
			Name: "can find items by aspectFilter.aspectName, aspectValueName",
			Params: map[string]string{
				"aspectFilter.aspectName":      "Size",
				"aspectFilter.aspectValueName": "10",
			},
		},
		{
			Name: "can find items by aspectFilter.aspectName, aspectValueName(0), aspectValueName(1)",
			Params: map[string]string{
				"aspectFilter.aspectName":         "Size",
				"aspectFilter.aspectValueName(0)": "10",
				"aspectFilter.aspectValueName(1)": "11",
			},
		},
		{
			Name:   "returns error if params contains aspectFilter.aspectName but not aspectValueName",
			Params: map[string]string{"aspectFilter.aspectName": "Size"},
			Err:    fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "aspectFilter.aspectValueName"),
		},
		{
			// aspectFilter.aspectValueName(1) will be ignored because indexing does not start at 0.
			Name: "returns error if params contains aspectFilter.aspectName, aspectValueName(1)",
			Params: map[string]string{
				"aspectFilter.aspectName":         "Size",
				"aspectFilter.aspectValueName(1)": "10",
			},
			Err: fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "aspectFilter.aspectValueName"),
		},
		{
			// aspectFilter.aspectValueName(1) will be ignored because indexing does not start at 0.
			// Therefore, only aspectFilter.aspectValueName is considered and this becomes a non-numbered aspectFilter.
			Name: "can find items by aspectFilter.aspectName, aspectValueName, aspectValueName(1)",
			Params: map[string]string{
				"aspectFilter.aspectName":         "Size",
				"aspectFilter.aspectValueName":    "10",
				"aspectFilter.aspectValueName(1)": "11",
			},
		},
		{
			// The aspectFilter will be ignored if no aspectFilter.aspectName param is found before other aspectFilter params.
			Name:   "can find items if params contains aspectFilter.aspectValueName only",
			Params: map[string]string{"aspectFilter.aspectValueName": "10"},
		},
		{
			Name: "returns error if params contain numbered and non-numbered aspectFilter syntax types",
			Params: map[string]string{
				"aspectFilter.aspectName":         "Size",
				"aspectFilter.aspectValueName":    "10",
				"aspectFilter(0).aspectName":      "Running",
				"aspectFilter(0).aspectValueName": "true",
			},
			Err: ebay.ErrInvalidIndexSyntax,
		},
		{
			Name: "returns error if params contain aspectFilter.aspectName, aspectValueName, aspectValueName(0)",
			Params: map[string]string{
				"aspectFilter.aspectName":         "Size",
				"aspectFilter.aspectValueName":    "10",
				"aspectFilter.aspectValueName(0)": "11",
			},
			Err: ebay.ErrInvalidIndexSyntax,
		},
		{
			Name: "returns error if params contain aspectFilter(0).aspectName, aspectValueName, aspectValueName(0)",
			Params: map[string]string{
				"aspectFilter(0).aspectName":         "Size",
				"aspectFilter(0).aspectValueName":    "10",
				"aspectFilter(0).aspectValueName(0)": "11",
			},
			Err: ebay.ErrInvalidIndexSyntax,
		},
		{
			Name: "can find items by aspectFilter(0).aspectName, aspectValueName",
			Params: map[string]string{
				"aspectFilter(0).aspectName":      "Size",
				"aspectFilter(0).aspectValueName": "10",
			},
		},
		{
			Name: "can find items by aspectFilter(0).aspectName, aspectValueName(0), aspectValueName(1)",
			Params: map[string]string{
				"aspectFilter(0).aspectName":         "Size",
				"aspectFilter(0).aspectValueName(0)": "10",
				"aspectFilter(0).aspectValueName(1)": "11",
			},
		},
		{
			Name: "can find items by 2 numbered aspectFilters",
			Params: map[string]string{
				"aspectFilter(0).aspectName":      "Size",
				"aspectFilter(0).aspectValueName": "10",
				"aspectFilter(1).aspectName":      "Running",
				"aspectFilter(1).aspectValueName": "true",
			},
		},
		{
			Name:   "returns error if params contains aspectFilter(0).aspectName but not aspectValueName",
			Params: map[string]string{"aspectFilter(0).aspectName": "Size"},
			Err:    fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "aspectFilter(0).aspectValueName"),
		},
		{
			// aspectFilter(0).aspectValueName(1) will be ignored because indexing does not start at 0.
			Name: "returns error if params contains aspectFilter(0).aspectName, aspectValueName(1)",
			Params: map[string]string{
				"aspectFilter(0).aspectName":         "Size",
				"aspectFilter(0).aspectValueName(1)": "10",
			},
			Err: fmt.Errorf("%w %q", ebay.ErrIncompleteFilterNameOnly, "aspectFilter(0).aspectValueName"),
		},
		{
			// aspectFilter(0).aspectValueName(1) will be ignored because indexing does not start at 0.
			// Therefore, only aspectFilter(0).aspectValueName is considered and this becomes a numbered aspectFilter.
			Name: "can find items by aspectFilter(0).aspectName, aspectValueName aspectValueName(1)",
			Params: map[string]string{
				"aspectFilter(0).aspectName":         "Size",
				"aspectFilter(0).aspectValueName":    "10",
				"aspectFilter(0).aspectValueName(1)": "11",
			},
		},
		{
			// The aspectFilter will be ignored if no aspectFilter(0).aspectName param is found before other aspectFilter params.
			Name:   "can find items if params contains aspectFilter(0).aspectValueName only",
			Params: map[string]string{"aspectFilter(0).aspectValueName": "10"},
		},
	}
)

func TestFindItemsByCategories(t *testing.T) {
	t.Parallel()
	params := map[string]string{"categoryId": "12345"}
	findItemsByCategoriesTCs := combineTestCases(t, findItemsByCategories, categoryIDTCs)
	testFindItems(t, params, findItemsByCategories, findItemsByCategoriesResp, findItemsByCategoriesTCs)
}

func TestFindItemsByKeywords(t *testing.T) {
	t.Parallel()
	params := map[string]string{"keywords": "marshmallows"}
	findItemsByKeywordsTCs := combineTestCases(t, findItemsByKeywords, keywordsTCs)
	testFindItems(t, params, findItemsByKeywords, findItemsByKeywordsResp, findItemsByKeywordsTCs)
}

func TestFindItemsAdvanced(t *testing.T) {
	t.Parallel()
	params := map[string]string{"categoryId": "12345"}
	findItemsAdvancedTCs := []findItemsTestCase{
		{
			Name: "can find items if params contains descriptionSearch=true",
			Params: map[string]string{
				"categoryId":        "1",
				"descriptionSearch": "true",
			},
		},
		{
			Name: "can find items if params contains descriptionSearch=false",
			Params: map[string]string{
				"categoryId":        "1",
				"descriptionSearch": "false",
			},
		},
		{
			Name: "returns error if params contains descriptionSearch with non-boolean value",
			Params: map[string]string{
				"categoryId":        "1",
				"descriptionSearch": "123",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrInvalidBooleanValue, "123"),
		},
		{
			Name:   "returns error if params contains descriptionSearch but not categoryId or keywords",
			Params: map[string]string{"descriptionSearch": "true"},
			Err:    ebay.ErrCategoryIDKeywordsMissing,
		},
	}
	combinedTCs := combineTestCases(
		t, findItemsAdvanced, categoryIDTCs, keywordsTCs, categoryIDKeywordsTCs, findItemsAdvancedTCs)
	testFindItems(t, params, findItemsAdvanced, findItemsAdvancedResp, combinedTCs)
}

func TestFindItemsByProduct(t *testing.T) {
	t.Parallel()
	params := map[string]string{
		"productId.@type": "ReferenceID",
		"productId":       "123",
	}
	findItemsByProductTCs := []findItemsTestCase{
		{
			Name:   "returns error if params contains productId but not productId.@type",
			Params: map[string]string{"productId": "123"},
			Err:    ebay.ErrProductIDMissing,
		},
		{
			Name:   "returns error if params contains productId.@type but not productId",
			Params: map[string]string{"productId.@type": "ReferenceID"},
			Err:    ebay.ErrProductIDMissing,
		},
		{
			Name: "returns error if params contains productId.@type=UnsupportedProductID, productId=1",
			Params: map[string]string{
				"productId.@type": "UnsupportedProductID",
				"productId":       "1",
			},
			Err: fmt.Errorf("%w: %q", ebay.ErrUnsupportedProductIDType, "UnsupportedProductID"),
		},
		{
			Name: "can find items if params contains productId.@type=ReferenceID, productId=1",
			Params: map[string]string{
				"productId.@type": "ReferenceID",
				"productId":       "1",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ReferenceID, productId=123",
			Params: map[string]string{
				"productId.@type": "ReferenceID",
				"productId":       "123",
			},
		},
		{
			Name:   "returns error if params contains productId.@type=ReferenceID, empty productId",
			Params: map[string]string{"productId.@type": "ReferenceID", "productId": ""},
			Err:    ebay.ErrInvalidProductIDLength,
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=0131103628",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "0131103628",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=954911659X",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "954911659X",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=802510897X",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "802510897X",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=7111075897",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "7111075897",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=986154142X",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "986154142X",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=9780131101630",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "9780131101630",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=9780131103627",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "9780131103627",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=9780133086249",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "9780133086249",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=9789332549449",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "9789332549449",
			},
		},
		{
			Name: "can find items if params contains productId.@type=ISBN, productId=9780131158177",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "9780131158177",
			},
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of length 9",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "111111111",
			},
			Err: ebay.ErrInvalidISBNLength,
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of length 11",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "11111111111",
			},
			Err: ebay.ErrInvalidISBNLength,
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of length 12",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "111111111111",
			},
			Err: ebay.ErrInvalidISBNLength,
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of length 14",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "11111111111111",
			},
			Err: ebay.ErrInvalidISBNLength,
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of invalid ISBN-10 (invalid first digit)",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "886154142X",
			},
			Err: ebay.ErrInvalidISBN,
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of invalid ISBN-13 (invalid first digit)",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "8780131158177",
			},
			Err: ebay.ErrInvalidISBN,
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of invalid ISBN-10 (invalid last digit)",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "9861541429",
			},
			Err: ebay.ErrInvalidISBN,
		},
		{
			Name: "returns error if params contains productId.@type=ISBN, productId of invalid ISBN-13 (invalid last digit)",
			Params: map[string]string{
				"productId.@type": "ISBN",
				"productId":       "9780131158178",
			},
			Err: ebay.ErrInvalidISBN,
		},
		{
			Name: "can find items if params contains productId.@type=UPC, productId=036000291452",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "036000291452",
			},
		},
		{
			Name: "can find items if params contains productId.@type=UPC, productId=194253378907",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "194253378907",
			},
		},
		{
			Name: "can find items if params contains productId.@type=UPC, productId=753575979881",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "753575979881",
			},
		},
		{
			Name: "can find items if params contains productId.@type=UPC, productId=194253402220",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "194253402220",
			},
		},
		{
			Name: "can find items if params contains productId.@type=UPC, productId=194253407980",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "194253407980",
			},
		},
		{
			Name: "returns error if params contains productId.@type=UPC, productId of length 11",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "11111111111",
			},
			Err: ebay.ErrInvalidUPCLength,
		},
		{
			Name: "returns error if params contains productId.@type=UPC, productId of length 13",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "1111111111111",
			},
			Err: ebay.ErrInvalidUPCLength,
		},
		{
			Name: "returns error if params contains productId.@type=UPC, productId of invalid UPC (invalid first digit)",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "294253407980",
			},
			Err: ebay.ErrInvalidUPC,
		},
		{
			Name: "returns error if params contains productId.@type=UPC, productId of invalid UPC (invalid last digit)",
			Params: map[string]string{
				"productId.@type": "UPC",
				"productId":       "194253407981",
			},
			Err: ebay.ErrInvalidUPC,
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=73513537",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "73513537",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=96385074",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "96385074",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=29033706",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "29033706",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=40170725",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "40170725",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=40123455",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "40123455",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=4006381333931",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "4006381333931",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=0194253373933",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "0194253373933",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=0194253374398",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "0194253374398",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=0194253381099",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "0194253381099",
			},
		},
		{
			Name: "can find items if params contains productId.@type=EAN, productId=0194253373476",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "0194253373476",
			},
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of length 7",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "1111111",
			},
			Err: ebay.ErrInvalidEANLength,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of length 9",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "111111111",
			},
			Err: ebay.ErrInvalidEANLength,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of length 10",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "1111111111",
			},
			Err: ebay.ErrInvalidEANLength,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of length 11",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "11111111111",
			},
			Err: ebay.ErrInvalidEANLength,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of length 12",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "111111111111",
			},
			Err: ebay.ErrInvalidEANLength,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of length 14",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "11111111111111",
			},
			Err: ebay.ErrInvalidEANLength,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of invalid EAN-8 (invalid first digit)",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "50123455",
			},
			Err: ebay.ErrInvalidEAN,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of invalid EAN-13 (invalid first digit)",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "1194253373476",
			},
			Err: ebay.ErrInvalidEAN,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of invalid EAN-8 (invalid last digit)",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "40123456",
			},
			Err: ebay.ErrInvalidEAN,
		},
		{
			Name: "returns error if params contains productId.@type=EAN, productId of invalid EAN-13 (invalid last digit)",
			Params: map[string]string{
				"productId.@type": "EAN",
				"productId":       "0194253373477",
			},
			Err: ebay.ErrInvalidEAN,
		},
	}
	combinedTCs := combineTestCases(t, findItemsByProduct, findItemsByProductTCs)
	testFindItems(t, params, findItemsByProduct, findItemsByProductResp, combinedTCs)
}

func TestFindItemsInEBayStores(t *testing.T) {
	t.Parallel()
	params := map[string]string{"storeName": "Supplytronics"}
	findItemsInEBayStoresTCs := []findItemsTestCase{
		{
			Name:   "can find items if params contains storeName=a",
			Params: map[string]string{"storeName": "a"},
		},
		{
			Name:   "returns error if params contains empty storeName",
			Params: map[string]string{"storeName": ""},
			Err:    ebay.ErrInvalidStoreNameLength,
		},
		{
			Name:   "can find items if params contains storeName=Ben &amp; Jerry's",
			Params: map[string]string{"storeName": "Ben &amp; Jerry's"},
		},
		{
			Name:   "returns error if params contains storeName=Ben & Jerry's",
			Params: map[string]string{"storeName": "Ben & Jerry's"},
			Err:    ebay.ErrInvalidStoreNameAmpersand,
		},
		{
			Name: "can find items if params contains 1 categoryId of length 1, keywords of length 2, storeName of length 1",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(2, true),
				"storeName":  "a",
			},
		},
		{
			Name: "returns error if params contains empty categoryId, keywords of length 2, storeName of length 1",
			Params: map[string]string{
				"categoryId": "",
				"keywords":   generateStringWithLen(2, true),
				"storeName":  "a",
			},
			Err: fmt.Errorf("%w: %s: %w", ebay.ErrInvalidCategoryID, `strconv.Atoi: parsing ""`, strconv.ErrSyntax),
		},
		{
			Name: "returns error if params contains 4 categoryIds, keywords of length 2, storeName of length 1",
			Params: map[string]string{
				"categoryId(0)": "1",
				"categoryId(1)": "2",
				"categoryId(2)": "3",
				"categoryId(3)": "4",
				"keywords":      generateStringWithLen(2, true),
				"storeName":     "a",
			},
			Err: ebay.ErrMaxCategoryIDs,
		},
		{
			Name: "returns error if params contains 1 categoryId of length 1, empty keywords, storeName of length 1",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   "",
				"storeName":  "a",
			},
			Err: ebay.ErrInvalidKeywordsLength,
		},
		{
			Name: "returns error if params contains 1 categoryId of length 1, 1 keyword of length 99, storeName of length 1",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(99, false),
				"storeName":  "a",
			},
			Err: ebay.ErrInvalidKeywordLength,
		},
		{
			Name: "can find items if params contains 1 categoryId of length 1, keywords of length 2, &-escaped storeName",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(2, true),
				"storeName":  "Ben &amp; Jerry's",
			},
		},
		{
			Name: "returns error if params contains 1 categoryId of length 1, keywords of length 2, empty storeName",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(2, true),
				"storeName":  "",
			},
			Err: ebay.ErrInvalidStoreNameLength,
		},
		{
			Name: "returns error if params contains 1 categoryId of length 1, keywords of length 2, storeName=Ben & Jerry's",
			Params: map[string]string{
				"categoryId": "1",
				"keywords":   generateStringWithLen(2, true),
				"storeName":  "Ben & Jerry's",
			},
			Err: ebay.ErrInvalidStoreNameAmpersand,
		},
	}
	combinedTCs := combineTestCases(
		t, findItemsInEBayStores, categoryIDTCs, keywordsTCs, categoryIDKeywordsTCs, findItemsInEBayStoresTCs)
	testFindItems(t, params, findItemsInEBayStores, findItemsInEBayStoresResp, combinedTCs)
}

var findMethodConfigs = map[string]struct {
	missingDesc string
	searchErr   error
	params      map[string]string
}{
	findItemsByCategories: {
		missingDesc: "categoryId",
		searchErr:   ebay.ErrCategoryIDMissing,
		params:      map[string]string{"categoryId": "12345"},
	},
	findItemsByKeywords: {
		missingDesc: "keywords",
		searchErr:   ebay.ErrKeywordsMissing,
		params:      map[string]string{"keywords": "marshmallows"},
	},
	findItemsAdvanced: {
		missingDesc: "categoryId or keywords",
		searchErr:   ebay.ErrCategoryIDKeywordsMissing,
		params:      map[string]string{"categoryId": "12345"},
	},
	findItemsByProduct: {
		missingDesc: "productId",
		searchErr:   ebay.ErrProductIDMissing,
		params:      map[string]string{"productId.@type": "ReferenceID", "productId": "123"},
	},
	findItemsInEBayStores: {
		missingDesc: "categoryId, keywords, or storeName",
		searchErr:   ebay.ErrCategoryIDKeywordsStoreNameMissing,
		params:      map[string]string{"storeName": "Supplytronics"},
	},
}

func combineTestCases(t *testing.T, findMethod string, tcs ...[]findItemsTestCase) []findItemsTestCase {
	t.Helper()
	config, ok := findMethodConfigs[findMethod]
	if !ok {
		t.Fatalf("unsupported findMethod: %s", findMethod)
	}
	missingParamTCs := append([]findItemsTestCase{}, missingSearchParamTCs...)
	commonTCs := append([]findItemsTestCase{}, testCases...)
	if findMethod != findItemsByProduct {
		missingParamTCs = append(missingParamTCs, aspectFilterMissingSearchParamTCs...)
		commonTCs = append(commonTCs, aspectFilterTestCases...)
	}
	for i := range missingParamTCs {
		missingParamTCs[i].Name += config.missingDesc
		missingParamTCs[i].Err = config.searchErr
	}
	for i := range commonTCs {
		paramsCopy := make(map[string]string)
		maps.Copy(paramsCopy, config.params)
		maps.Copy(paramsCopy, commonTCs[i].Params)
		commonTCs[i].Params = paramsCopy
	}
	combinedTCs := append([]findItemsTestCase{}, missingParamTCs...)
	combinedTCs = append(combinedTCs, commonTCs...)
	for _, cs := range tcs {
		combinedTCs = append(combinedTCs, cs...)
	}
	return combinedTCs
}

func testFindItems(t *testing.T, params map[string]string, findMethod string, wantResp ebay.ResultProvider, tcs []findItemsTestCase) {
	t.Helper()
	t.Run(fmt.Sprintf("can find items by %s", findMethod), func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := json.Marshal(wantResp)
			if err != nil {
				t.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(body)
			if err != nil {
				t.Fatal(err)
			}
		}))
		defer ts.Close()
		client := ts.Client()
		fc := ebay.NewFindingClient(client, appID)
		fc.URL = ts.URL

		var resp ebay.ResultProvider
		var err error
		switch findMethod {
		case findItemsByCategories:
			resp, err = fc.FindItemsByCategories(context.Background(), params)
		case findItemsByKeywords:
			resp, err = fc.FindItemsByKeywords(context.Background(), params)
		case findItemsAdvanced:
			resp, err = fc.FindItemsAdvanced(context.Background(), params)
		case findItemsByProduct:
			resp, err = fc.FindItemsByProduct(context.Background(), params)
		case findItemsInEBayStores:
			resp, err = fc.FindItemsInEBayStores(context.Background(), params)
		default:
			t.Fatalf("unsupported findMethod: %s", findMethod)
		}
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(resp, wantResp) {
			t.Errorf("got %v, want %v", resp, wantResp)
		}
	})

	t.Run("returns error if the client returns an error", func(t *testing.T) {
		t.Parallel()
		fc := ebay.NewFindingClient(&http.Client{}, appID)
		fc.URL = "http://localhost"

		var err error
		switch findMethod {
		case findItemsByCategories:
			_, err = fc.FindItemsByCategories(context.Background(), params)
		case findItemsByKeywords:
			_, err = fc.FindItemsByKeywords(context.Background(), params)
		case findItemsAdvanced:
			_, err = fc.FindItemsAdvanced(context.Background(), params)
		case findItemsByProduct:
			_, err = fc.FindItemsByProduct(context.Background(), params)
		case findItemsInEBayStores:
			_, err = fc.FindItemsInEBayStores(context.Background(), params)
		default:
			t.Fatalf("unsupported findMethod: %s", findMethod)
		}
		if err == nil {
			t.Fatal("err == nil; want != nil")
		}
		want := ebay.APIError{Err: ebay.ErrFailedRequest, StatusCode: http.StatusInternalServerError}
		var got *ebay.APIError
		if !errors.As(err, &got) {
			t.Fatalf("error %v does not wrap ebay.APIError", err)
		}
		if !strings.HasPrefix(got.Error(), want.Error()) {
			t.Errorf("expected error with prefix %q, got: %q", want.Error(), got.Error())
		}
		if got.StatusCode != want.StatusCode {
			t.Errorf("got status %q, want %q", got.StatusCode, want.StatusCode)
		}
	})

	badStatusCodes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusPaymentRequired,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusNotAcceptable,
		http.StatusProxyAuthRequired,
		http.StatusRequestTimeout,
		http.StatusConflict,
		http.StatusGone,
		http.StatusLengthRequired,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType,
		http.StatusRequestedRangeNotSatisfiable,
		http.StatusExpectationFailed,
		http.StatusTeapot,
		http.StatusMisdirectedRequest,
		http.StatusUnprocessableEntity,
		http.StatusLocked,
		http.StatusFailedDependency,
		http.StatusTooEarly,
		http.StatusUpgradeRequired,
		http.StatusPreconditionRequired,
		http.StatusTooManyRequests,
		http.StatusRequestHeaderFieldsTooLarge,
		http.StatusUnavailableForLegalReasons,
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
		http.StatusVariantAlsoNegotiates,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected,
		http.StatusNotExtended,
		http.StatusNetworkAuthenticationRequired,
	}

	t.Run("returns error if the client request was not successful", func(t *testing.T) {
		t.Parallel()
		for _, statusCode := range badStatusCodes {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
			}))
			defer ts.Close()
			client := ts.Client()
			fc := ebay.NewFindingClient(client, appID)
			fc.URL = ts.URL

			var err error
			switch findMethod {
			case findItemsByCategories:
				_, err = fc.FindItemsByCategories(context.Background(), params)
			case findItemsByKeywords:
				_, err = fc.FindItemsByKeywords(context.Background(), params)
			case findItemsAdvanced:
				_, err = fc.FindItemsAdvanced(context.Background(), params)
			case findItemsByProduct:
				_, err = fc.FindItemsByProduct(context.Background(), params)
			case findItemsInEBayStores:
				_, err = fc.FindItemsInEBayStores(context.Background(), params)
			default:
				t.Fatalf("unsupported findMethod: %s", findMethod)
			}
			want := fmt.Errorf("%w %d", ebay.ErrInvalidStatus, statusCode)
			assertAPIError(t, err, want, http.StatusInternalServerError)
		}
	})

	t.Run("returns error if the response cannot be parsed into find items response", func(t *testing.T) {
		t.Parallel()
		badData := `[123.1, 234.2]`
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(badData))
			if err != nil {
				t.Fatal(err)
			}
		}))
		defer ts.Close()
		client := ts.Client()
		fc := ebay.NewFindingClient(client, appID)
		fc.URL = ts.URL

		var err error
		switch findMethod {
		case findItemsByCategories:
			_, err = fc.FindItemsByCategories(context.Background(), params)
		case findItemsByKeywords:
			_, err = fc.FindItemsByKeywords(context.Background(), params)
		case findItemsAdvanced:
			_, err = fc.FindItemsAdvanced(context.Background(), params)
		case findItemsByProduct:
			_, err = fc.FindItemsByProduct(context.Background(), params)
		case findItemsInEBayStores:
			_, err = fc.FindItemsInEBayStores(context.Background(), params)
		default:
			t.Fatalf("unsupported findMethod: %s", findMethod)
		}
		want := fmt.Errorf("%w: json: cannot unmarshal array into Go value of type ebay.%sResponse", ebay.ErrDecodeAPIResponse, findMethod)
		assertAPIError(t, err, want, http.StatusInternalServerError)
	})

	for _, tc := range tcs {
		testCase := tc
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := json.Marshal(wantResp)
				if err != nil {
					t.Fatal(err)
				}
				w.WriteHeader(http.StatusOK)
				_, err = w.Write(body)
				if err != nil {
					t.Fatal(err)
				}
			}))
			defer ts.Close()
			client := ts.Client()
			fc := ebay.NewFindingClient(client, appID)
			fc.URL = ts.URL

			var resp ebay.ResultProvider
			var err error
			switch findMethod {
			case findItemsByCategories:
				resp, err = fc.FindItemsByCategories(context.Background(), testCase.Params)
			case findItemsByKeywords:
				resp, err = fc.FindItemsByKeywords(context.Background(), testCase.Params)
			case findItemsAdvanced:
				resp, err = fc.FindItemsAdvanced(context.Background(), testCase.Params)
			case findItemsByProduct:
				resp, err = fc.FindItemsByProduct(context.Background(), testCase.Params)
			case findItemsInEBayStores:
				resp, err = fc.FindItemsInEBayStores(context.Background(), testCase.Params)
			default:
				t.Fatalf("unsupported findMethod: %s", findMethod)
			}
			if testCase.Err != nil {
				assertAPIError(t, err, testCase.Err, http.StatusBadRequest)
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(resp, wantResp) {
					t.Errorf("got %v, want %v", resp, wantResp)
				}
			}
		})
	}
}

func assertAPIError(tb testing.TB, got, wantErr error, wantStatusCode int) {
	tb.Helper()
	var gotAPIError *ebay.APIError
	if !errors.As(got, &gotAPIError) {
		tb.Fatalf("error %v does not wrap ebay.APIError", got)
	}
	want := &ebay.APIError{Err: wantErr, StatusCode: wantStatusCode}
	if gotAPIError.Err.Error() != want.Err.Error() {
		tb.Errorf("got error %q, want %q", gotAPIError.Err.Error(), want.Err.Error())
	}
	if gotAPIError.StatusCode != want.StatusCode {
		tb.Errorf("got status %q, want %q", gotAPIError.StatusCode, want.StatusCode)
	}
}

func generateStringWithLen(length int, includeSpaces bool) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	var sbuilder strings.Builder
	charSet := letters
	if !includeSpaces {
		charSet = letters[:len(letters)-1] // Exclude the space character
	}
	for i := 0; i < length; i++ {
		sbuilder.WriteByte(charSet[i%len(charSet)])
	}
	return sbuilder.String()
}

func generateFilterParams(filterName string, count int) map[string]string {
	params := make(map[string]string)
	params["itemFilter.name"] = filterName
	for i := 0; i < count; i++ {
		params[fmt.Sprintf("itemFilter.value(%d)", i)] = strconv.Itoa(i)
	}
	return params
}
