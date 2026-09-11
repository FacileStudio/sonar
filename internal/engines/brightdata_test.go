package engines

import "testing"

func TestDecodeBrightDataStringBody(t *testing.T) {
	raw := []byte(`{"status_code":200,"body":"{\"organic\":[{\"title\":\"Kubernetes\",\"display_link\":\"https://kubernetes.io\",\"description\":\"open source\"}]}"}`)
	inner, status, err := decodeBrightData(raw)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if len(inner.Organic) != 1 || inner.Organic[0].Title != "Kubernetes" {
		t.Fatalf("organic mismatch: %+v", inner.Organic)
	}
}

func TestDecodeBrightDataObjectBody(t *testing.T) {
	raw := []byte(`{"status_code":200,"body":{"organic":[{"title":"Golang","display_link":"https://go.dev"}]}}`)
	inner, status, err := decodeBrightData(raw)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if len(inner.Organic) != 1 || inner.Organic[0].Title != "Golang" {
		t.Fatalf("organic mismatch: %+v", inner.Organic)
	}
}

func TestDecodeBrightDataUpstreamError(t *testing.T) {
	raw := []byte(`{"status_code":502,"headers":{"x-brd-error-code":"captcha"},"body":""}`)
	inner, status, err := decodeBrightData(raw)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if status != 502 {
		t.Fatalf("status = %d, want 502", status)
	}
	if len(inner.Organic) != 0 {
		t.Fatalf("expected no organic, got %d", len(inner.Organic))
	}
}
