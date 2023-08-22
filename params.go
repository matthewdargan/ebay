// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

package ebay

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

const (
	findItemsByCategoryOperationName   = "findItemsByCategory"
	findItemsByKeywordsOperationName   = "findItemsByKeywords"
	findItemsAdvancedOperationName     = "findItemsAdvanced"
	findItemsByProductOperationName    = "findItemsByProduct"
	findItemsInEBayStoresOperationName = "findItemsIneBayStores"
	findingServiceVersion              = "1.0.0"
	findingResponseDataFormat          = "JSON"
)

var (
	// ErrCategoryIDMissing is returned when the 'categoryId' parameter is missing in a findItemsByCategory request.
	ErrCategoryIDMissing = errors.New("category ID parameter is missing")

	// ErrCategoryIDKeywordsMissing is returned when the 'categoryId' and 'keywords' parameters
	// are missing in a findItemsAdvanced request.
	ErrCategoryIDKeywordsMissing = errors.New("both category ID and keywords parameters are missing")

	// ErrProductIDMissing is returned when the 'productId' or 'productId.@type' parameters
	// are missing in a findItemsByProduct request.
	ErrProductIDMissing = errors.New("product ID parameter or product ID type are missing")

	// ErrCategoryIDKeywordsStoreNameMissing is returned when the 'categoryId', 'keywords', and 'storeName' parameters
	// are missing in a findItemsIneBayStores request.
	ErrCategoryIDKeywordsStoreNameMissing = errors.New("category ID, keywords, and store name parameters are missing")

	// ErrInvalidIndexSyntax is returned when index and non-index syntax are used in the params.
	ErrInvalidIndexSyntax = errors.New("invalid filter syntax: both index and non-index syntax are present")

	maxCategoryIDs = 3

	// ErrMaxCategoryIDs is returned when the 'categoryId' parameter contains more category IDs than the maximum allowed.
	ErrMaxCategoryIDs = fmt.Errorf("maximum category IDs to specify is %d", maxCategoryIDs)

	maxCategoryIDLen = 10

	// ErrInvalidCategoryIDLength is returned when an individual category ID in the 'categoryId' parameter
	// exceed the maximum length of 10 characters or is empty.
	ErrInvalidCategoryIDLength = fmt.Errorf("invalid category ID length: must be between 1 and %d characters", maxCategoryIDLen)

	// ErrInvalidCategoryID is returned when an individual category ID in the 'categoryId' parameter
	// contains an invalid category ID.
	ErrInvalidCategoryID = errors.New("invalid category ID")

	// ErrKeywordsMissing is returned when the 'keywords' parameter is missing.
	ErrKeywordsMissing = errors.New("keywords parameter is missing")

	minKeywordsLen, maxKeywordsLen = 2, 350

	// ErrInvalidKeywordsLength is returned when the 'keywords' parameter as a whole
	// exceeds the maximum length of 350 characters or has a length less than 2 characters.
	ErrInvalidKeywordsLength = fmt.Errorf("invalid keywords length: must be between %d and %d characters", minKeywordsLen, maxKeywordsLen)

	maxKeywordLen = 98

	// ErrInvalidKeywordLength is returned when an individual keyword in the 'keywords' parameter
	// exceeds the maximum length of 98 characters.
	ErrInvalidKeywordLength = fmt.Errorf("invalid keyword length: must be no more than %d characters", maxKeywordLen)

	// ErrInvalidProductIDLength is returned when the 'productId' parameter is empty.
	ErrInvalidProductIDLength = errors.New("invalid product ID length")

	isbnShortLen, isbnLongLen = 10, 13

	// ErrInvalidISBNLength is returned when the 'productId.type' parameter is an ISBN (International Standard Book Number)
	// and the 'productId' parameter is not exactly 10 or 13 characters.
	ErrInvalidISBNLength = fmt.Errorf("invalid ISBN length: must be either %d or %d characters", isbnShortLen, isbnLongLen)

	// ErrInvalidISBN is returned when the 'productId.type' parameter is an ISBN (International Standard Book Number)
	// and the 'productId' parameter contains an invalid ISBN.
	ErrInvalidISBN = errors.New("invalid ISBN")

	upcLen = 12

	// ErrInvalidUPCLength is returned when the 'productId.type' parameter is a UPC (Universal Product Code)
	// and the 'productId' parameter is not 12 digits.
	ErrInvalidUPCLength = fmt.Errorf("invalid UPC length: must be %d digits", upcLen)

	// ErrInvalidUPC is returned when the 'productId.type' parameter is a UPC (Universal Product Code)
	// and the 'productId' parameter contains an invalid UPC.
	ErrInvalidUPC = errors.New("invalid UPC")

	eanShortLen, eanLongLen = 8, 13

	// ErrInvalidEANLength is returned when the 'productId.type' parameter is an EAN (European Article Number)
	// and the 'productId' parameter is not exactly 8 or 13 characters.
	ErrInvalidEANLength = fmt.Errorf("invalid EAN length: must be either %d or %d characters", eanShortLen, eanLongLen)

	// ErrInvalidEAN is returned when the 'productId.type' parameter is an EAN (European Article Number)
	// and the 'productId' parameter contains an invalid EAN.
	ErrInvalidEAN = errors.New("invalid EAN")

	// ErrUnsupportedProductIDType is returned when the 'productId.type' parameter has an unsupported type.
	ErrUnsupportedProductIDType = errors.New("unsupported product ID type")

	// ErrInvalidStoreNameLength is returned when the 'storeName' parameter is empty.
	ErrInvalidStoreNameLength = errors.New("invalid store name length")

	// ErrInvalidStoreNameAmpersand is returned when the 'storeName' parameter contains unescaped '&' characters.
	ErrInvalidStoreNameAmpersand = errors.New("storeName contains unescaped '&' characters")

	// ErrInvalidGlobalID is returned when the 'Global-ID' or an item filter 'values' parameter contains an invalid global ID.
	ErrInvalidGlobalID = errors.New("invalid global ID")

	// ErrInvalidBooleanValue is returned when a parameter has an invalid boolean value.
	ErrInvalidBooleanValue = errors.New("invalid boolean value, allowed values are true and false")

	// ErrBuyerPostalCodeMissing is returned when the LocalSearchOnly, MaxDistance item filter,
	// or DistanceNearest sortOrder is used, but the buyerPostalCode parameter is missing in the request.
	ErrBuyerPostalCodeMissing = errors.New("buyerPostalCode is missing")

	// ErrInvalidOutputSelector is returned when the 'outputSelector' parameter contains an invalid output selector.
	ErrInvalidOutputSelector = errors.New("invalid output selector")

	maxCustomIDLen = 256

	// ErrInvalidCustomIDLength is returned when the 'affiliate.customId' parameter
	// exceeds the maximum length of 256 characters.
	ErrInvalidCustomIDLength = fmt.Errorf("invalid affiliate custom ID length: must be no more than %d characters", maxCustomIDLen)

	// ErrIncompleteAffiliateParams is returned when an affiliate is missing
	// either the 'networkId' or 'trackingId' parameter, as both 'networkId' and 'trackingId'
	// are required when either one is specified.
	ErrIncompleteAffiliateParams = errors.New("incomplete affiliate: both network and tracking IDs must be specified together")

	// ErrInvalidNetworkID is returned when the 'affiliate.networkId' parameter
	// contains an invalid network ID.
	ErrInvalidNetworkID = errors.New("invalid affiliate network ID")

	beFreeID, ebayPartnerNetworkID = 2, 9

	// ErrInvalidNetworkIDRange is returned when the 'affiliate.networkId' parameter
	// is outside the valid range of 2 (Be Free) and 9 (eBay Partner Network).
	ErrInvalidNetworkIDRange = fmt.Errorf("invalid affiliate network ID: must be between %d and %d", beFreeID, ebayPartnerNetworkID)

	// ErrInvalidTrackingID is returned when the 'affiliate.networkId' parameter is 9 (eBay Partner Network)
	// and the 'affiliate.trackingId' parameter contains an invalid tracking ID.
	ErrInvalidTrackingID = errors.New("invalid affiliate tracking ID")

	// ErrInvalidCampaignID is returned when the 'affiliate.networkId' parameter is 9 (eBay Partner Network)
	// and the 'affiliate.trackingId' parameter is not a 10-digit number (eBay Partner Network's Campaign ID).
	ErrInvalidCampaignID = errors.New("invalid affiliate Campaign ID length: must be a 10-digit number")

	// ErrInvalidPostalCode is returned when the 'buyerPostalCode' parameter contains an invalid postal code.
	ErrInvalidPostalCode = errors.New("invalid postal code")

	// ErrInvalidEntriesPerPage is returned when the 'paginationInput.entriesPerPage' parameter
	// contains an invalid entries value.
	ErrInvalidEntriesPerPage = errors.New("invalid pagination entries per page")

	minPaginationValue, maxPaginationValue = 1, 100

	// ErrInvalidEntriesPerPageRange is returned when the 'paginationInput.entriesPerPage' parameter
	// is outside the valid range of 1 to 100.
	ErrInvalidEntriesPerPageRange = fmt.Errorf("invalid pagination entries per page, must be between %d and %d", minPaginationValue, maxPaginationValue)

	// ErrInvalidPageNumber is returned when the 'paginationInput.pageNumber' parameter
	// contains an invalid pages value.
	ErrInvalidPageNumber = errors.New("invalid pagination page number")

	// ErrInvalidPageNumberRange is returned when the 'paginationInput.pageNumber' parameter
	// is outside the valid range of 1 to 100.
	ErrInvalidPageNumberRange = fmt.Errorf("invalid pagination page number, must be between %d and %d", minPaginationValue, maxPaginationValue)

	// ErrAuctionListingMissing is returned when the 'sortOrder' parameter BidCountFewest or BidCountMost,
	// but a 'Auction' listing type is not specified in the item filters.
	ErrAuctionListingMissing = errors.New("'Auction' listing type required for sorting by bid count")

	// ErrUnsupportedSortOrderType is returned when the 'sortOrder' parameter has an unsupported type.
	ErrUnsupportedSortOrderType = errors.New("invalid sort order type")

	// ErrInvalidRequest is returned when the eBay Finding API request is invalid.
	ErrInvalidRequest = errors.New("invalid request")
)

