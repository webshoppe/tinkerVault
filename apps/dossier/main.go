//go:build windows

package main

//go:generate go-winres make --in winres/winres.json --out rsrc --arch amd64

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/jchv/go-webview2"
	"github.com/webshoppe/dossier/internal/app"
	"github.com/webshoppe/dossier/internal/dossier"
)

//go:embed ui/index.html ui/flag-data.inc.js
var uiFS embed.FS

func main() {
	// Log to a file next to the exe when running as windowsgui (no console).
	setupLogging()

	api := app.NewAPI()
	defer api.Close()

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     envDebug(),
		AutoFocus: true,
		// Keep WebView2 user data inside %LOCALAPPDATA%/Dossier/webview
		DataPath: webviewDataPath(),
		WindowOptions: webview2.WindowOptions{
			Title:  "Dossier",
			Width:  1280,
			Height: 840,
			Center: true,
			// RT_GROUP_ICON #1 embedded via go-winres (rsrc_windows_amd64.syso)
			IconId: 1,
		},
	})
	if w == nil {
		log.Fatal("Failed to create WebView2 window. Is the WebView2 Runtime installed?")
	}
	defer w.Destroy()

	w.SetSize(1280, 840, webview2.HintNone)

	// Bind API methods (appear as global JS functions returning Promises)
	mustBind(w, "GetStatus", api.GetStatus)
	mustBind(w, "OpenLastDossier", api.OpenLastDossier)
	mustBind(w, "PickAndOpenDossier", api.PickAndOpenDossier)
	mustBind(w, "CreateDossier", api.CreateDossier)
	mustBind(w, "OpenDossierPath", api.OpenDossierPath)
	mustBind(w, "ListRecentDossiers", api.ListRecentDossiers)
	mustBind(w, "SwitchDossier", api.SwitchDossier)
	mustBind(w, "CloseDossier", api.CloseDossier)
	mustBind(w, "ListDocuments", api.ListDocuments)
	mustBind(w, "ImportFiles", api.ImportFiles)
	mustBind(w, "ImportDropped", api.ImportDropped)
	mustBind(w, "GetDocument", api.GetDocument)
	mustBind(w, "DeleteDocument", api.DeleteDocument)
	mustBind(w, "OpenDocumentExternally", api.OpenDocumentExternally)
	mustBind(w, "OpenDocumentWithApp", api.OpenDocumentWithApp)
	mustBind(w, "PickOpenApp", api.PickOpenApp)
	mustBind(w, "SetOpenWithPreference", api.SetOpenWithPreference)
	mustBind(w, "Rescan", api.Rescan)
	mustBind(w, "ListNotes", api.ListNotes)
	mustBind(w, "CreateNote", api.CreateNote)
	mustBind(w, "GetNote", api.GetNote)
	mustBind(w, "SaveNote", api.SaveNote)
	mustBind(w, "DeleteNote", api.DeleteNote)
	mustBind(w, "ListStickies", api.ListStickies)
	mustBind(w, "CreateSticky", api.CreateSticky)
	mustBind(w, "UpdateSticky", api.UpdateSticky)
	mustBind(w, "DeleteSticky", api.DeleteSticky)
	mustBind(w, "StickyMeta", api.StickyMeta)
	mustBind(w, "ListCanvases", api.ListCanvases)
	mustBind(w, "CreateCanvas", api.CreateCanvas)
	mustBind(w, "GetCanvas", api.GetCanvas)
	mustBind(w, "SaveCanvas", api.SaveCanvas)
	mustBind(w, "DeleteCanvas", api.DeleteCanvas)
	mustBind(w, "ListKanbanBoards", api.ListKanbanBoards)
	mustBind(w, "CreateKanbanBoard", api.CreateKanbanBoard)
	mustBind(w, "GetKanbanBoard", api.GetKanbanBoard)
	mustBind(w, "SaveKanbanBoard", api.SaveKanbanBoard)
	mustBind(w, "DeleteKanbanBoard", api.DeleteKanbanBoard)
	mustBind(w, "ListDecisions", api.ListDecisions)
	mustBind(w, "CreateDecision", api.CreateDecision)
	mustBind(w, "UpdateDecision", api.UpdateDecision)
	mustBind(w, "AttachDocumentVersion", api.AttachDocumentVersion)
	mustBind(w, "DeleteDecision", api.DeleteDecision)
	mustBind(w, "OpenDecisionVersion", api.OpenDecisionVersion)
	mustBind(w, "Search", api.Search)
	mustBind(w, "FormatSize", api.FormatSize)
	mustBind(w, "OpenDossierFolder", api.OpenDossierFolder)
	mustBind(w, "GetAppSettings", api.GetAppSettings)
	mustBind(w, "SaveAppSettings", api.SaveAppSettings)
	mustBind(w, "DismissIntro", api.DismissIntro)
	mustBind(w, "ResetIntros", api.ResetIntros)
	mustBind(w, "AskDossier", api.AskDossier)

	// Early script: prevent WebView2 from navigating/opening dropped files before
	// our page handlers run. dragover must call preventDefault to become a drop target.
	w.Init(`
document.addEventListener('dragover', function(e){ e.preventDefault(); e.stopPropagation(); }, true);
document.addEventListener('drop', function(e){ e.preventDefault(); e.stopPropagation(); }, true);
`)

	// Boot auto-open (before UI so GetStatus().open is true when successful):
	//  1) DOSSIER_AUTO_OPEN=1; force-on for debug/automation (unchanged, separate concept).
	//  2) Else Settings.autoOpenLast; opt-in user preference (default off).
	// Invalid/missing last path never hangs: log and fall through to launcher.
	if os.Getenv("DOSSIER_AUTO_OPEN") == "1" {
		if _, err := api.OpenLastDossier(); err != nil {
			log.Printf("DOSSIER_AUTO_OPEN: %v", err)
		} else {
			log.Printf("DOSSIER_AUTO_OPEN: opened last dossier")
		}
	} else if shouldAutoOpenLastFromSettings() {
		if _, err := api.OpenLastDossier(); err != nil {
			log.Printf("autoOpenLast: %v", err)
		} else {
			log.Printf("autoOpenLast: opened last dossier")
		}
	}

	html, err := loadUI()
	if err != nil {
		log.Fatal(err)
	}
	w.SetHtml(html)
	w.Run()
}

