package dossier

import (
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher incrementally reindexes files under documents/ and notes/.
type Watcher struct {
	store    *Store
	watcher  *fsnotify.Watcher
	stopCh   chan struct{}
	debounce map[string]time.Time
	mu       sync.Mutex
	onChange func(event string)
}

func NewWatcher(store *Store, onChange func(string)) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	ww := &Watcher{
		store:    store,
		watcher:  w,
		stopCh:   make(chan struct{}),
		debounce: map[string]time.Time{},
		onChange: onChange,
	}
	// Watch documents and notes dirs
	for _, sub := range []string{DocumentsDir, NotesDir} {
		dir := filepath.Join(store.Root, sub)
		if err := w.Add(dir); err != nil {
			log.Printf("watcher: add %s: %v", dir, err)
		}
	}
	go ww.loop()
	return ww, nil
}

func (w *Watcher) Close() error {
	close(w.stopCh)
	return w.watcher.Close()
}

func (w *Watcher) loop() {
	for {
		select {
		case <-w.stopCh:
			return
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		case ev, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handle(ev)
		}
	}
}

func (w *Watcher) handle(ev fsnotify.Event) {
	name := filepath.Base(ev.Name)
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, "~") || strings.HasSuffix(name, ".tmp") {
		return
	}
	// Debounce rapid write storms
	w.mu.Lock()
	w.debounce[ev.Name] = time.Now()
	w.mu.Unlock()

	path := ev.Name
	time.AfterFunc(400*time.Millisecond, func() {
		w.mu.Lock()
		t, ok := w.debounce[path]
		if !ok || time.Since(t) < 350*time.Millisecond {
			w.mu.Unlock()
			return
		}
		delete(w.debounce, path)
		w.mu.Unlock()
		w.process(path, ev.Op)
	})
}

func (w *Watcher) process(path string, op fsnotify.Op) {
	rel, err := filepath.Rel(w.store.Root, path)
	if err != nil {
		return
	}
	rel = filepath.ToSlash(rel)
	isDoc := strings.HasPrefix(rel, DocumentsDir+"/")
	isNote := strings.HasPrefix(rel, NotesDir+"/")
	if !isDoc && !isNote {
		return
	}

	if op&fsnotify.Remove != 0 || op&fsnotify.Rename != 0 {
		if isDoc {
			if d, err := w.store.GetDocumentByRelPath(rel); err == nil && d != nil {
				// Remove DB row only (file already gone)
				w.store.mu.Lock()
				_, _ = w.store.db.Exec(`DELETE FROM documents WHERE id=?`, d.ID)
				_ = w.store.removeFTS(d.ID)
				w.store.mu.Unlock()
			}
		}
		if isNote {
			w.store.mu.Lock()
			var id string
			err := w.store.db.QueryRow(`SELECT id FROM notes WHERE rel_path=?`, rel).Scan(&id)
			if err == nil {
				_, _ = w.store.db.Exec(`DELETE FROM notes WHERE id=?`, id)
				_ = w.store.removeFTS(id)
			}
			w.store.mu.Unlock()
		}
		if w.onChange != nil {
			w.onChange("remove:" + rel)
		}
		return
	}

	if op&fsnotify.Write != 0 || op&fsnotify.Create != 0 {
		if isDoc {
			if _, err := w.store.AttachExisting(path); err != nil {
				log.Printf("watcher reindex doc %s: %v", path, err)
			}
		}
		if isNote {
			if _, err := w.store.RescanNotes(); err != nil {
				log.Printf("watcher reindex notes: %v", err)
			}
		}
		if w.onChange != nil {
			w.onChange("update:" + rel)
		}
	}
}