type findParamsValidator interface {
	validate(params map[string]string) error
	newRequest(ctx context.Context, url string) (*http.Request, error)
}

type findItemsByCategoryParams struct {
	appID           string
	globalID        *string
	aspectFilters   []aspectFilter
	categoryIDs     []string
	itemFilters     []itemFilter
	outputSelectors []string
	affiliate       *affiliate
	buyerPostalCode *string
	paginationInput *paginationInput
	sortOrder       *string
}

type aspectFilter struct {
	aspectName       string
	aspectValueNames []string
}

type itemFilter struct {
	name       string
	values     []string
	paramName  *string
	paramValue *string
}

type affiliate struct {
	customID     *string
	geoTargeting *string
	networkID    *string
	trackingID   *string
}

type paginationInput struct {
	entriesPerPage *string
	pageNumber     *string
}

func (fp *findItemsByCategoryParams) validate(params map[string]string) error {
	_, ok := params["categoryId"]
	_, nOk := params["categoryId(0)"]
	if !ok && !nOk {
		return ErrCategoryIDMissing
	}
	categoryIDs, err := processCategoryIDs(params)
	if err != nil {
		return err
	}
	fp.categoryIDs = categoryIDs
	globalID, ok := params["Global-ID"]
	if ok {
		err := validateGlobalID(globalID)
		if err != nil {
			return err
		}
		fp.globalID = &globalID
	}
	fp.aspectFilters, err = processAspectFilters(params)
	if err != nil {
		return err
	}
	fp.itemFilters, err = processItemFilters(params)
	if err != nil {
		return err
	}
	fp.outputSelectors, err = processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.affiliate, err = processAffiliate(params)
	if err != nil {
		return err
	}
	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}
	fp.paginationInput, err = processPaginationInput(params)
	if err != nil {
		return err
	}
	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}
	return nil
}

