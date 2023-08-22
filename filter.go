// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

package ebay

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"
	"unicode"
)

var (
	// ErrIncompleteFilterNameOnly is returned when a filter is missing the 'value' parameter.
	ErrIncompleteFilterNameOnly = errors.New("incomplete item filter: missing")

	// ErrIncompleteItemFilterParam is returned when an item filter is missing
	// either the 'paramName' or 'paramValue' parameter, as both 'paramName' and 'paramValue'
	// are required when either one is specified.
	ErrIncompleteItemFilterParam = errors.New("incomplete item filter: both paramName and paramValue must be specified together")

	// ErrUnsupportedItemFilterType is returned when an item filter 'name' parameter has an unsupported type.
	ErrUnsupportedItemFilterType = errors.New("unsupported item filter type")

	// ErrInvalidCountryCode is returned when an item filter 'values' parameter contains an invalid country code.
	ErrInvalidCountryCode = errors.New("invalid country code")

	// ErrInvalidCondition is returned when an item filter 'values' parameter contains an invalid condition ID or name.
	ErrInvalidCondition = errors.New("invalid condition")

	// ErrInvalidCurrencyID is returned when an item filter 'values' parameter contains an invalid currency ID.
	ErrInvalidCurrencyID = errors.New("invalid currency ID")

	// ErrInvalidDateTime is returned when an item filter 'values' parameter contains an invalid date time.
	ErrInvalidDateTime = errors.New("invalid date time value")

	maxExcludeCategories = 25

	// ErrMaxExcludeCategories is returned when an item filter 'values' parameter
	// contains more categories to exclude than the maximum allowed.
	ErrMaxExcludeCategories = fmt.Errorf("maximum categories to exclude is %d", maxExcludeCategories)

	maxExcludeSellers = 100

	// ErrMaxExcludeSellers is returned when an item filter 'values' parameter
	// contains more categories to exclude than the maximum allowed.
	ErrMaxExcludeSellers = fmt.Errorf("maximum sellers to exclude is %d", maxExcludeSellers)

	// ErrExcludeSellerCannotBeUsedWithSellers is returned when there is an attempt to use
	// the ExcludeSeller item filter together with either the Seller or TopRatedSellerOnly item filters.
	ErrExcludeSellerCannotBeUsedWithSellers = errors.New(
		"'ExcludeSeller' item filter cannot be used together with either the Seller or TopRatedSellerOnly item filters")

	// ErrInvalidInteger is returned when an item filter 'values' parameter contains an invalid integer.
	ErrInvalidInteger = errors.New("invalid integer")

	// ErrInvalidNumericFilter is returned when a numeric item filter is invalid.
	ErrInvalidNumericFilter = errors.New("invalid numeric item filter")

	// ErrInvalidExpeditedShippingType is returned when an item filter 'values' parameter
	// contains an invalid expedited shipping type.
	ErrInvalidExpeditedShippingType = errors.New("invalid expedited shipping type")

	// ErrInvalidAllListingType is returned when an item filter 'values' parameter
	// contains the 'All' listing type and other listing types.
	ErrInvalidAllListingType = errors.New("'All' listing type cannot be combined with other listing types")

	// ErrInvalidListingType is returned when an item filter 'values' parameter contains an invalid listing type.
	ErrInvalidListingType = errors.New("invalid listing type")

	// ErrDuplicateListingType is returned when an item filter 'values' parameter contains duplicate listing types.
	ErrDuplicateListingType = errors.New("duplicate listing type")

	// ErrInvalidAuctionListingTypes is returned when an item filter 'values' parameter
	// contains both 'Auction' and 'AuctionWithBIN' listing types.
	ErrInvalidAuctionListingTypes = errors.New("'Auction' and 'AuctionWithBIN' listing types cannot be combined")

	// ErrMaxDistanceMissing is returned when the LocalSearchOnly item filter is used,
	// but the MaxDistance item filter is missing in the request.
	ErrMaxDistanceMissing = errors.New("MaxDistance item filter is missing when using LocalSearchOnly item filter")

	maxLocatedIns = 25

	// ErrMaxLocatedIns is returned when an item filter 'values' parameter
	// contains more countries to locate items in than the maximum allowed.
	ErrMaxLocatedIns = fmt.Errorf("maximum countries to locate items in is %d", maxLocatedIns)

	// ErrInvalidPrice is returned when an item filter 'values' parameter contains an invalid price.
	ErrInvalidPrice = errors.New("invalid price")

	// ErrInvalidPriceParamName is returned when an item filter 'paramName' parameter
	// contains anything other than "Currency".
	ErrInvalidPriceParamName = errors.New(`invalid price parameter name, must be "Currency"`)

	// ErrInvalidMaxPrice is returned when an item filter 'values' parameter
	// contains a maximum price less than a minimum price.
	ErrInvalidMaxPrice = errors.New("maximum price must be greater than or equal to minimum price")

	maxSellers = 100

	// ErrMaxSellers is returned when an item filter 'values' parameter
	// contains more categories to include than the maximum allowed.
	ErrMaxSellers = fmt.Errorf("maximum sellers to include is %d", maxExcludeSellers)

	// ErrSellerCannotBeUsedWithOtherSellers is returned when there is an attempt to use
	// the Seller item filter together with either the ExcludeSeller or TopRatedSellerOnly item filters.
	ErrSellerCannotBeUsedWithOtherSellers = errors.New(
		"'Seller' item filter cannot be used together with either the ExcludeSeller or TopRatedSellerOnly item filters")

	// ErrMultipleSellerBusinessTypes is returned when an item filter 'values' parameter
	// contains multiple seller business types.
	ErrMultipleSellerBusinessTypes = errors.New("multiple seller business types found")

	// ErrInvalidSellerBusinessType is returned when an item filter 'values' parameter
	// contains an invalid seller business type.
	ErrInvalidSellerBusinessType = errors.New("invalid seller business type")

	// ErrTopRatedSellerCannotBeUsedWithSellers is returned when there is an attempt to use
	// the TopRatedSellerOnly item filter together with either the Seller or ExcludeSeller item filters.
	ErrTopRatedSellerCannotBeUsedWithSellers = errors.New(
		"'TopRatedSellerOnly' item filter cannot be used together with either the Seller or ExcludeSeller item filters")

	// ErrInvalidValueBoxInventory is returned when an item filter 'values' parameter
	// contains an invalid value box inventory.
	ErrInvalidValueBoxInventory = errors.New("invalid value box inventory")
)

