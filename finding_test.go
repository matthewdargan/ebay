// Copyright 2023 Matthew P. Dargan.
// SPDX-License-Identifier: Apache-2.0

package ebay

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewFindingClient(t *testing.T) {
	t.Parallel()
	client := http.DefaultClient
	appID := "ebay-app-id"
	got := NewFindingClient(client, appID)
	want := &FindingClient{
		Client: client,
		AppID:  appID,
		URL:    findingURL,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewFindingClient() = %v, want %v", got, want)
	}
}

func TestFindingClient_FindItemsAdvanced(t *testing.T) {
	t.Parallel()
	t.Run("ResponseSuccess", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&FindItemsAdvancedResponse{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer ts.Close()
		client := NewFindingClient(ts.Client(), "ebay-app-id")
		client.URL = ts.URL
		params := map[string]string{"categoryId": "123", "keywords": "testword"}
		got, err := client.FindItemsAdvanced(context.Background(), params)
		if err != nil {
			t.Errorf("FindingClient.FindItemsAdvanced() error = %v, want nil", err)
			return
		}
		want := &FindItemsAdvancedResponse{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FindingClient.FindItemsAdvanced() = %v, want %v", got, want)
		}
	})

	t.Run("HTTPNewRequestError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://example.com/\x00invalid"
		_, err := client.FindItemsAdvanced(context.Background(), map[string]string{})
		if !errors.Is(err, ErrNewRequest) {
			t.Errorf("FindingClient.FindItemsAdvanced() error = %v, want %v", err, ErrNewRequest)
		}
	})

	t.Run("ClientDoError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://localhost"
		_, err := client.FindItemsAdvanced(context.Background(), map[string]string{})
		if !errors.Is(err, ErrFailedRequest) {
			t.Errorf("FindingClient.FindItemsAdvanced() error = %v, want %v", err, ErrFailedRequest)
		}
	})

	t.Run("InvalidStatusError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsAdvanced(context.Background(), map[string]string{})
		if !errors.Is(err, ErrInvalidStatus) {
			t.Errorf("FindingClient.FindItemsAdvanced() error = %v, want %v", err, ErrInvalidStatus)
		}
	})

	t.Run("JSONDecodeError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`baddata123`))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsAdvanced(context.Background(), map[string]string{})
		if !errors.Is(err, ErrDecodeAPIResponse) {
			t.Errorf("FindingClient.FindItemsAdvanced() error = %v, want %v", err, ErrDecodeAPIResponse)
		}
	})
}

func TestFindingClient_FindItemsByCategory(t *testing.T) {
	t.Parallel()
	t.Run("ResponseSuccess", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&FindItemsByCategoryResponse{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer ts.Close()
		client := NewFindingClient(ts.Client(), "ebay-app-id")
		client.URL = ts.URL
		params := map[string]string{"categoryId": "123"}
		got, err := client.FindItemsByCategory(context.Background(), params)
		if err != nil {
			t.Errorf("FindingClient.FindItemsByCategory() error = %v, want nil", err)
			return
		}
		want := &FindItemsByCategoryResponse{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FindingClient.FindItemsByCategory() = %v, want %v", got, want)
		}
	})

	t.Run("HTTPNewRequestError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://example.com/\x00invalid"
		_, err := client.FindItemsByCategory(context.Background(), map[string]string{})
		if !errors.Is(err, ErrNewRequest) {
			t.Errorf("FindingClient.FindItemsByCategory() error = %v, want %v", err, ErrNewRequest)
		}
	})

	t.Run("ClientDoError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://localhost"
		_, err := client.FindItemsByCategory(context.Background(), map[string]string{})
		if !errors.Is(err, ErrFailedRequest) {
			t.Errorf("FindingClient.FindItemsByCategory() error = %v, want %v", err, ErrFailedRequest)
		}
	})

	t.Run("InvalidStatusError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsByCategory(context.Background(), map[string]string{})
		if !errors.Is(err, ErrInvalidStatus) {
			t.Errorf("FindingClient.FindItemsByCategory() error = %v, want %v", err, ErrInvalidStatus)
		}
	})

	t.Run("JSONDecodeError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`baddata123`))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsByCategory(context.Background(), map[string]string{})
		if !errors.Is(err, ErrDecodeAPIResponse) {
			t.Errorf("FindingClient.FindItemsByCategory() error = %v, want %v", err, ErrDecodeAPIResponse)
		}
	})
}

