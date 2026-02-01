package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIClient_BasicAuth(t *testing.T) {
	username := "testuser"
	password := "testpass"
	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != expectedAuth {
			t.Errorf("Expected Authorization header %q, got %q", expectedAuth, auth)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, username, password)
	_, err := client.GetItems("")
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
}

func TestAPIClient_NoAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" {
			t.Errorf("Expected no Authorization header, got %q", auth)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, "", "")
	_, err := client.GetItems("")
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
}
