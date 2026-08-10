package app

// Version is the product version string shown in the UI and returned by GetStatus.
// It must stay in sync with winres/winres.json (PE FileVersion / ProductVersion).
// 2.0.1 = Gap-fix: Notes/Decisions two-pane width (hard-pinned flex basis, was soft/shrinkable).
const Version = "2.0.1"