func (fp *findItemsByCategoryParams) newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	qry := req.URL.Query()
	if fp.globalID != nil {
		qry.Add("Global-ID", *fp.globalID)
	}
	qry.Add("OPERATION-NAME", findItemsByCategoryOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", fp.appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)
	for i, f := range fp.aspectFilters {
		qry.Add(fmt.Sprintf("aspectFilter(%d).aspectName", i), f.aspectName)
		for j, v := range f.aspectValueNames {
			qry.Add(fmt.Sprintf("aspectFilter(%d).aspectValueName(%d)", i, j), v)
		}
	}
	for i := range fp.categoryIDs {
		qry.Add(fmt.Sprintf("categoryId(%d)", i), fp.categoryIDs[i])
	}
	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}
		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}
	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}
	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}
	req.URL.RawQuery = qry.Encode()
	return req, nil
}

type findItemsByKeywordsParams struct {
	appID           string
	globalID        *string
	aspectFilters   []aspectFilter
	itemFilters     []itemFilter
	keywords        string
	outputSelectors []string
	affiliate       *affiliate
	buyerPostalCode *string
	paginationInput *paginationInput
	sortOrder       *string
}

func (fp *findItemsByKeywordsParams) validate(params map[string]string) error {
	keywords, err := processKeywords(params)
	if err != nil {
		return err
	}
	fp.keywords = keywords
	globalID, ok := params["Global-ID"]
	if ok {
		err := validateGlobalID(globalID)
		if err != nil {
			return err
		}
		fp.globalID = &globalID
	}
	fp.aspectFilters, err = processAspectFilters(params)
	if err != nil {
		return err
	}
	fp.itemFilters, err = processItemFilters(params)
	if err != nil {
		return err
	}
	fp.outputSelectors, err = processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.affiliate, err = processAffiliate(params)
	if err != nil {
		return err
	}
	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}
	fp.paginationInput, err = processPaginationInput(params)
	if err != nil {
		return err
	}
	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}
	return nil
}

