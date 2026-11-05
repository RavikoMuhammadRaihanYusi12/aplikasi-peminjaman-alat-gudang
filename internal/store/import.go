package store

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ImportResult merangkum hasil impor CSV.
type ImportResult struct {
	Created int
	Updated int
}

// ImportToolsCSV mengimpor master alat dari file CSV.
// Format kolom: code,name,category,max_hours
// Baris header opsional. Kode yang sudah ada akan diperbarui, bukan diduplikasi.
func (s *Store) ImportToolsCSV(ctx context.Context, r io.Reader) (ImportResult, error) {
	var result ImportResult

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	line := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return result, fmt.Errorf("baca CSV alat baris %d: %w", line+1, err)
		}
		line++
		stripBOM(record)

		if isBlankRecord(record) {
			continue
		}
		if line == 1 && isToolsHeader(record) {
			continue
		}
		if len(record) < 2 {
			return result, fmt.Errorf("baris %d: minimal kolom code dan name", line)
		}

		tool := &Tool{
			Code: record[0],
			Name: record[1],
		}
		if len(record) >= 3 {
			tool.Category = record[2]
		}
		if len(record) >= 4 && strings.TrimSpace(record[3]) != "" {
			maxHours, convErr := strconv.Atoi(strings.TrimSpace(record[3]))
			if convErr != nil {
				return result, fmt.Errorf("baris %d: max_hours bukan angka: %w", line, convErr)
			}
			tool.MaxHours = &maxHours
		}

		var (
			existingID int64
			existing   bool
		)
		err = s.db.QueryRowContext(ctx, "SELECT id FROM tools WHERE code = ?",
			strings.ToUpper(strings.TrimSpace(tool.Code))).Scan(&existingID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// insert baru
		case err != nil:
			return result, err
		default:
			tool.ID = existingID
			existing = true
		}

		if err := s.UpsertTool(ctx, tool); err != nil {
			return result, fmt.Errorf("baris %d: %w", line, err)
		}

		if existing {
			result.Updated++
		} else {
			result.Created++
		}
	}

	return result, nil
}

// ImportTechniciansCSV mengimpor master teknisi dari file CSV.
// Format kolom: name,section
// Teknisi dengan nama yang sama (tanpa membedakan huruf besar/kecil) diperbarui.
func (s *Store) ImportTechniciansCSV(ctx context.Context, r io.Reader) (ImportResult, error) {
	var result ImportResult

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	line := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return result, fmt.Errorf("baca CSV teknisi baris %d: %w", line+1, err)
		}
		line++
		stripBOM(record)

		if isBlankRecord(record) {
			continue
		}
		if line == 1 && isTechniciansHeader(record) {
			continue
		}
		if len(record) < 1 || strings.TrimSpace(record[0]) == "" {
			return result, fmt.Errorf("baris %d: kolom name wajib diisi", line)
		}

		tech := &Technician{Name: record[0]}
		if len(record) >= 2 {
			tech.Section = record[1]
		}

		var (
			existingID int64
			existing   bool
		)
		err = s.db.QueryRowContext(ctx, "SELECT id FROM technicians WHERE LOWER(name) = LOWER(?)",
			strings.TrimSpace(tech.Name)).Scan(&existingID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// insert baru
		case err != nil:
			return result, err
		default:
			tech.ID = existingID
			existing = true
		}

		if err := s.UpsertTechnician(ctx, tech); err != nil {
			return result, fmt.Errorf("baris %d: %w", line, err)
		}

		if existing {
			result.Updated++
		} else {
			result.Created++
		}
	}

	return result, nil
}

// stripBOM menghapus UTF-8 BOM yang sering ditambahkan Excel pada kolom pertama.
func stripBOM(record []string) {
	if len(record) > 0 {
		record[0] = strings.TrimPrefix(record[0], "\ufeff")
	}
}

func isBlankRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

func isToolsHeader(record []string) bool {
	if len(record) == 0 {
		return false
	}
	first := strings.ToLower(strings.TrimSpace(record[0]))
	return first == "code" || first == "kode"
}

func isTechniciansHeader(record []string) bool {
	if len(record) == 0 {
		return false
	}
	first := strings.ToLower(strings.TrimSpace(record[0]))
	return first == "name" || first == "nama"
}
