package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caedis/noreza/internal/mapping"
	"github.com/caedis/noreza/internal/web/templates"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all connections
	CheckOrigin: func(r *http.Request) bool { return true },
}

//go:embed static
var staticFiles embed.FS

func RunServer(ctx context.Context, port int, store *mapping.Store, deviceDesc, serial string) {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServerFS(staticFiles))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		templates.Layout(deviceDesc, serial).Render(r.Context(), w)
	})

	mux.HandleFunc("GET /profiles", func(w http.ResponseWriter, r *http.Request) {
		templates.ProfileList(store.ListProfiles()).Render(r.Context(), w)
	})

	mux.HandleFunc("POST /profiles", func(w http.ResponseWriter, r *http.Request) {
		profileName := r.Header.Get("HX-Prompt")
		profileName = strings.ReplaceAll(profileName, ".", "")
		profileName = strings.ReplaceAll(profileName, ".json", "")
		profileName = strings.TrimSpace(profileName)
		if profileName == "" {
			http.Error(w, "profile name missing", http.StatusBadRequest)
			return
		}

		if err := store.CreateNewProfile(fmt.Sprintf("%s.json", profileName)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		templates.ProfileList(store.ListProfiles()).Render(r.Context(), w)
	})

	mux.HandleFunc("DELETE /profiles/{profile}", func(w http.ResponseWriter, r *http.Request) {
		profileName := r.PathValue("profile")

		// delete should trigger store.RemoveProfile
		if err := os.Remove(filepath.Join(store.ProfilePath, profileName+".json")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.EditorDefault(true).Render(r.Context(), w)

		templates.ProfileList(store.ListProfiles()).Render(r.Context(), w)
	})

	mux.HandleFunc("POST /profiles/{profile}/activate", func(w http.ResponseWriter, r *http.Request) {
		profile := r.PathValue("profile")
		if profile == "" {
			http.Error(w, "missing name", http.StatusBadRequest)
			return
		}

		if err := store.SetActiveProfile(profile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.ProfileList(store.ListProfiles()).Render(r.Context(), w)
	})

	mux.HandleFunc("GET /profiles/{profile}/update", func(w http.ResponseWriter, r *http.Request) {
		profile := r.PathValue("profile")
		vals := r.URL.Query()

		keyType := vals.Get("type")
		subKey := vals.Get("subkey")
		index, _ := strconv.Atoi(vals.Get("index"))
		templates.EditorModal(profile, uint8(index), keyType, subKey).Render(r.Context(), w)
	})

	mux.HandleFunc("PATCH /profiles/{profile}/update", func(w http.ResponseWriter, r *http.Request) {
		profile := r.PathValue("profile")

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		modeRaw, err := strconv.Atoi(r.PostForm.Get("mode"))
		if err != nil {
			http.Error(w, "invalid mode", http.StatusBadRequest)
			return
		}
		mode := mapping.KeyMode(modeRaw)

		value := r.PostForm.Get("value")
		var code int
		var ok bool
		switch mode {
		case mapping.Mouse:
			code, ok = mapping.MouseToCode[value]
		case mapping.Keyboard:
			code, ok = mapping.KeyToCode[value]
		}

		if !ok {
			http.Error(w, "unsupported key", http.StatusBadRequest)
			return
		}

		keyType := r.PostForm.Get("type")
		subKey := r.PostForm.Get("subkey")
		index, err := strconv.Atoi(r.PostForm.Get("index"))
		if err != nil {
			http.Error(w, "invalid index", http.StatusBadRequest)
			return
		}

		// Update mapping
		mappings := store.RawMappings.Load()
		m, ok := (*mappings)[profile]
		if !ok {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}

		m.UpdateBinding(keyType, subKey, uint8(index), mode, code)

		// Recompile FlatMapping
		flat := mapping.CompileFlatMapping(*m)
		maps := store.Mappings.Load()
		if maps != nil {
			(*maps)[profile] = flat
			store.Mappings.Store(maps)
		}

		// Update activeMapping if needed
		if store.ActiveProfile.Load() == profile {
			store.ActiveMapping.Store(flat)
		}

		if err := store.SaveProfile(profile); err != nil {
			http.Error(w, "error saving profile", http.StatusInternalServerError)
			return
		}

		device, err := mapping.GetDeviceFromID(store.ProductID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.Editor(*m, profile, device).Render(r.Context(), w)
	})

	mux.HandleFunc("GET /profiles/{profile}/editor", func(w http.ResponseWriter, r *http.Request) {
		profile := r.PathValue("profile")
		mappings := *store.RawMappings.Load()
		m := mappings[profile]

		device, err := mapping.GetDeviceFromID(store.ProductID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		store.BroadcastEvent(mapping.SSEEvent{
			Type: mapping.EventSelectedProfile,
			Data: profile,
		})

		templates.Editor(*m, profile, device).Render(r.Context(), w)
	})

	mux.HandleFunc("PATCH /profiles/{profile}/clear", func(w http.ResponseWriter, r *http.Request) {
		profile := r.PathValue("profile")

		mappings := store.RawMappings.Load()
		m, ok := (*mappings)[profile]
		if !ok {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}

		m.ClearBindings()

		// Recompile FlatMapping
		flat := mapping.CompileFlatMapping(*m)
		maps := store.Mappings.Load()
		if maps != nil {
			(*maps)[profile] = flat
			store.Mappings.Store(maps)
		}

		// Update activeMapping if needed
		if store.ActiveProfile.Load() == profile {
			store.ActiveMapping.Store(flat)
		}

		if err := store.SaveProfile(profile); err != nil {
			http.Error(w, "error saving profile", http.StatusInternalServerError)
			return
		}

		device, err := mapping.GetDeviceFromID(store.ProductID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templates.Editor(*m, profile, device).Render(r.Context(), w)
	})

	mux.HandleFunc("GET /profiles/{profile}/settings", func(w http.ResponseWriter, r *http.Request) {
		profile := r.PathValue("profile")

		mappings := *store.RawMappings.Load()
		windowProfile := mappings[profile].WindowProfile

		templates.SettingsModal(profile, mappings[profile].AxisDeadzone, windowProfile).Render(r.Context(), w)
	})

	mux.HandleFunc("PATCH /profiles/{profile}/settings/update", func(w http.ResponseWriter, r *http.Request) {
		profile := r.PathValue("profile")

		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		namePattern := r.FormValue("nameRegex")
		classPattern := r.FormValue("classRegex")
		deadzoneRaw := r.FormValue("deadzone")
		deadzone, _ := strconv.Atoi(deadzoneRaw)

		mappings := *store.RawMappings.Load()
		m, ok := mappings[profile]
		if !ok {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}

		m.WindowProfile.NamePattern = namePattern
		m.WindowProfile.ClassPattern = classPattern

		m.AxisDeadzone = int16(deadzone)

		path := filepath.Join(store.ProfilePath, profile+".json")
		if err := m.WriteToFile(path); err != nil {
			http.Error(w, "error saving profile", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, ": initial keepalive\n\n")
		flusher.Flush()

		ch := store.Subscribe()
		defer store.Unsubscribe(ch)

		clientGone := r.Context().Done()
		keepalive := time.NewTicker(25 * time.Second)
		defer keepalive.Stop()

		fmt.Fprintf(w, "event: activeProfile\n")
		fmt.Fprintf(w, "data: \"%s\"\n\n", store.ActiveProfile.Load())
		flusher.Flush()

		for {
			select {
			case <-keepalive.C:
				fmt.Fprintf(w, ": keepalive\n\n")
				flusher.Flush()
			case <-clientGone:
				return
			case evt := <-*ch:
				fmt.Fprintf(w, "event: %s\n", evt.Type)
				data, _ := json.Marshal(evt.Data)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: mux,
	}
	defer srv.Shutdown(ctx)

	go func() {
		log.Printf("Web server listening on http://localhost:%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Closing web server")
}