func (fp *findItemsByKeywordsParams) newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	qry := req.URL.Query()
	if fp.globalID != nil {
		qry.Add("Global-ID", *fp.globalID)
	}
	qry.Add("OPERATION-NAME", findItemsByKeywordsOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", fp.appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)
	for i, f := range fp.aspectFilters {
		qry.Add(fmt.Sprintf("aspectFilter(%d).aspectName", i), f.aspectName)
		for j, v := range f.aspectValueNames {
			qry.Add(fmt.Sprintf("aspectFilter(%d).aspectValueName(%d)", i, j), v)
		}
	}
	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}
		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}
	qry.Add("keywords", fp.keywords)
	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}
	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}
	req.URL.RawQuery = qry.Encode()
	return req, nil
}

type findItemsAdvancedParams struct {
	appID             string
	globalID          *string
	aspectFilters     []aspectFilter
	categoryIDs       []string
	descriptionSearch *string
	itemFilters       []itemFilter
	keywords          *string
	outputSelectors   []string
	affiliate         *affiliate
	buyerPostalCode   *string
	paginationInput   *paginationInput
	sortOrder         *string
}

func (fp *findItemsAdvancedParams) validate(params map[string]string) error {
	_, cOk := params["categoryId"]
	_, csOk := params["categoryId(0)"]
	_, ok := params["keywords"]
	if !cOk && !csOk && !ok {
		return ErrCategoryIDKeywordsMissing
	}
	if cOk || csOk {
		categoryIDs, err := processCategoryIDs(params)
		if err != nil {
			return err
		}
		fp.categoryIDs = categoryIDs
	}
	if ok {
		keywords, err := processKeywords(params)
		if err != nil {
			return err
		}
		fp.keywords = &keywords
	}
	globalID, ok := params["Global-ID"]
	if ok {
		err := validateGlobalID(globalID)
		if err != nil {
			return err
		}
		fp.globalID = &globalID
	}
	aspectFilters, err := processAspectFilters(params)
	if err != nil {
		return err
	}
	fp.aspectFilters = aspectFilters
	ds, ok := params["descriptionSearch"]
	if ok {
		if ds != trueValue && ds != falseValue {
			return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, ds)
		}
		fp.descriptionSearch = &ds
	}
	fp.itemFilters, err = processItemFilters(params)
	if err != nil {
		return err
	}
	fp.outputSelectors, err = processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.affiliate, err = processAffiliate(params)
	if err != nil {
		return err
	}
	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}
	fp.paginationInput, err = processPaginationInput(params)
	if err != nil {
		return err
	}
	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}
	return nil
}

func (fp *findItemsAdvancedParams) newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	qry := req.URL.Query()
	if fp.globalID != nil {
		qry.Add("Global-ID", *fp.globalID)
	}
	qry.Add("OPERATION-NAME", findItemsAdvancedOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", fp.appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)
	for i, f := range fp.aspectFilters {
		qry.Add(fmt.Sprintf("aspectFilter(%d).aspectName", i), f.aspectName)
		for j, v := range f.aspectValueNames {
			qry.Add(fmt.Sprintf("aspectFilter(%d).aspectValueName(%d)", i, j), v)
		}
	}
	for i := range fp.categoryIDs {
		qry.Add(fmt.Sprintf("categoryId(%d)", i), fp.categoryIDs[i])
	}
	if fp.descriptionSearch != nil {
		qry.Add("descriptionSearch", *fp.descriptionSearch)
	}
	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}
		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}
	if fp.keywords != nil {
		qry.Add("keywords", *fp.keywords)
	}
	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}
	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}
	req.URL.RawQuery = qry.Encode()
	return req, nil
}

type findItemsByProductParams struct {
	appID           string
	globalID        *string
	itemFilters     []itemFilter
	outputSelectors []string
	product         productID
	affiliate       *affiliate
	buyerPostalCode *string
	paginationInput *paginationInput
	sortOrder       *string
}

