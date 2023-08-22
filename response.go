// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

package ebay

import "time"

// A ResultProvider represents results from eBay Finding API endpoints.
type ResultProvider interface {
	// Results returns results from the eBay Finding API endpoint.
	Results() []FindItemsResponse
}

// FindItemsByCategoriesResponse represents the response from [FindingClient.FindItemsByCategories].
type FindItemsByCategoriesResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsByCategoryResponse"`
}

func (r FindItemsByCategoriesResponse) Results() []FindItemsResponse {
	return r.ItemsResponse
}

// FindItemsByKeywordsResponse represents the response from [FindingClient.FindItemsByKeywords].
type FindItemsByKeywordsResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsByKeywordsResponse"`
}

func (r FindItemsByKeywordsResponse) Results() []FindItemsResponse {
	return r.ItemsResponse
}

// FindItemsAdvancedResponse represents the response from [FindingClient.FindItemsAdvanced].
type FindItemsAdvancedResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsAdvancedResponse"`
}

func (r FindItemsAdvancedResponse) Results() []FindItemsResponse {
	return r.ItemsResponse
}

// FindItemsByProductResponse represents the response from [FindingClient.FindItemsByProduct].
type FindItemsByProductResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsByProductResponse"`
}

func (r FindItemsByProductResponse) Results() []FindItemsResponse {
	return r.ItemsResponse
}

// FindItemsInEBayStoresResponse represents the response from [FindingClient.FindItemsInEBayStores].
type FindItemsInEBayStoresResponse struct {
	ItemsResponse []FindItemsResponse `json:"findItemsIneBayStoresResponse"`
}

func (r FindItemsInEBayStoresResponse) Results() []FindItemsResponse {
	return r.ItemsResponse
}

// FindItemsResponse represents the base response container for all Finding Service operations.
//
// See [BaseServiceResponse] for details about generic response fields.
// See [BaseFindingServiceResponse] for details about fields specific to the Finding API.
//
// [BaseServiceResponse]: https://developer.ebay.com/Devzone/finding/CallRef/types/BaseServiceResponse.html
// [BaseFindingServiceResponse]: https://developer.ebay.com/Devzone/finding/CallRef/types/BaseFindingServiceResponse.html
type FindItemsResponse struct {
	Ack              []string           `json:"ack"`
	ErrorMessage     []ErrorMessage     `json:"errorMessage"`
	ItemSearchURL    []string           `json:"itemSearchURL"`
	PaginationOutput []PaginationOutput `json:"paginationOutput"`
	SearchResult     []SearchResult     `json:"searchResult"`
	Timestamp        []time.Time        `json:"timestamp"`
	Version          []string           `json:"version"`
}

// ErrorMessage is a message containing information regarding an error or warning that occurred
// when eBay processed the request. It is not returned when the ack value is Success.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/ErrorMessage.html.
type ErrorMessage struct {
	Error []ErrorData `json:"error"`
}

// ErrorData represents error details.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/ErrorData.html.
type ErrorData struct {
	Category    []string `json:"category"`
	Domain      []string `json:"domain"`
	ErrorID     []string `json:"errorId"`
	ExceptionID []string `json:"exceptionId"`
	Message     []string `json:"message"`
	Parameter   []string `json:"parameter"`
	Severity    []string `json:"severity"`
	Subdomain   []string `json:"subdomain"`
}

// PaginationOutput represents the pagination data for an item search.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/PaginationOutput.html.
type PaginationOutput struct {
	EntriesPerPage []string `json:"entriesPerPage"`
	PageNumber     []string `json:"pageNumber"`
	TotalEntries   []string `json:"totalEntries"`
	TotalPages     []string `json:"totalPages"`
}

// SearchResult represents returned item listings, if any.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/SearchResult.html.
type SearchResult struct {
	Count string       `json:"@count"`
	Item  []SearchItem `json:"item"`
}

// SearchItem represents the data of a single item that matches the search criteria.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/SearchItem.html.
type SearchItem struct {
	AutoPay                 []string            `json:"autoPay"`
	CharityID               []string            `json:"charityId"`
	Compatibility           []string            `json:"compatibility"`
	Condition               []Condition         `json:"condition"`
	Country                 []string            `json:"country"`
	DiscountPriceInfo       []DiscountPriceInfo `json:"discountPriceInfo"`
	Distance                []Distance          `json:"distance"`
	EBayPlusEnabled         []string            `json:"eBayPlusEnabled"`
	EekStatus               []string            `json:"eekStatus"`
	GalleryInfoContainer    []GalleryURL        `json:"galleryInfoContainer"`
	GalleryPlusPictureURL   []string            `json:"galleryPlusPictureURL"`
	GalleryURL              []string            `json:"galleryURL"`
	GlobalID                []string            `json:"globalId"`
	IsMultiVariationListing []string            `json:"isMultiVariationListing"`
	ItemID                  []string            `json:"itemId"`
	ListingInfo             []ListingInfo       `json:"listingInfo"`
	Location                []string            `json:"location"`
	PaymentMethod           []string            `json:"paymentMethod"`
	PictureURLLarge         []string            `json:"pictureURLLarge"`
	PictureURLSuperSize     []string            `json:"pictureURLSuperSize"`
	PostalCode              []string            `json:"postalCode"`
	PrimaryCategory         []Category          `json:"primaryCategory"`
	ProductID               []ProductID         `json:"productId"`
	ReturnsAccepted         []string            `json:"returnsAccepted"`
	SecondaryCategory       []Category          `json:"secondaryCategory"`
	SellerInfo              []SellerInfo        `json:"sellerInfo"`
	SellingStatus           []SellingStatus     `json:"sellingStatus"`
	ShippingInfo            []ShippingInfo      `json:"shippingInfo"`
	StoreInfo               []Storefront        `json:"storeInfo"`
	Subtitle                []string            `json:"subtitle"`
	Title                   []string            `json:"title"`
	TopRatedListing         []string            `json:"topRatedListing"`
	UnitPrice               []UnitPriceInfo     `json:"unitPrice"`
	ViewItemURL             []string            `json:"viewItemURL"`
}

