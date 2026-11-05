package web

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"alat-gudang/internal/store"
)

type KelolaPageVM struct {
	ActiveNav   string             `json:"active_nav"`
	SubPage     string             `json:"sub_page"` // "alat" | "teknisi" | "pengaturan"
	Tools       []store.ToolRowVM  `json:"tools"`
	Technicians []store.Technician `json:"technicians"`
	Settings    store.Settings     `json:"settings"`
	Toast       *store.ToastVM     `json:"toast"`
	// EditTool/EditTech terisi bila halaman dibuka dengan ?edit=ID.
	// Form yang sama dipakai untuk tambah dan ubah.
	EditTool *store.ToolRowVM  `json:"edit_tool"`
	EditTech *store.Technician `json:"edit_tech"`
}

func (s *Server) kelola(w http.ResponseWriter, r *http.Request) {
	bagian := r.PathValue("bagian")
	switch bagian {
	case "alat", "teknisi", "pengaturan":
		// valid
	default:
		http.Error(w, "Halaman Kelola tidak ditemukan", http.StatusNotFound)
		return
	}

	tools, err := s.store.ListAllTools(r.Context())
	if err != nil {
		http.Error(w, "Gagal memuat daftar alat", http.StatusInternalServerError)
		return
	}

	techs, err := s.store.ListTechnicians(r.Context())
	if err != nil {
		http.Error(w, "Gagal memuat daftar teknisi", http.StatusInternalServerError)
		return
	}

	settings, err := s.store.GetSettings(r.Context())
	if err != nil {
		http.Error(w, "Gagal memuat pengaturan", http.StatusInternalServerError)
		return
	}

	vm := KelolaPageVM{
		ActiveNav:   "kelola",
		SubPage:     bagian,
		Tools:       tools,
		Technicians: techs,
		Settings:    settings,
		Toast:       GetFlashToast(w, r),
	}

	// ?edit=ID membuka form ubah tanpa menambah route baru.
	if id, ok := parsePositiveID(r.URL.Query().Get("edit")); ok {
		switch bagian {
		case "alat":
			for i := range tools {
				if tools[i].ID == id {
					vm.EditTool = &tools[i]
					break
				}
			}
		case "teknisi":
			for i := range techs {
				if techs[i].ID == id {
					vm.EditTech = &techs[i]
					break
				}
			}
		}
	}

	_ = s.renderer.RenderPage(w, r, "kelola.html", vm)
}

func (s *Server) simpanAlat(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "Form tidak valid. Coba lagi."}, "/kelola/alat")
		return
	}

	var (
		id       int64
		idRaw    = strings.TrimSpace(r.FormValue("id"))
		maxHours *int
	)
	if idRaw != "" {
		parsed, ok := parsePositiveID(idRaw)
		if !ok {
			flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "ID alat tidak valid."}, "/kelola/alat")
			return
		}
		id = parsed
	}

	if raw := strings.TrimSpace(r.FormValue("max_hours")); raw != "" {
		val, err := strconv.Atoi(raw)
		if err != nil || val <= 0 {
			flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "Batas waktu harus berupa angka lebih dari 0 jam."}, "/kelola/alat")
			return
		}
		maxHours = &val
	}

	tool := &store.Tool{
		ID:       id,
		Code:     r.FormValue("code"),
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
		MaxHours: maxHours,
	}

	if err := s.store.UpsertTool(r.Context(), tool); err != nil {
		var message string
		switch {
		case errors.Is(err, store.ErrDuplicateCode):
			message = fmt.Sprintf("Kode alat %s sudah digunakan. Gunakan kode lain.", strings.ToUpper(strings.TrimSpace(tool.Code)))
		case errors.Is(err, store.ErrNotFound):
			message = "Alat yang ingin diubah tidak ditemukan."
		case errors.Is(err, store.ErrInvalid):
			message = "Kode dan nama alat wajib diisi (batas waktu 1-8760 jam bila diisi)."
		default:
			message = "Ada gangguan saat menyimpan alat. Coba lagi atau hubungi IT."
		}
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: message}, "/kelola/alat")
		return
	}

	action := "ditambahkan"
	if id > 0 {
		action = "diperbarui"
	}
	flashRedirect(w, r, &store.ToastVM{
		Kind:    "success",
		Message: fmt.Sprintf("Alat %s (%s) berhasil %s.", strings.TrimSpace(tool.Name), tool.Code, action),
	}, "/kelola/alat")
}

