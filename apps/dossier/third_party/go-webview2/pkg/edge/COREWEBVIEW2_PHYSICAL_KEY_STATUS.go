package edge

// COREWEBVIEW2_PHYSICAL_KEY_STATUS matches the WebView2 C ABI.
// Windows BOOL is a 4-byte int, not a Go bool (1 byte). Using bool here
// misaligned WasKeyDown/IsKeyReleased and caused AcceleratorKeyCallback to be
// skipped whenever the library gated on !status.WasKeyDown (t6.3-diag finding).
type COREWEBVIEW2_PHYSICAL_KEY_STATUS struct {
	RepeatCount   uint32
	ScanCode      uint32
	IsExtendedKey int32 // BOOL
	IsMenuKeyDown int32 // BOOL
	WasKeyDown    int32 // BOOL
	IsKeyReleased int32 // BOOL
}
