package poteto

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/poteto-go/poteto/constant"
)

func TestCanTrust(t *testing.T) {
	iph := &ipHandler{}

	// test setter too
	iph.SetIsTrustPrivateIP(true)
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	// test register too
	iph.RegisterTrustIPRange(ipnet)

	tests := []struct {
		name     string
		ip       net.IP
		expected bool
	}{
		{"test trusted IP", net.ParseIP("10.0.0.1"), true},
		{"test untrusted IP", net.ParseIP("111.0.0.0"), false},
		{"test private IP", net.ParseIP("192.168.0.1"), true},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			result := iph.CanTrust(it.ip)
			if result != it.expected {
				t.Errorf("Not matched")
			}
		})
	}
}

func TestGetIPFromXFFHeader(t *testing.T) {
	iph := &ipHandler{}
	_, ipnet, _ := net.ParseCIDR("10.0.0.0/24")
	iph.RegisterTrustIPRange(ipnet)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(constant.HeaderXForwardedFor, "11.0.0.1, 12.0.0.1, 10.0.0.2, 10.0.0.1")
	ctx := NewContext(w, req).(*context)

	ipString, _ := iph.GetIPFromXFFHeader(ctx)
	if ipString != "12.0.0.1" {
		t.Errorf("Not matched")
	}
}

func TestGetRealIP(t *testing.T) {
	iph := &ipHandler{}
	_, ipnet, _ := net.ParseCIDR("13.0.0.0/24")
	iph.RegisterTrustIPRange(ipnet)

	tests := []struct {
		name        string
		headerKey   string
		headerValue string
		expected    string
	}{
		{
			"Get from Real Ip",
			constant.HeaderXRealIp,
			"11.0.0.1",
			"11.0.0.1",
		},
		{
			"Get from XFF",
			constant.HeaderXForwardedFor,
			"11.0.0.1",
			"11.0.0.1",
		},
		{
			"All trusted case",
			constant.HeaderXForwardedFor,
			"13.0.0.1",
			"13.0.0.1",
		},
		{
			"Get from RemoteAddr",
			"",
			"",
			"192.0.2.1",
		},
	}

	for _, it := range tests {
		t.Run(it.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			if it.headerKey != "" {
				req.Header.Set(it.headerKey, it.headerValue)
			}
			ctx := NewContext(w, req).(*context)

			ipString, _ := iph.RealIP(ctx)
			if ipString[0:8] != it.expected[0:8] {
				t.Errorf(ipString)
				t.Errorf("Not matched")
			}
		})
	}
}

func TestGetRemoteIpThrowError(t *testing.T) {
	iph := &ipHandler{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "hello"
	ctx := NewContext(w, req).(*context)

	if _, err := iph.GetRemoteIP(ctx); err == nil {
		t.Errorf("Don't throw error")
	}
}