func TestFindingClient_FindItemsByKeywords(t *testing.T) {
	t.Parallel()
	t.Run("ResponseSuccess", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&FindItemsByKeywordsResponse{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer ts.Close()
		client := NewFindingClient(ts.Client(), "ebay-app-id")
		client.URL = ts.URL
		params := map[string]string{"keywords": "testword"}
		got, err := client.FindItemsByKeywords(context.Background(), params)
		if err != nil {
			t.Errorf("FindingClient.FindItemsByKeywords() error = %v, want nil", err)
			return
		}
		want := &FindItemsByKeywordsResponse{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FindingClient.FindItemsByKeywords() = %v, want %v", got, want)
		}
	})

	t.Run("HTTPNewRequestError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://example.com/\x00invalid"
		_, err := client.FindItemsByKeywords(context.Background(), map[string]string{})
		if !errors.Is(err, ErrNewRequest) {
			t.Errorf("FindingClient.FindItemsByKeywords() error = %v, want %v", err, ErrNewRequest)
		}
	})

	t.Run("ClientDoError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://localhost"
		_, err := client.FindItemsByKeywords(context.Background(), map[string]string{})
		if !errors.Is(err, ErrFailedRequest) {
			t.Errorf("FindingClient.FindItemsByKeywords() error = %v, want %v", err, ErrFailedRequest)
		}
	})

	t.Run("InvalidStatusError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsByKeywords(context.Background(), map[string]string{})
		if !errors.Is(err, ErrInvalidStatus) {
			t.Errorf("FindingClient.FindItemsByKeywords() error = %v, want %v", err, ErrInvalidStatus)
		}
	})

	t.Run("JSONDecodeError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`baddata123`))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsByKeywords(context.Background(), map[string]string{})
		if !errors.Is(err, ErrDecodeAPIResponse) {
			t.Errorf("FindingClient.FindItemsByKeywords() error = %v, want %v", err, ErrDecodeAPIResponse)
		}
	})
}

func TestFindingClient_FindItemsByProduct(t *testing.T) {
	t.Parallel()
	t.Run("ResponseSuccess", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&FindItemsByProductResponse{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer ts.Close()
		client := NewFindingClient(ts.Client(), "ebay-app-id")
		client.URL = ts.URL
		params := map[string]string{"productId.@type": "ReferenceID", "productId": "123"}
		got, err := client.FindItemsByProduct(context.Background(), params)
		if err != nil {
			t.Errorf("FindingClient.FindItemsByProduct() error = %v, want nil", err)
			return
		}
		want := &FindItemsByProductResponse{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FindingClient.FindItemsByProduct() = %v, want %v", got, want)
		}
	})

	t.Run("HTTPNewRequestError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://example.com/\x00invalid"
		_, err := client.FindItemsByProduct(context.Background(), map[string]string{})
		if !errors.Is(err, ErrNewRequest) {
			t.Errorf("FindingClient.FindItemsByProduct() error = %v, want %v", err, ErrNewRequest)
		}
	})

	t.Run("ClientDoError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://localhost"
		_, err := client.FindItemsByProduct(context.Background(), map[string]string{})
		if !errors.Is(err, ErrFailedRequest) {
			t.Errorf("FindingClient.FindItemsByProduct() error = %v, want %v", err, ErrFailedRequest)
		}
	})

	t.Run("InvalidStatusError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsByProduct(context.Background(), map[string]string{})
		if !errors.Is(err, ErrInvalidStatus) {
			t.Errorf("FindingClient.FindItemsByProduct() error = %v, want %v", err, ErrInvalidStatus)
		}
	})

	t.Run("JSONDecodeError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`baddata123`))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsByProduct(context.Background(), map[string]string{})
		if !errors.Is(err, ErrDecodeAPIResponse) {
			t.Errorf("FindingClient.FindItemsByProduct() error = %v, want %v", err, ErrDecodeAPIResponse)
		}
	})
}

func TestFindingClient_FindItemsInEBayStores(t *testing.T) {
	t.Parallel()
	t.Run("ResponseSuccess", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&FindItemsInEBayStoresResponse{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer ts.Close()
		client := NewFindingClient(ts.Client(), "ebay-app-id")
		client.URL = ts.URL
		params := map[string]string{"storeName": "teststore"}
		got, err := client.FindItemsInEBayStores(context.Background(), params)
		if err != nil {
			t.Errorf("FindingClient.FindItemsInEBayStores() error = %v, want nil", err)
			return
		}
		want := &FindItemsInEBayStoresResponse{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FindingClient.FindItemsInEBayStores() = %v, want %v", got, want)
		}
	})

	t.Run("HTTPNewRequestError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://example.com/\x00invalid"
		_, err := client.FindItemsInEBayStores(context.Background(), map[string]string{})
		if !errors.Is(err, ErrNewRequest) {
			t.Errorf("FindingClient.FindItemsInEBayStores() error = %v, want %v", err, ErrNewRequest)
		}
	})

	t.Run("ClientDoError", func(t *testing.T) {
		t.Parallel()
		client := NewFindingClient(http.DefaultClient, "ebay-app-id")
		client.URL = "http://localhost"
		_, err := client.FindItemsInEBayStores(context.Background(), map[string]string{})
		if !errors.Is(err, ErrFailedRequest) {
			t.Errorf("FindingClient.FindItemsInEBayStores() error = %v, want %v", err, ErrFailedRequest)
		}
	})

	t.Run("InvalidStatusError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsInEBayStores(context.Background(), map[string]string{})
		if !errors.Is(err, ErrInvalidStatus) {
			t.Errorf("FindingClient.FindItemsInEBayStores() error = %v, want %v", err, ErrInvalidStatus)
		}
	})

	t.Run("JSONDecodeError", func(t *testing.T) {
		t.Parallel()
		errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`baddata123`))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}))
		defer errorSrv.Close()
		client := NewFindingClient(errorSrv.Client(), "ebay-app-id")
		client.URL = errorSrv.URL
		_, err := client.FindItemsInEBayStores(context.Background(), map[string]string{})
		if !errors.Is(err, ErrDecodeAPIResponse) {
			t.Errorf("FindingClient.FindItemsInEBayStores() error = %v, want %v", err, ErrDecodeAPIResponse)
		}
	})
}