type productID struct {
	idType string
	value  string
}

func (fp *findItemsByProductParams) validate(params map[string]string) error {
	productIDType, ptOk := params["productId.@type"]
	productValue, pOk := params["productId"]
	if !ptOk || !pOk {
		return ErrProductIDMissing
	}
	fp.product = productID{idType: productIDType, value: productValue}
	err := fp.product.processProductID()
	if err != nil {
		return err
	}
	globalID, ok := params["Global-ID"]
	if ok {
		err := validateGlobalID(globalID)
		if err != nil {
			return err
		}
		fp.globalID = &globalID
	}
	fp.itemFilters, err = processItemFilters(params)
	if err != nil {
		return err
	}
	fp.outputSelectors, err = processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.affiliate, err = processAffiliate(params)
	if err != nil {
		return err
	}
	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}
	fp.paginationInput, err = processPaginationInput(params)
	if err != nil {
		return err
	}
	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}
	return nil
}

func (fp *findItemsByProductParams) newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	qry := req.URL.Query()
	if fp.globalID != nil {
		qry.Add("Global-ID", *fp.globalID)
	}
	qry.Add("OPERATION-NAME", findItemsByProductOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", fp.appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)
	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}
		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}
	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}
	qry.Add("productId.@type", fp.product.idType)
	qry.Add("productId", fp.product.value)
	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}
	req.URL.RawQuery = qry.Encode()
	return req, nil
}

type findItemsInEBayStoresParams struct {
	appID           string
	globalID        *string
	aspectFilters   []aspectFilter
	categoryIDs     []string
	itemFilters     []itemFilter
	keywords        *string
	outputSelectors []string
	storeName       *string
	affiliate       *affiliate
	buyerPostalCode *string
	paginationInput *paginationInput
	sortOrder       *string
}

func (fp *findItemsInEBayStoresParams) validate(params map[string]string) error {
	_, cOk := params["categoryId"]
	_, csOk := params["categoryId(0)"]
	_, kwOk := params["keywords"]
	storeName, ok := params["storeName"]
	if !cOk && !csOk && !kwOk && !ok {
		return ErrCategoryIDKeywordsStoreNameMissing
	}
	if cOk || csOk {
		categoryIDs, err := processCategoryIDs(params)
		if err != nil {
			return err
		}
		fp.categoryIDs = categoryIDs
	}
	if kwOk {
		keywords, err := processKeywords(params)
		if err != nil {
			return err
		}
		fp.keywords = &keywords
	}
	if ok {
		err := validateStoreName(storeName)
		if err != nil {
			return err
		}
		fp.storeName = &storeName
	}
	globalID, ok := params["Global-ID"]
	if ok {
		err := validateGlobalID(globalID)
		if err != nil {
			return err
		}
		fp.globalID = &globalID
	}
	aspectFilters, err := processAspectFilters(params)
	if err != nil {
		return err
	}
	fp.aspectFilters = aspectFilters
	fp.itemFilters, err = processItemFilters(params)
	if err != nil {
		return err
	}
	fp.outputSelectors, err = processOutputSelectors(params)
	if err != nil {
		return err
	}
	fp.affiliate, err = processAffiliate(params)
	if err != nil {
		return err
	}
	buyerPostalCode, ok := params["buyerPostalCode"]
	if ok {
		if !isValidPostalCode(buyerPostalCode) {
			return ErrInvalidPostalCode
		}
		fp.buyerPostalCode = &buyerPostalCode
	}
	fp.paginationInput, err = processPaginationInput(params)
	if err != nil {
		return err
	}
	sortOrder, ok := params["sortOrder"]
	if ok {
		err := validateSortOrder(sortOrder, fp.itemFilters, fp.buyerPostalCode != nil)
		if err != nil {
			return err
		}
		fp.sortOrder = &sortOrder
	}
	return nil
}