func processAspectFilters(params map[string]string) ([]aspectFilter, error) {
	_, nonNumberedExists := params["aspectFilter.aspectName"]
	_, numberedExists := params["aspectFilter(0).aspectName"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidIndexSyntax
	}
	if nonNumberedExists {
		return processNonNumberedAspectFilter(params)
	}
	return processNumberedAspectFilters(params)
}

func processNonNumberedAspectFilter(params map[string]string) ([]aspectFilter, error) {
	filterValues, err := parseFilterValues(params, "aspectFilter.aspectValueName")
	if err != nil {
		return nil, err
	}
	filter := aspectFilter{
		aspectName:       params["aspectFilter.aspectName"],
		aspectValueNames: filterValues,
	}
	return []aspectFilter{filter}, nil
}

func processNumberedAspectFilters(params map[string]string) ([]aspectFilter, error) {
	var aspectFilters []aspectFilter
	for i := 0; ; i++ {
		name, ok := params[fmt.Sprintf("aspectFilter(%d).aspectName", i)]
		if !ok {
			break
		}
		filterValues, err := parseFilterValues(params, fmt.Sprintf("aspectFilter(%d).aspectValueName", i))
		if err != nil {
			return nil, err
		}
		aspectFilter := aspectFilter{
			aspectName:       name,
			aspectValueNames: filterValues,
		}
		aspectFilters = append(aspectFilters, aspectFilter)
	}
	return aspectFilters, nil
}

func processItemFilters(params map[string]string) ([]itemFilter, error) {
	_, nonNumberedExists := params["itemFilter.name"]
	_, numberedExists := params["itemFilter(0).name"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidIndexSyntax
	}
	if nonNumberedExists {
		return processNonNumberedItemFilter(params)
	}
	return processNumberedItemFilters(params)
}