func mustBind(w webview2.WebView, name string, f interface{}) {
	if err := w.Bind(name, f); err != nil {
		log.Fatalf("bind %s: %v", name, err)
	}
}

func loadUI() (string, error) {
	b, err := fs.ReadFile(uiFS, "ui/index.html")
	if err != nil {
		return "", err
	}
	html := string(b)
	// Inject offline Twemoji flag SVGs; Windows Segoe UI Emoji cannot glyph regional-indicator flags
	// (renders as CA/US letter pairs). Visual flag display uses embedded SVG images.
	if flags, err := fs.ReadFile(uiFS, "ui/flag-data.inc.js"); err == nil {
		html = strings.Replace(html, "/*__FLAG_DATA__*/", string(flags), 1)
	}
	return html, nil
}

func envDebug() bool {
	return os.Getenv("DOSSIER_DEBUG") == "1"
}

// shouldAutoOpenLastFromSettings is true only when the user opted in and the
// remembered path still looks like a dossier (IsDossierFolder). Does not create
// a new workspace from a random leftover folder.
func shouldAutoOpenLastFromSettings() bool {
	cfg, err := dossier.LoadConfig()
	if err != nil || cfg == nil {
		return false
	}
	if !cfg.Settings.AutoOpenLastOrDefault() {
		return false
	}
	path := strings.TrimSpace(cfg.LastDossierPath)
	if path == "" || !dossier.IsDossierFolder(path) {
		return false
	}
	return true
}

func webviewDataPath() string {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return ""
	}
	return local + `\Dossier\webview`
}

func setupLogging() {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return
	}
	dir := local + `\Dossier`
	_ = os.MkdirAll(dir, 0o755)
	f, err := os.OpenFile(dir+`\dossier.log`, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