// Condition describes an item's condition.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/Condition.html.
type Condition struct {
	ConditionDisplayName []string `json:"conditionDisplayName"`
	ConditionID          []string `json:"conditionId"`
}

// DiscountPriceInfo clarifies the discount treatment of an item that a seller can list.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/DiscountPriceInfo.html.
type DiscountPriceInfo struct {
	MinimumAdvertisedPriceExposure []string `json:"minimumAdvertisedPriceExposure"`
	OriginalRetailPrice            []Price  `json:"originalRetailPrice"`
	PricingTreatment               []string `json:"pricingTreatment"`
	SoldOffEbay                    []string `json:"soldOffEbay"`
	SoldOnEbay                     []string `json:"soldOnEbay"`
}

// Price specifies a monetary amount.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/Amount.html.
type Price struct {
	CurrencyID string `json:"@currencyId"`
	Value      string `json:"__value__"`
}

// Distance is the distance that the item is from the buyer, calculated using buyerPostalCode.
// The unit for distance varies by site, and is either miles or kilometers.
//
// This value is only returned for distance-based searches which involves specifying a buyerPostalCode
// and either sort by Distance, or use a combination of the MaxDistance LocalSearch itemFilters.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/Distance.html.
type Distance struct {
	Unit  string `json:"@unit"`
	Value string `json:"__value__"`
}

// GalleryURL is the URL for the Gallery thumbnail image.
// This value is only returned if the seller uploaded images for the item or
// the item was listed using a product identifier.
type GalleryURL struct {
	GallerySize string `json:"@gallerySize"`
	Value       string `json:"__value__"`
}

// ListingInfo represents information specific to an item listing.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/ListingInfo.html.
type ListingInfo struct {
	BestOfferEnabled       []string    `json:"bestOfferEnabled"`
	BuyItNowAvailable      []string    `json:"buyItNowAvailable"`
	BuyItNowPrice          []Price     `json:"buyItNowPrice"`
	ConvertedBuyItNowPrice []Price     `json:"convertedBuyItNowPrice"`
	EndTime                []time.Time `json:"endTime"`
	Gift                   []string    `json:"gift"`
	ListingType            []string    `json:"listingType"`
	StartTime              []time.Time `json:"startTime"`
	WatchCount             []string    `json:"watchCount"`
}

// Category represents details about a category.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/Category.html.
type Category struct {
	CategoryID   []string `json:"categoryId"`
	CategoryName []string `json:"categoryName"`
}

// ProductID represents the unique identifier for a single product.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/ProductId.html.
type ProductID struct {
	Type  string `json:"@type"`
	Value string `json:"__value__"`
}

// SellerInfo represents information about a listing's seller.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/SellerInfo.html.
type SellerInfo struct {
	FeedbackRatingStar      []string `json:"feedbackRatingStar"`
	FeedbackScore           []string `json:"feedbackScore"`
	PositiveFeedbackPercent []string `json:"positiveFeedbackPercent"`
	SellerUserName          []string `json:"sellerUserName"`
	TopRatedSeller          []string `json:"topRatedSeller"`
}

// SellingStatus represents an item's selling details.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/SellingStatus.html.
type SellingStatus struct {
	BidCount              []string `json:"bidCount"`
	ConvertedCurrentPrice []Price  `json:"convertedCurrentPrice"`
	CurrentPrice          []Price  `json:"currentPrice"`
	SellingState          []string `json:"sellingState"`
	TimeLeft              []string `json:"timeLeft"`
}

// ShippingInfo represents an item's shipping details.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/ShippingInfo.html.
type ShippingInfo struct {
	ExpeditedShipping       []string `json:"expeditedShipping"`
	HandlingTime            []string `json:"handlingTime"`
	IntermediatedShipping   []string `json:"intermediatedShipping"`
	OneDayShippingAvailable []string `json:"oneDayShippingAvailable"`
	ShippingServiceCost     []Price  `json:"shippingServiceCost"`
	ShippingType            []string `json:"shippingType"`
	ShipToLocations         []string `json:"shipToLocations"`
}

// Storefront denotes whether the item is a storefront listing.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/Storefront.html.
type Storefront struct {
	StoreName []string `json:"storeName"`
	StoreURL  []string `json:"storeURL"`
}

// UnitPriceInfo represents the type (e.g kg,lb) and quantity of a unit.
// See https://developer.ebay.com/Devzone/finding/CallRef/types/UnitPriceInfo.html.
type UnitPriceInfo struct {
	Quantity []string `json:"quantity"`
	Type     []string `json:"type"`
}