func processNonNumberedItemFilter(params map[string]string) ([]itemFilter, error) {
	filterValues, err := parseFilterValues(params, "itemFilter.value")
	if err != nil {
		return nil, err
	}
	filter := itemFilter{
		name:   params["itemFilter.name"],
		values: filterValues,
	}
	pn, pnOk := params["itemFilter.paramName"]
	pv, pvOk := params["itemFilter.paramValue"]
	if pnOk != pvOk {
		return nil, ErrIncompleteItemFilterParam
	}
	if pnOk && pvOk {
		filter.paramName = &pn
		filter.paramValue = &pv
	}
	err = handleItemFilterType(&filter, nil, params)
	if err != nil {
		return nil, err
	}
	return []itemFilter{filter}, nil
}

func processNumberedItemFilters(params map[string]string) ([]itemFilter, error) {
	var itemFilters []itemFilter
	for i := 0; ; i++ {
		name, ok := params[fmt.Sprintf("itemFilter(%d).name", i)]
		if !ok {
			break
		}
		filterValues, err := parseFilterValues(params, fmt.Sprintf("itemFilter(%d).value", i))
		if err != nil {
			return nil, err
		}
		itemFilter := itemFilter{
			name:   name,
			values: filterValues,
		}
		pn, pnOk := params[fmt.Sprintf("itemFilter(%d).paramName", i)]
		pv, pvOk := params[fmt.Sprintf("itemFilter(%d).paramValue", i)]
		if pnOk != pvOk {
			return nil, ErrIncompleteItemFilterParam
		}
		if pnOk && pvOk {
			itemFilter.paramName = &pn
			itemFilter.paramValue = &pv
		}
		itemFilters = append(itemFilters, itemFilter)
	}
	for i := range itemFilters {
		err := handleItemFilterType(&itemFilters[i], itemFilters, params)
		if err != nil {
			return nil, err
		}
	}
	return itemFilters, nil
}

func parseFilterValues(params map[string]string, filterAttr string) ([]string, error) {
	var filterValues []string
	for i := 0; ; i++ {
		k := fmt.Sprintf("%s(%d)", filterAttr, i)
		if v, ok := params[k]; ok {
			filterValues = append(filterValues, v)
		} else {
			break
		}
	}
	if v, ok := params[filterAttr]; ok {
		filterValues = append(filterValues, v)
	}
	if len(filterValues) == 0 {
		return nil, fmt.Errorf("%w %q", ErrIncompleteFilterNameOnly, filterAttr)
	}
	_, nonNumberedExists := params[filterAttr]
	_, numberedExists := params[filterAttr+"(0)"]
	if nonNumberedExists && numberedExists {
		return nil, ErrInvalidIndexSyntax
	}
	return filterValues, nil
}

const (
	// ItemFilterType enumeration values from the eBay documentation.
	// See https://developer.ebay.com/devzone/finding/CallRef/types/ItemFilterType.html.
	authorizedSellerOnly  = "AuthorizedSellerOnly"
	availableTo           = "AvailableTo"
	bestOfferOnly         = "BestOfferOnly"
	charityOnly           = "CharityOnly"
	condition             = "Condition"
	currency              = "Currency"
	endTimeFrom           = "EndTimeFrom"
	endTimeTo             = "EndTimeTo"
	excludeAutoPay        = "ExcludeAutoPay"
	excludeCategory       = "ExcludeCategory"
	excludeSeller         = "ExcludeSeller"
	expeditedShippingType = "ExpeditedShippingType"
	feedbackScoreMax      = "FeedbackScoreMax"
	feedbackScoreMin      = "FeedbackScoreMin"
	freeShippingOnly      = "FreeShippingOnly"
	hideDuplicateItems    = "HideDuplicateItems"
	listedIn              = "ListedIn"
	listingType           = "ListingType"
	localPickupOnly       = "LocalPickupOnly"
	localSearchOnly       = "LocalSearchOnly"
	locatedIn             = "LocatedIn"
	lotsOnly              = "LotsOnly"
	maxBids               = "MaxBids"
	maxDistance           = "MaxDistance"
	maxHandlingTime       = "MaxHandlingTime"
	maxPrice              = "MaxPrice"
	maxQuantity           = "MaxQuantity"
	minBids               = "MinBids"
	minPrice              = "MinPrice"
	minQuantity           = "MinQuantity"
	modTimeFrom           = "ModTimeFrom"
	returnsAcceptedOnly   = "ReturnsAcceptedOnly"
	seller                = "Seller"
	sellerBusinessType    = "SellerBusinessType"
	soldItemsOnly         = "SoldItemsOnly"
	startTimeFrom         = "StartTimeFrom"
	startTimeTo           = "StartTimeTo"
	topRatedSellerOnly    = "TopRatedSellerOnly"
	valueBoxInventory     = "ValueBoxInventory"

	trueValue           = "true"
	falseValue          = "false"
	trueNum             = "1"
	falseNum            = "0"
	smallestMaxDistance = 5
)