func (fp *findItemsInEBayStoresParams) newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	qry := req.URL.Query()
	if fp.globalID != nil {
		qry.Add("Global-ID", *fp.globalID)
	}
	qry.Add("OPERATION-NAME", findItemsInEBayStoresOperationName)
	qry.Add("SERVICE-VERSION", findingServiceVersion)
	qry.Add("SECURITY-APPNAME", fp.appID)
	qry.Add("RESPONSE-DATA-FORMAT", findingResponseDataFormat)
	for i, f := range fp.aspectFilters {
		qry.Add(fmt.Sprintf("aspectFilter(%d).aspectName", i), f.aspectName)
		for j, v := range f.aspectValueNames {
			qry.Add(fmt.Sprintf("aspectFilter(%d).aspectValueName(%d)", i, j), v)
		}
	}
	for i := range fp.categoryIDs {
		qry.Add(fmt.Sprintf("categoryId(%d)", i), fp.categoryIDs[i])
	}
	for i, f := range fp.itemFilters {
		qry.Add(fmt.Sprintf("itemFilter(%d).name", i), f.name)
		for j, v := range f.values {
			qry.Add(fmt.Sprintf("itemFilter(%d).value(%d)", i, j), v)
		}
		if f.paramName != nil && f.paramValue != nil {
			qry.Add(fmt.Sprintf("itemFilter(%d).paramName", i), *f.paramName)
			qry.Add(fmt.Sprintf("itemFilter(%d).paramValue", i), *f.paramValue)
		}
	}
	if fp.keywords != nil {
		qry.Add("keywords", *fp.keywords)
	}
	for i := range fp.outputSelectors {
		qry.Add(fmt.Sprintf("outputSelector(%d)", i), fp.outputSelectors[i])
	}
	if fp.storeName != nil {
		qry.Add("storeName", *fp.storeName)
	}
	if fp.affiliate != nil {
		if fp.affiliate.customID != nil {
			qry.Add("affiliate.customId", *fp.affiliate.customID)
		}
		if fp.affiliate.geoTargeting != nil {
			qry.Add("affiliate.geoTargeting", *fp.affiliate.geoTargeting)
		}
		if fp.affiliate.networkID != nil {
			qry.Add("affiliate.networkId", *fp.affiliate.networkID)
		}
		if fp.affiliate.trackingID != nil {
			qry.Add("affiliate.trackingId", *fp.affiliate.trackingID)
		}
	}
	if fp.buyerPostalCode != nil {
		qry.Add("buyerPostalCode", *fp.buyerPostalCode)
	}
	if fp.paginationInput != nil {
		if fp.paginationInput.entriesPerPage != nil {
			qry.Add("paginationInput.entriesPerPage", *fp.paginationInput.entriesPerPage)
		}
		if fp.paginationInput.pageNumber != nil {
			qry.Add("paginationInput.pageNumber", *fp.paginationInput.pageNumber)
		}
	}
	if fp.sortOrder != nil {
		qry.Add("sortOrder", *fp.sortOrder)
	}
	req.URL.RawQuery = qry.Encode()
	return req, nil
}

func processCategoryIDs(params map[string]string) ([]string, error) {
	categoryID, nonNumberedExists := params["categoryId"]
	_, numberedExists := params["categoryId(0)"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidIndexSyntax
	}
	if nonNumberedExists {
		err := validateCategoryID(categoryID)
		if err != nil {
			return nil, err
		}
		return []string{categoryID}, nil
	}
	var categoryIDs []string
	for i := 0; ; i++ {
		cID, ok := params[fmt.Sprintf("categoryId(%d)", i)]
		if !ok {
			break
		}
		err := validateCategoryID(cID)
		if err != nil {
			return nil, err
		}
		categoryIDs = append(categoryIDs, cID)
		if len(categoryIDs) > maxCategoryIDs {
			return nil, ErrMaxCategoryIDs
		}
	}
	return categoryIDs, nil
}

func validateCategoryID(id string) error {
	if len(id) > maxCategoryIDLen {
		return ErrInvalidCategoryIDLength
	}
	_, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCategoryID, err)
	}
	return nil
}

func processKeywords(params map[string]string) (string, error) {
	keywords, ok := params["keywords"]
	if !ok {
		return "", ErrKeywordsMissing
	}
	if len(keywords) < minKeywordsLen || len(keywords) > maxKeywordsLen {
		return "", ErrInvalidKeywordsLength
	}
	individualKeywords := splitKeywords(keywords)
	for _, k := range individualKeywords {
		if len(k) > maxKeywordLen {
			return "", ErrInvalidKeywordLength
		}
	}
	return keywords, nil
}

// Split keywords based on special characters acting as search operators.
// See https://developer.ebay.com/api-docs/user-guides/static/finding-user-guide/finding-searching-by-keywords.html.
func splitKeywords(keywords string) []string {
	const specialChars = ` ,()"-*@+`
	return strings.FieldsFunc(keywords, func(r rune) bool {
		return strings.ContainsRune(specialChars, r)
	})
}

