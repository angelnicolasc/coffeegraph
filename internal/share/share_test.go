package share

import "testing"

func TestSanitize(t *testing.T) {
	in := "ok\nsk-ant-secret\nBearer abc\nhello\nGITHUB_TOKEN=x\nTELEGRAM_BOT_TOKEN=y"
	got := Sanitize(in)
	if got != "ok\nhello" {
		t.Fatalf("Sanitize() = %q, want %q", got, "ok\nhello")
	}
}