// Valid Currency ID values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/Enums/currencyIdList.html.
var validCurrencyIDs = []string{
	"AUD", "CAD", "CHF", "CNY", "EUR", "GBP", "HKD", "INR", "MYR", "PHP", "PLN", "SEK", "SGD", "TWD", "USD",
}

func handleItemFilterType(filter *itemFilter, itemFilters []itemFilter, params map[string]string) error {
	switch filter.name {
	case authorizedSellerOnly, bestOfferOnly, charityOnly, excludeAutoPay, freeShippingOnly, hideDuplicateItems,
		localPickupOnly, lotsOnly, returnsAcceptedOnly, soldItemsOnly:
		if filter.values[0] != trueValue && filter.values[0] != falseValue {
			return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, filter.values[0])
		}
	case availableTo:
		if !isValidCountryCode(filter.values[0]) {
			return fmt.Errorf("%w: %q", ErrInvalidCountryCode, filter.values[0])
		}
	case condition:
		if !isValidCondition(filter.values[0]) {
			return fmt.Errorf("%w: %q", ErrInvalidCondition, filter.values[0])
		}
	case currency:
		if !slices.Contains(validCurrencyIDs, filter.values[0]) {
			return fmt.Errorf("%w: %q", ErrInvalidCurrencyID, filter.values[0])
		}
	case endTimeFrom, endTimeTo, startTimeFrom, startTimeTo:
		if !isValidDateTime(filter.values[0], true) {
			return fmt.Errorf("%w: %q", ErrInvalidDateTime, filter.values[0])
		}
	case excludeCategory:
		err := validateExcludeCategories(filter.values)
		if err != nil {
			return err
		}
	case excludeSeller:
		err := validateExcludeSellers(filter.values, itemFilters)
		if err != nil {
			return err
		}
	case expeditedShippingType:
		if filter.values[0] != "Expedited" && filter.values[0] != "OneDayShipping" {
			return fmt.Errorf("%w: %q", ErrInvalidExpeditedShippingType, filter.values[0])
		}
	case feedbackScoreMax, feedbackScoreMin:
		err := validateNumericFilter(filter, itemFilters, 0, feedbackScoreMax, feedbackScoreMin)
		if err != nil {
			return err
		}
	case listedIn:
		err := validateGlobalID(filter.values[0])
		if err != nil {
			return err
		}
	case listingType:
		err := validateListingTypes(filter.values)
		if err != nil {
			return err
		}
	case localSearchOnly:
		err := validateLocalSearchOnly(filter.values, itemFilters, params)
		if err != nil {
			return err
		}
	case locatedIn:
		err := validateLocatedIns(filter.values)
		if err != nil {
			return err
		}
	case maxBids, minBids:
		err := validateNumericFilter(filter, itemFilters, 0, maxBids, minBids)
		if err != nil {
			return err
		}
	case maxDistance:
		if _, ok := params["buyerPostalCode"]; !ok {
			return ErrBuyerPostalCodeMissing
		}
		if !isValidIntegerInRange(filter.values[0], smallestMaxDistance) {
			return invalidIntegerError(filter.values[0], smallestMaxDistance)
		}
	case maxHandlingTime:
		if !isValidIntegerInRange(filter.values[0], 1) {
			return invalidIntegerError(filter.values[0], 1)
		}
	case maxPrice, minPrice:
		err := validatePriceRange(filter, itemFilters)
		if err != nil {
			return err
		}
	case maxQuantity, minQuantity:
		err := validateNumericFilter(filter, itemFilters, 1, maxQuantity, minQuantity)
		if err != nil {
			return err
		}
	case modTimeFrom:
		if !isValidDateTime(filter.values[0], false) {
			return fmt.Errorf("%w: %q", ErrInvalidDateTime, filter.values[0])
		}
	case seller:
		err := validateSellers(filter.values, itemFilters)
		if err != nil {
			return err
		}
	case sellerBusinessType:
		err := validateSellerBusinessType(filter.values)
		if err != nil {
			return err
		}
	case topRatedSellerOnly:
		err := validateTopRatedSellerOnly(filter.values[0], itemFilters)
		if err != nil {
			return err
		}
	case valueBoxInventory:
		if filter.values[0] != trueNum && filter.values[0] != falseNum {
			return fmt.Errorf("%w: %q", ErrInvalidValueBoxInventory, filter.values[0])
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedItemFilterType, filter.name)
	}
	return nil
}