const (
	// Product ID type enumeration values from the eBay documentation.
	// See https://developer.ebay.com/Devzone/finding/CallRef/types/ProductId.html.
	referenceID = "ReferenceID"
	isbn        = "ISBN"
	upc         = "UPC"
	ean         = "EAN"
)

func (p *productID) processProductID() error {
	switch p.idType {
	case referenceID:
		if len(p.value) < 1 {
			return ErrInvalidProductIDLength
		}
	case isbn:
		if len(p.value) != isbnShortLen && len(p.value) != isbnLongLen {
			return ErrInvalidISBNLength
		}
		if !isValidISBN(p.value) {
			return ErrInvalidISBN
		}
	case upc:
		if len(p.value) != upcLen {
			return ErrInvalidUPCLength
		}
		if !isValidEAN(p.value) {
			return ErrInvalidUPC
		}
	case ean:
		if len(p.value) != eanShortLen && len(p.value) != eanLongLen {
			return ErrInvalidEANLength
		}
		if !isValidEAN(p.value) {
			return ErrInvalidEAN
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedProductIDType, p.idType)
	}
	return nil
}

func isValidISBN(isbn string) bool {
	if len(isbn) == isbnShortLen {
		var sum, acc int
		for i, r := range isbn {
			digit := int(r - '0')
			if !isDigit(digit) {
				if i == 9 && r == 'X' {
					digit = 10
				} else {
					return false
				}
			}

			acc += digit
			sum += acc
		}
		return sum%11 == 0
	}

	const altMultiplier = 3
	var sum int
	for i, r := range isbn {
		digit := int(r - '0')
		if !isDigit(digit) {
			return false
		}
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * altMultiplier
		}
	}
	return sum%10 == 0
}

func isDigit(digit int) bool {
	return digit >= 0 && digit <= 9
}

func isValidEAN(ean string) bool {
	const altMultiplier = 3
	var sum int
	for i, r := range ean[:len(ean)-1] {
		digit := int(r - '0')
		if !isDigit(digit) {
			return false
		}
		switch {
		case len(ean) == eanShortLen && i%2 == 0,
			len(ean) == eanLongLen && i%2 != 0,
			len(ean) == upcLen && i%2 == 0:
			sum += digit * altMultiplier
		default:
			sum += digit
		}
	}
	checkDigit := int(ean[len(ean)-1] - '0')
	if !isDigit(checkDigit) {
		return false
	}
	return (sum+checkDigit)%10 == 0
}

func validateStoreName(storeName string) error {
	if storeName == "" {
		return ErrInvalidStoreNameLength
	}
	if strings.Contains(storeName, "&") && !strings.Contains(storeName, "&amp;") {
		return ErrInvalidStoreNameAmpersand
	}
	return nil
}

// Valid Global ID values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/Enums/GlobalIdList.html.
var validGlobalIDs = []string{
	"EBAY-AT",
	"EBAY-AU",
	"EBAY-CH",
	"EBAY-DE",
	"EBAY-ENCA",
	"EBAY-ES",
	"EBAY-FR",
	"EBAY-FRBE",
	"EBAY-FRCA",
	"EBAY-GB",
	"EBAY-HK",
	"EBAY-IE",
	"EBAY-IN",
	"EBAY-IT",
	"EBAY-MOTOR",
	"EBAY-MY",
	"EBAY-NL",
	"EBAY-NLBE",
	"EBAY-PH",
	"EBAY-PL",
	"EBAY-SG",
	"EBAY-US",
}

func validateGlobalID(globalID string) error {
	if !slices.Contains(validGlobalIDs, globalID) {
		return fmt.Errorf("%w: %q", ErrInvalidGlobalID, globalID)
	}
	return nil
}

// Valid OutputSelectorType values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/callref/types/OutputSelectorType.html.
var validOutputSelectors = []string{
	"AspectHistogram",
	"CategoryHistogram",
	"ConditionHistogram",
	"GalleryInfo",
	"PictureURLLarge",
	"PictureURLSuperSize",
	"SellerInfo",
	"StoreInfo",
	"UnitPriceInfo",
}

func processOutputSelectors(params map[string]string) ([]string, error) {
	outputSelector, nonNumberedExists := params["outputSelector"]
	_, numberedExists := params["outputSelector(0)"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidIndexSyntax
	}
	if nonNumberedExists {
		if !slices.Contains(validOutputSelectors, outputSelector) {
			return nil, ErrInvalidOutputSelector
		}
		return []string{outputSelector}, nil
	}
	var os []string
	for i := 0; ; i++ {
		s, ok := params[fmt.Sprintf("outputSelector(%d)", i)]
		if !ok {
			break
		}
		if !slices.Contains(validOutputSelectors, s) {
			return nil, ErrInvalidOutputSelector
		}
		os = append(os, s)
	}
	return os, nil
}

