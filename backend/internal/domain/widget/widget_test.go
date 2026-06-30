package widget

import "testing"

func TestKindValidAcceptsKnownWidgetKinds(t *testing.T) {
	kinds := []Kind{
		KindMood,
		KindLoveTap,
		KindDrawing,
		KindPhoto,
		KindCountdown,
		KindTheirWorld,
	}

	for _, kind := range kinds {
		if !kind.Valid() {
			t.Fatalf("expected %q to be valid", kind)
		}
	}
}

func TestKindValidRejectsUnknownWidgetKind(t *testing.T) {
	if Kind("unknown").Valid() {
		t.Fatal("unknown widget kind should be invalid")
	}
}