func isValidCountryCode(value string) bool {
	const countryCodeLen = 2
	if len(value) != countryCodeLen {
		return false
	}
	for _, r := range value {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// Valid Condition IDs from the eBay documentation.
// See https://developer.ebay.com/Devzone/finding/CallRef/Enums/conditionIdList.html#ConditionDefinitions.
var validConditionIDs = []int{1000, 1500, 1750, 2000, 2010, 2020, 2030, 2500, 2750, 3000, 4000, 5000, 6000, 7000}

func isValidCondition(value string) bool {
	cID, err := strconv.Atoi(value)
	if err == nil {
		return slices.Contains(validConditionIDs, cID)
	}
	// Value is a condition name, refer to the eBay documentation for condition name definitions.
	// See https://developer.ebay.com/Devzone/finding/CallRef/Enums/conditionIdList.html.
	return true
}

func isValidDateTime(value string, future bool) bool {
	dateTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false
	}
	if dateTime.Location() != time.UTC {
		return false
	}
	now := time.Now().UTC()
	if future && dateTime.Before(now) {
		return false
	}
	if !future && dateTime.After(now) {
		return false
	}
	return true
}

func validateExcludeCategories(values []string) error {
	if len(values) > maxExcludeCategories {
		return ErrMaxExcludeCategories
	}
	for _, v := range values {
		if !isValidIntegerInRange(v, 0) {
			return invalidIntegerError(v, 0)
		}
	}
	return nil
}

func validateExcludeSellers(values []string, itemFilters []itemFilter) error {
	if len(values) > maxExcludeSellers {
		return ErrMaxExcludeSellers
	}
	for _, f := range itemFilters {
		if f.name == seller || f.name == topRatedSellerOnly {
			return ErrExcludeSellerCannotBeUsedWithSellers
		}
	}
	return nil
}

func validateNumericFilter(
	filter *itemFilter, itemFilters []itemFilter, minAllowedValue int, filterA, filterB string,
) error {
	v, err := strconv.Atoi(filter.values[0])
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInteger, err)
	}
	if minAllowedValue > v {
		return invalidIntegerError(filter.values[0], minAllowedValue)
	}
	var filterAValue, filterBValue *int
	for _, f := range itemFilters {
		if f.name == filterA {
			val, err := strconv.Atoi(f.values[0])
			if err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidInteger, err)
			}
			filterAValue = &val
		} else if f.name == filterB {
			val, err := strconv.Atoi(f.values[0])
			if err != nil {
				return fmt.Errorf("%w: %w", ErrInvalidInteger, err)
			}
			filterBValue = &val
		}
	}
	if filterAValue != nil && filterBValue != nil && *filterBValue > *filterAValue {
		return fmt.Errorf("%w: %q must be greater than or equal to %q", ErrInvalidNumericFilter, filterA, filterB)
	}
	return nil
}

func invalidIntegerError(value string, min int) error {
	return fmt.Errorf("%w: %q (minimum value: %d)", ErrInvalidInteger, value, min)
}

func isValidIntegerInRange(value string, min int) bool {
	n, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	return n >= min
}

// Valid Listing Type values from the eBay documentation.
// See https://developer.ebay.com/devzone/finding/CallRef/types/ItemFilterType.html#ListingType.
var validListingTypes = []string{"Auction", "AuctionWithBIN", "Classified", "FixedPrice", "StoreInventory", "All"}

