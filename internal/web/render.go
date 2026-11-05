package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"alat-gudang/internal/store"
	"alat-gudang/internal/timefmt"
)

// Version diisi dari main saat build agar /healthz melaporkan versi yang sama.
var Version = "dev"

type Renderer struct {
	devMode   bool
	templates map[string]*template.Template
	mu        sync.RWMutex
}

func NewRenderer(devMode bool) (*Renderer, error) {
	r := &Renderer{
		devMode:   devMode,
		templates: make(map[string]*template.Template),
	}

	if err := r.parseAll(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Renderer) parseAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	funcMap := template.FuncMap{
		"rel": func(tStr string) string {
			if t, err := timefmt.ParseUTC(tStr); err == nil {
				return timefmt.RelativeID(t, time.Now())
			}
			return tStr
		},
		"clock": func(tStr string) string {
			if t, err := timefmt.ParseUTC(tStr); err == nil {
				return timefmt.FormatClock(t, time.Local)
			}
			return tStr
		},
	}

	pages := []string{"beranda.html", "pinjam.html", "alat_list.html", "alat_detail.html", "kelola.html"}
	partials := []string{"beranda_live.html", "beranda_content.html", "dialog_kembali.html", "toast.html"}

	for _, page := range pages {
		var tpl *template.Template
		var err error

		if r.devMode {
			// Dev mode: baca dari disk
			pattern := filepath.Join("internal", "web", "templates")
			files := []string{
				filepath.Join(pattern, "layout.html"),
				filepath.Join(pattern, "pages", page),
			}
			for _, p := range partials {
				files = append(files, filepath.Join(pattern, "partials", p))
			}
			tpl, err = template.New(page).Funcs(funcMap).ParseFiles(files...)
		} else {
			// Prod: baca dari embed.FS
			tpl = template.New(page).Funcs(funcMap)
			// Parse layout
			layoutBytes, errLayout := EmbedFS.ReadFile("templates/layout.html")
			if errLayout != nil {
				return fmt.Errorf("layout embed gagal: %w", errLayout)
			}
			if _, err = tpl.New("layout.html").Parse(string(layoutBytes)); err != nil {
				return fmt.Errorf("parse layout.html gagal: %w", err)
			}

			// Parse partials
			for _, p := range partials {
				pBytes, pErr := EmbedFS.ReadFile("templates/partials/" + p)
				if pErr == nil {
					if _, err = tpl.New(p).Parse(string(pBytes)); err != nil {
						return fmt.Errorf("parse partial %s gagal: %w", p, err)
					}
				}
			}

			// Parse page directly into root tpl
			pageBytes, pageErr := EmbedFS.ReadFile("templates/pages/" + page)
			if pageErr != nil {
				return fmt.Errorf("page %s embed gagal: %w", page, pageErr)
			}
			if _, err = tpl.Parse(string(pageBytes)); err != nil {
				return fmt.Errorf("parse page %s gagal: %w", page, err)
			}
		}

		if err != nil {
			return fmt.Errorf("parse template %s gagal: %w", page, err)
		}
		r.templates[page] = tpl
	}

	// Parse individual partials for direct fragment rendering
	for _, p := range partials {
		var tpl *template.Template
		var err error

		if r.devMode {
			pPath := filepath.Join("internal", "web", "templates", "partials", p)
			tpl, err = template.New(p).Funcs(funcMap).ParseFiles(pPath)
		} else {
			pBytes, pErr := EmbedFS.ReadFile("templates/partials/" + p)
			if pErr != nil {
				return pErr
			}
			tpl, err = template.New(p).Funcs(funcMap).Parse(string(pBytes))
		}

		if err != nil {
			return fmt.Errorf("parse partial %s gagal: %w", p, err)
		}
		r.templates[p] = tpl
	}

	return nil
}

func (r *Renderer) RenderPage(w io.Writer, req *http.Request, pageName string, data any) error {
	if r.devMode {
		_ = r.parseAll()
	}

	r.mu.RLock()
	tpl, ok := r.templates[pageName]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("template %s tidak terdaftar", pageName)
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	return tpl.ExecuteTemplate(w, "layout.html", data)
}

func (r *Renderer) RenderPartial(w io.Writer, req *http.Request, partialName string, data any) error {
	if r.devMode {
		_ = r.parseAll()
	}

	r.mu.RLock()
	tpl, ok := r.templates[partialName]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("partial %s tidak terdaftar", partialName)
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	return tpl.Execute(w, data)
}

// Flash Toast helper via cookie
const flashCookieName = "ag_flash_toast"

func SetFlashToast(w http.ResponseWriter, toast *store.ToastVM) {
	if toast == nil {
		return
	}
	bytes, err := json.Marshal(toast)
	if err != nil {
		return
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30,
	})
}

func GetFlashToast(w http.ResponseWriter, r *http.Request) *store.ToastVM {
	cookie, err := r.Cookie(flashCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}

	// Delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	decoded, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil
	}

	var toast store.ToastVM
	if err := json.Unmarshal(decoded, &toast); err != nil {
		return nil
	}
	return &toast
}

func StaticFS(devMode bool) http.FileSystem {
	if devMode {
		return http.Dir(filepath.Join("internal", "web", "static"))
	}
	sub, _ := fs.Sub(EmbedFS, "static")
	return http.FS(sub)
}
