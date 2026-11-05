package store

import (
	"database/sql"
	"errors"
	"fmt"
)

type Condition string

const (
	Baik  Condition = "baik"
	Rusak Condition = "rusak"
)

var (
	ErrNotFound        = errors.New("data tidak ditemukan")
	ErrInvalid         = errors.New("input tidak valid")
	ErrCannotUndo      = errors.New("sudah ada pencatatan lain setelahnya")
	ErrAlreadyReturned = errors.New("sudah tercatat dikembalikan")
	ErrDuplicateCode   = errors.New("kode alat sudah digunakan")
)

type UnavailableError struct {
	ToolID     int64
	Code       string
	Reason     string // "dipinjam" | "rusak" | "nonaktif"
	HolderName string // terisi bila Reason == "dipinjam"
}

func (e *UnavailableError) Error() string {
	if e.Reason == "dipinjam" {
		return fmt.Sprintf("Alat %s sedang dipinjam oleh %s", e.Code, e.HolderName)
	}
	if e.Reason == "rusak" {
		return fmt.Sprintf("Alat %s sedang berstatus rusak", e.Code)
	}
	return fmt.Sprintf("Alat %s sedang tidak aktif", e.Code)
}

type Technician struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Section   string `json:"section"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

type Tool struct {
	ID            int64     `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Category      string    `json:"category"`
	MaxHours      *int      `json:"max_hours"`
	Condition     Condition `json:"condition"`
	ConditionNote string    `json:"condition_note"`
	Active        bool      `json:"active"`
	CreatedAt     string    `json:"created_at"`
}

type Loan struct {
	ID              int64          `json:"id"`
	ToolID          int64          `json:"tool_id"`
	TechnicianID    int64          `json:"technician_id"`
	BorrowedAt      string         `json:"borrowed_at"`
	BorrowShift     string         `json:"borrow_shift"`
	ReturnedAt      sql.NullString `json:"returned_at"`
	ReturnShift     sql.NullString `json:"return_shift"`
	ReturnCondition sql.NullString `json:"return_condition"`
	ReturnNote      string         `json:"return_note"`
	VoidedAt        sql.NullString `json:"voided_at"`
}

type Settings struct {
	DefaultMaxHours int    `json:"default_max_hours"`
	ShiftPagiStart  string `json:"shift_pagi_start"`
	ShiftMalamStart string `json:"shift_malam_start"`
}

// View Models for HTML rendering
type DashboardItemVM struct {
	LoanID              int64  `json:"loan_id"`
	ToolID              int64  `json:"tool_id"`
	ToolCode            string `json:"tool_code"`
	ToolName            string `json:"tool_name"`
	TechnicianID        int64  `json:"technician_id"`
	TechnicianName      string `json:"technician_name"`
	TechnicianSection   string `json:"technician_section"`
	BorrowedAt          string `json:"borrowed_at"`
	BorrowShift         string `json:"borrow_shift"`
	BorrowedAtFormatted string `json:"borrowed_at_formatted"`
	DurationRelative    string `json:"duration_relative"`
	IsOverdue           bool   `json:"is_overdue"`
}

type DamagedToolVM struct {
	ID                  int64  `json:"id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	ConditionNote       string `json:"condition_note"`
	LastBorrowerName    string `json:"last_borrower_name"`
	ReturnedAt          string `json:"returned_at"`
	ReturnedAtFormatted string `json:"returned_at_formatted"`
}

type DashboardStatsVM struct {
	BorrowedCount int `json:"borrowed_count"`
	OverdueCount  int `json:"overdue_count"`
	DamagedCount  int `json:"damaged_count"`
}

type DashboardVM struct {
	ActiveNav       string            `json:"active_nav"`
	Stats           DashboardStatsVM  `json:"stats"`
	OverdueLoans    []DashboardItemVM `json:"overdue_loans"`
	ActiveLoans     []DashboardItemVM `json:"active_loans"`
	DamagedTools    []DamagedToolVM   `json:"damaged_tools"`
	Query           string            `json:"query"`
	DefaultMaxHours int               `json:"default_max_hours"`
	Toast           *ToastVM          `json:"toast"`
}

type ToolRowVM struct {
	ID               int64     `json:"id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	MaxHours         *int      `json:"max_hours"`
	Condition        Condition `json:"condition"`
	ConditionNote    string    `json:"condition_note"`
	Active           bool      `json:"active"`
	Status           string    `json:"status"` // "tersedia" | "dipinjam" | "rusak"
	CurrentBorrower  string    `json:"current_borrower"`
	LastBorrowerName string    `json:"last_borrower_name"`
}

type ToolHistoryItemVM struct {
	ID                  int64  `json:"id"`
	TechnicianName      string `json:"technician_name"`
	TechnicianSection   string `json:"technician_section"`
	BorrowedAtFormatted string `json:"borrowed_at_formatted"`
	BorrowShift         string `json:"borrow_shift"`
	ReturnedAtFormatted string `json:"returned_at_formatted"`
	ReturnShift         string `json:"return_shift"`
	ReturnCondition     string `json:"return_condition"`
	ReturnNote          string `json:"return_note"`
}

type ToolDetailVM struct {
	ActiveNav string              `json:"active_nav"`
	Tool      ToolRowVM           `json:"tool"`
	LastLoan  *ToolHistoryItemVM  `json:"last_loan"`
	History   []ToolHistoryItemVM `json:"history"`
	Toast     *ToastVM            `json:"toast"`
}

type ToastVM struct {
	Kind        string  `json:"kind"` // "success" | "warn" | "danger"
	Message     string  `json:"message"`
	Subtext     string  `json:"subtext"`
	UndoLoanIDs []int64 `json:"undo_loan_ids"`
	UndoKind    string  `json:"undo_kind"` // "pinjam" | "kembali"
}

// ReturnDialogVM adalah view model untuk dialog pengembalian.
// Dibuat di layer store agar handler tidak perlu menulis SQL sendiri.
type ReturnDialogVM struct {
	LoanID              int64
	ToolID              int64
	ToolCode            string
	ToolName            string
	TechnicianName      string
	BorrowedAtFormatted string
}