func validateListingTypes(values []string) error {
	seenTypes := make(map[string]bool)
	hasAuction, hasAuctionWithBIN := false, false
	for _, v := range values {
		if v == "All" && len(values) > 1 {
			return ErrInvalidAllListingType
		}
		found := false
		for _, lt := range validListingTypes {
			if v == lt {
				found = true
				if v == "Auction" {
					hasAuction = true
				} else if v == "AuctionWithBIN" {
					hasAuctionWithBIN = true
				}

				break
			}
		}
		if !found {
			return fmt.Errorf("%w: %q", ErrInvalidListingType, v)
		}
		if seenTypes[v] {
			return fmt.Errorf("%w: %q", ErrDuplicateListingType, v)
		}
		if hasAuction && hasAuctionWithBIN {
			return ErrInvalidAuctionListingTypes
		}
		seenTypes[v] = true
	}
	return nil
}

func validateLocalSearchOnly(values []string, itemFilters []itemFilter, params map[string]string) error {
	if _, ok := params["buyerPostalCode"]; !ok {
		return ErrBuyerPostalCodeMissing
	}
	foundMaxDistance := slices.ContainsFunc(itemFilters, func(f itemFilter) bool {
		return f.name == maxDistance
	})
	if !foundMaxDistance {
		return ErrMaxDistanceMissing
	}
	if values[0] != trueValue && values[0] != falseValue {
		return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, values[0])
	}
	return nil
}

func validateLocatedIns(values []string) error {
	if len(values) > maxLocatedIns {
		return ErrMaxLocatedIns
	}
	for _, v := range values {
		if !isValidCountryCode(v) {
			return fmt.Errorf("%w: %q", ErrInvalidCountryCode, v)
		}
	}
	return nil
}

func validatePriceRange(filter *itemFilter, itemFilters []itemFilter) error {
	price, err := parsePrice(filter)
	if err != nil {
		return err
	}
	var relatedFilterName string
	if filter.name == maxPrice {
		relatedFilterName = minPrice
	} else if filter.name == minPrice {
		relatedFilterName = maxPrice
	}
	for i := range itemFilters {
		if itemFilters[i].name == relatedFilterName {
			relatedPrice, err := parsePrice(&itemFilters[i])
			if err != nil {
				return err
			}
			if (filter.name == maxPrice && price < relatedPrice) ||
				(filter.name == minPrice && price > relatedPrice) {
				return ErrInvalidMaxPrice
			}
		}
	}
	return nil
}

func parsePrice(filter *itemFilter) (float64, error) {
	const minAllowedPrice float64 = 0.0
	price, err := strconv.ParseFloat(filter.values[0], 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidPrice, err)
	}
	if minAllowedPrice > price {
		return 0, fmt.Errorf("%w: %f (minimum value: %f)", ErrInvalidPrice, price, minAllowedPrice)
	}
	if filter.paramName != nil && *filter.paramName != currency {
		return 0, fmt.Errorf("%w: %q", ErrInvalidPriceParamName, *filter.paramName)
	}
	if filter.paramValue != nil && !slices.Contains(validCurrencyIDs, *filter.paramValue) {
		return 0, fmt.Errorf("%w: %q", ErrInvalidCurrencyID, *filter.paramValue)
	}
	return price, nil
}

func validateSellers(values []string, itemFilters []itemFilter) error {
	if len(values) > maxSellers {
		return ErrMaxSellers
	}
	for _, f := range itemFilters {
		if f.name == excludeSeller || f.name == topRatedSellerOnly {
			return ErrSellerCannotBeUsedWithOtherSellers
		}
	}
	return nil
}

func validateSellerBusinessType(values []string) error {
	if len(values) > 1 {
		return fmt.Errorf("%w", ErrMultipleSellerBusinessTypes)
	}
	if values[0] != "Business" && values[0] != "Private" {
		return fmt.Errorf("%w: %q", ErrInvalidSellerBusinessType, values[0])
	}
	return nil
}

func validateTopRatedSellerOnly(value string, itemFilters []itemFilter) error {
	if value != trueValue && value != falseValue {
		return fmt.Errorf("%w: %q", ErrInvalidBooleanValue, value)
	}
	for _, f := range itemFilters {
		if f.name == seller || f.name == excludeSeller {
			return ErrTopRatedSellerCannotBeUsedWithSellers
		}
	}
	return nil
}
