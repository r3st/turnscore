package service

import (
	"bytes"
	"context"
	"fmt"
	"image/png"

	"github.com/google/uuid"
	gofpdf "github.com/jung-kurt/gofpdf"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/storage"
)

const (
	qrCodeSize = 256

	pdfCardW      = 63.0  // mm
	pdfCardH      = 90.0  // mm
	pdfCols       = 3
	pdfRows       = 2
	pdfMarginLeft = 10.5  // mm
	pdfMarginTop  = 10.0  // mm
	pdfCardGapX   = 0.0   // no gap — cards are flush
	pdfCardGapY   = 0.0

	qrStorageKeyFmt = "tables/%s/qrcode.png"
)

// QRCodeService generates QR codes and PDF exports for tournament tables.
type QRCodeService struct {
	tableRepo   domain.TableRepository
	memberRepo  domain.TournamentMemberRepository
	tourneyRepo domain.TournamentRepository
	storage     storage.Storage
	frontendURL string
}

// NewQRCodeService creates a new QRCodeService.
func NewQRCodeService(
	tableRepo domain.TableRepository,
	memberRepo domain.TournamentMemberRepository,
	tourneyRepo domain.TournamentRepository,
	stor storage.Storage,
	frontendURL string,
) *QRCodeService {
	return &QRCodeService{
		tableRepo:   tableRepo,
		memberRepo:  memberRepo,
		tourneyRepo: tourneyRepo,
		storage:     stor,
		frontendURL: frontendURL,
	}
}

// GenerateForTable generates a QR code PNG for a single table, stores it, and returns the PNG bytes.
// Caller must be organizer or helper.
func (s *QRCodeService) GenerateForTable(ctx context.Context, userID uuid.UUID, slug string, number int) ([]byte, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("GenerateForTable find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, domain.ErrForbidden
	}

	table, err := s.tableRepo.FindByTournamentAndNumber(ctx, t.ID, number)
	if err != nil {
		return nil, fmt.Errorf("GenerateForTable find table: %w", err)
	}

	qrURL := fmt.Sprintf("%s/rate/%s/%d", s.frontendURL, slug, number)

	pngBytes, err := qrcode.Encode(qrURL, qrcode.Medium, qrCodeSize)
	if err != nil {
		return nil, fmt.Errorf("GenerateForTable encode QR: %w", err)
	}

	storageKey := fmt.Sprintf(qrStorageKeyFmt, table.ID)
	if putErr := s.storage.Put(ctx, storageKey, bytes.NewReader(pngBytes), int64(len(pngBytes)), "image/png"); putErr != nil {
		return nil, fmt.Errorf("GenerateForTable store QR: %w", putErr)
	}

	if updErr := s.tableRepo.UpdateQRCodePath(ctx, table.ID, storageKey); updErr != nil {
		return nil, fmt.Errorf("GenerateForTable update path: %w", updErr)
	}

	return pngBytes, nil
}

// ExportPDF generates an A4 PDF with QR cards (3 per row, 6 per page) for all tables.
// Caller must be organizer or helper.
func (s *QRCodeService) ExportPDF(ctx context.Context, userID uuid.UUID, slug string) ([]byte, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("ExportPDF find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, domain.ErrForbidden
	}

	tables, err := s.tableRepo.ListByTournamentID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("ExportPDF list tables: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMarginLeft, pdfMarginTop, pdfMarginLeft)
	pdf.AddPage()

	cardsPerPage := pdfCols * pdfRows
	col := 0
	row := 0

	for i, table := range tables {
		if i > 0 && i%cardsPerPage == 0 {
			pdf.AddPage()
			col = 0
			row = 0
		}

		x := pdfMarginLeft + float64(col)*(pdfCardW+pdfCardGapX)
		y := pdfMarginTop + float64(row)*(pdfCardH+pdfCardGapY)

		qrURL := fmt.Sprintf("%s/rate/%s/%d", s.frontendURL, slug, table.Number)
		pngBytes, encErr := qrcode.Encode(qrURL, qrcode.Medium, qrCodeSize)
		if encErr != nil {
			return nil, fmt.Errorf("ExportPDF encode QR table %d: %w", table.Number, encErr)
		}

		// Embed QR image via PNG reader registered with gofpdf.
		imgName := fmt.Sprintf("qr_%d", table.Number)
		imgOpts := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
		pdf.RegisterImageOptionsReader(imgName, imgOpts, bytes.NewReader(pngBytes))

		qrSize := pdfCardW - 10.0 // leave 5mm margin each side
		pdf.ImageOptions(imgName, x+5.0, y+5.0, qrSize, qrSize, false, imgOpts, 0, "")

		// Table number.
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetXY(x, y+pdfCardW)
		pdf.CellFormat(pdfCardW, 8, fmt.Sprintf("Table %d", table.Number), "", 0, "C", false, 0, "")

		// Table name (if set).
		if table.Name != nil {
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetXY(x, y+pdfCardW+8)
			pdf.CellFormat(pdfCardW, 6, *table.Name, "", 0, "C", false, 0, "")
		}

		// Draw card border.
		pdf.SetDrawColor(180, 180, 180)
		pdf.Rect(x, y, pdfCardW, pdfCardH, "D")

		col++
		if col >= pdfCols {
			col = 0
			row++
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("ExportPDF output: %w", err)
	}
	return buf.Bytes(), nil
}

// buildRatersPDF generates a simple A4 PDF table listing rater nicknames and codes.
func buildRatersPDF(tournamentName string, slug string, raters []domain.Rater) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, tournamentName+" - Teilnehmerliste", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, "Slug: "+slug, "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// Header row
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(95, 8, "Spitzname / Nickname", "1", 0, "L", true, 0, "")
	pdf.CellFormat(45, 8, "Code", "1", 0, "C", true, 0, "")
	pdf.CellFormat(45, 8, "Turnier-Slug", "1", 1, "L", true, 0, "")

	// Data rows
	pdf.SetFont("Helvetica", "", 11)
	fill := false
	pdf.SetFillColor(245, 245, 245)
	for _, r := range raters {
		pdf.CellFormat(95, 8, r.Nickname, "1", 0, "L", fill, 0, "")
		pdf.CellFormat(45, 8, r.Code, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(45, 8, slug, "1", 1, "L", fill, 0, "")
		fill = !fill
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("buildRatersPDF output: %w", err)
	}
	return buf.Bytes(), nil
}

// pngBytesToImage decodes PNG bytes — helper used indirectly via registered readers.
func pngBytesToImage(data []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