func (s *Server) simpanTeknisi(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "Form tidak valid. Coba lagi."}, "/kelola/teknisi")
		return
	}

	var id int64
	if raw := strings.TrimSpace(r.FormValue("id")); raw != "" {
		parsed, ok := parsePositiveID(raw)
		if !ok {
			flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "ID teknisi tidak valid."}, "/kelola/teknisi")
			return
		}
		id = parsed
	}

	tech := &store.Technician{
		ID:      id,
		Name:    r.FormValue("name"),
		Section: r.FormValue("section"),
	}

	if err := s.store.UpsertTechnician(r.Context(), tech); err != nil {
		var message string
		switch {
		case errors.Is(err, store.ErrNotFound):
			message = "Teknisi yang ingin diubah tidak ditemukan."
		case errors.Is(err, store.ErrInvalid):
			message = "Nama teknisi wajib diisi."
		default:
			message = "Ada gangguan saat menyimpan teknisi. Coba lagi atau hubungi IT."
		}
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: message}, "/kelola/teknisi")
		return
	}

	action := "ditambahkan"
	if id > 0 {
		action = "diperbarui"
	}
	flashRedirect(w, r, &store.ToastVM{
		Kind:    "success",
		Message: fmt.Sprintf("Teknisi %s (%s) berhasil %s.", strings.TrimSpace(tech.Name), strings.TrimSpace(tech.Section), action),
	}, "/kelola/teknisi")
}

func (s *Server) ubahAktif(w http.ResponseWriter, r *http.Request) {
	jenis := r.PathValue("jenis")
	if jenis != "alat" && jenis != "teknisi" {
		http.Error(w, "Jenis data tidak valid", http.StatusBadRequest)
		return
	}

	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "ID tidak valid."}, "/kelola/"+jenis)
		return
	}

	if err := r.ParseForm(); err != nil {
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "Form tidak valid. Coba lagi."}, "/kelola/"+jenis)
		return
	}
	activeVal := r.FormValue("active") == "1"

	var err error
	if jenis == "alat" {
		err = s.store.SetToolActive(r.Context(), id, activeVal)
	} else {
		err = s.store.SetTechnicianActive(r.Context(), id, activeVal)
	}

	if err != nil {
		message := "Ada gangguan saat mengubah status. Coba lagi atau hubungi IT."
		if errors.Is(err, store.ErrNotFound) {
			message = "Data tidak ditemukan. Muat ulang halaman."
		}
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: message}, "/kelola/"+jenis)
		return
	}

	state := "dinonaktifkan"
	if activeVal {
		state = "diaktifkan"
	}
	flashRedirect(w, r, &store.ToastVM{
		Kind:    "warn",
		Message: fmt.Sprintf("Data berhasil %s.", state),
	}, "/kelola/"+jenis)
}

func (s *Server) simpanPengaturan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "Form tidak valid. Coba lagi."}, "/kelola/pengaturan")
		return
	}

	defaultHours, err := strconv.Atoi(strings.TrimSpace(r.FormValue("default_max_hours")))
	if err != nil {
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: "Batas waktu default harus berupa angka."}, "/kelola/pengaturan")
		return
	}

	settings := store.Settings{
		DefaultMaxHours: defaultHours,
		ShiftPagiStart:  r.FormValue("shift_pagi_start"),
		ShiftMalamStart: r.FormValue("shift_malam_start"),
	}

	if err := s.store.UpdateSettings(r.Context(), settings); err != nil {
		message := "Ada gangguan saat menyimpan pengaturan. Coba lagi atau hubungi IT."
		if errors.Is(err, store.ErrInvalid) {
			message = "Pengaturan tidak valid. Pastikan batas 1-8760 jam dan jam shift berformat HH:MM serta berbeda."
		}
		flashRedirect(w, r, &store.ToastVM{Kind: "danger", Message: message}, "/kelola/pengaturan")
		return
	}

	flashRedirect(w, r, &store.ToastVM{
		Kind:    "success",
		Message: "Pengaturan sistem berhasil disimpan.",
	}, "/kelola/pengaturan")
}