func processAffiliate(params map[string]string) (*affiliate, error) {
	var aff affiliate
	customID, ok := params["affiliate.customId"]
	if ok {
		if len(customID) > maxCustomIDLen {
			return nil, ErrInvalidCustomIDLength
		}
		aff.customID = &customID
	}
	geoTargeting, ok := params["affiliate.geoTargeting"]
	if ok {
		if geoTargeting != trueValue && geoTargeting != falseValue {
			return nil, fmt.Errorf("%w: %q", ErrInvalidBooleanValue, geoTargeting)
		}
		aff.geoTargeting = &geoTargeting
	}
	networkID, nOk := params["affiliate.networkId"]
	trackingID, tOk := params["affiliate.trackingId"]
	if nOk != tOk {
		return nil, ErrIncompleteAffiliateParams
	}
	if !nOk {
		return &aff, nil
	}
	nID, err := strconv.Atoi(networkID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidNetworkID, err)
	}
	if nID < beFreeID || nID > ebayPartnerNetworkID {
		return nil, ErrInvalidNetworkIDRange
	}
	if nID == ebayPartnerNetworkID {
		err := validateTrackingID(trackingID)
		if err != nil {
			return nil, err
		}
	}
	aff.networkID = &networkID
	aff.trackingID = &trackingID
	return &aff, nil
}

func validateTrackingID(trackingID string) error {
	_, err := strconv.Atoi(trackingID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTrackingID, err)
	}
	const maxCampIDLen = 10
	if len(trackingID) != maxCampIDLen {
		return ErrInvalidCampaignID
	}
	return nil
}

func isValidPostalCode(postalCode string) bool {
	const minPostalCodeLen = 3
	return len(postalCode) >= minPostalCodeLen
}

func processPaginationInput(params map[string]string) (*paginationInput, error) {
	entriesPerPage, eOk := params["paginationInput.entriesPerPage"]
	pageNumber, pOk := params["paginationInput.pageNumber"]
	if !eOk && !pOk {
		return &paginationInput{}, nil
	}
	var pInput paginationInput
	if eOk {
		v, err := strconv.Atoi(entriesPerPage)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidEntriesPerPage, err)
		}
		if v < minPaginationValue || v > maxPaginationValue {
			return nil, ErrInvalidEntriesPerPageRange
		}
		pInput.entriesPerPage = &entriesPerPage
	}
	if pOk {
		v, err := strconv.Atoi(pageNumber)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidPageNumber, err)
		}
		if v < minPaginationValue || v > maxPaginationValue {
			return nil, ErrInvalidPageNumberRange
		}
		pInput.pageNumber = &pageNumber
	}
	return &pInput, nil
}

const (
	// SortOrderType enumeration values from the eBay documentation.
	// See https://developer.ebay.com/devzone/finding/CallRef/types/SortOrderType.html.
	bestMatch                = "BestMatch"
	bidCountFewest           = "BidCountFewest"
	bidCountMost             = "BidCountMost"
	countryAscending         = "CountryAscending"
	countryDescending        = "CountryDescending"
	currentPriceHighest      = "CurrentPriceHighest"
	distanceNearest          = "DistanceNearest"
	endTimeSoonest           = "EndTimeSoonest"
	pricePlusShippingHighest = "PricePlusShippingHighest"
	pricePlusShippingLowest  = "PricePlusShippingLowest"
	startTimeNewest          = "StartTimeNewest"
	watchCountDecreaseSort   = "WatchCountDecreaseSort"
)

func validateSortOrder(sortOrder string, itemFilters []itemFilter, hasBuyerPostalCode bool) error {
	switch sortOrder {
	case bestMatch, countryAscending, countryDescending, currentPriceHighest, endTimeSoonest,
		pricePlusShippingHighest, pricePlusShippingLowest, startTimeNewest, watchCountDecreaseSort:
		return nil
	case bidCountFewest, bidCountMost:
		hasAuctionListing := slices.ContainsFunc(itemFilters, func(f itemFilter) bool {
			return f.name == listingType && slices.Contains(f.values, "Auction")
		})
		if !hasAuctionListing {
			return ErrAuctionListingMissing
		}
	case distanceNearest:
		if !hasBuyerPostalCode {
			return ErrBuyerPostalCodeMissing
		}
	default:
		return ErrUnsupportedSortOrderType
	}
	return nil
}
