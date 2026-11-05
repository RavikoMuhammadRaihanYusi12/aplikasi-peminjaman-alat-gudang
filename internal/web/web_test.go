package web

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"alat-gudang/internal/store"
)

func TestRenderBeranda(t *testing.T) {
	renderer, err := NewRenderer(false)
	if err != nil {
		t.Fatalf("NewRenderer error: %v", err)
	}

	vm := store.DashboardVM{
		ActiveNav: "beranda",
	}

	var buf bytes.Buffer
	req := httptest.NewRequest("GET", "/", nil)
	err = renderer.RenderPage(&buf, req, "beranda.html", vm)
	if err != nil {
		t.Fatalf("RenderPage error: %v", err)
	}

	t.Logf("Rendered HTML size: %d bytes", buf.Len())
	if buf.Len() == 0 {
		t.Fatalf("Rendered HTML is empty!")
	}
}
