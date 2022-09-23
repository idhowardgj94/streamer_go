package main_test

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewHub(t *testing.T) {
	tests := []struct {
		name string
		want *Hub
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHub(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServeWs(t *testing.T) {
	type args struct {
		hub *Hub
		w   http.ResponseWriter
		r   *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ServeWs(tt.args.hub, tt.args.w, tt.args.r)
		})
	}
}

func TestServeHome(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ServeHome(tt.args.w, tt.args.r)
		})
	}
}
