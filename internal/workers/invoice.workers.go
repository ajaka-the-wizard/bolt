package workers

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ajaka-the-wizard/bolt/internal/domain"
	"github.com/ajaka-the-wizard/bolt/internal/models"
	"github.com/ajaka-the-wizard/bolt/internal/store"
	"github.com/signintech/gopdf"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

type invoiceWorkers struct {
	store *store.Store
}

func (i *invoiceWorkers) GenerateInvoice(ctx context.Context, id string) {
	logger := slog.Default().With(slog.String("group", domain.BoltRedisInvoiceConsumerGroup), slog.String("id", id))
	go func() {
		t := time.NewTicker(time.Second * 2)
		defer t.Stop()

		for range t.C {
			data, err := i.store.FetchNextTask(ctx, id, domain.BoltRedisInvoiceStreamKey, domain.BoltRedisInvoiceConsumerGroup, logger)
			if err == ctx.Err() {
				logger.Error("Cancellation error")
				break
			}
			logger.Info("Processing job", slog.Any("job", data))
			// TODO implement actual work
			time.Sleep(time.Minute)
		}

	}()
}

func InitInvoiceWorkers(ctx context.Context, store *store.Store, logger *slog.Logger, company *domain.CompanyInfo) {
	s := invoiceWorkers{
		store: store,
	}
	s.GenerateInvoice(ctx, "worker1")
	s.GenerateInvoice(ctx, "worker2")
	s.GenerateInvoice(ctx, "worker3")
	logger.Info("Invoice generating workers initialized")
}

// GenerateInvoicePDF renders a PDF invoice for the given order and writes it
// to outputPath (e.g. "invoice_ORD-001.pdf"). Fonts are embedded — no external
// font files required.
func GenerateInvoicePDF(order models.Order, company *domain.CompanyInfo, outputPath string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	// Load embedded Go fonts (clean, sans-serif — swap TTF paths here if desired)
	if err := pdf.AddTTFFontByReader(fontRegular, bytes.NewReader(goregular.TTF)); err != nil {
		return fmt.Errorf("load regular font: %w", err)
	}
	if err := pdf.AddTTFFontByReader(fontBold, bytes.NewReader(gobold.TTF)); err != nil {
		return fmt.Errorf("load bold font: %w", err)
	}

	y := 40.0

	y = drawHeader(&pdf, order, company, y)
	y = divider(&pdf, y, 8, 1.5)
	y = drawAddressBlock(&pdf, order, y)
	y = divider(&pdf, y, 10, 0.5)
	y = drawItemsTable(&pdf, order, y)
	y = divider(&pdf, y, 8, 0.5)
	y = drawTotals(&pdf, order, y)
	y = divider(&pdf, y, 14, 0.5)
	drawPaymentSection(&pdf, order, y)
	drawFooter(&pdf, company)

	return pdf.WritePdf(outputPath)
}

// ─── Section renderers ────────────────────────────────────────────────────────

func drawHeader(pdf *gopdf.GoPdf, order models.Order, company *domain.CompanyInfo, y float64) float64 {
	// Left: company name
	setFont(pdf, fontBold, 18)
	put(pdf, mLeft, y, company.Name)
	y += 26

	// Left: company contact block
	setFont(pdf, fontRegular, 9)
	for _, line := range companyLines(company) {
		put(pdf, mLeft, y, line)
		y += 13
	}

	// Right: "INVOICE" title
	ry := 40.0
	setFont(pdf, fontBold, 24)
	put(pdf, rightX, ry, "INVOICE")
	ry += 32

	// Right: invoice meta rows
	meta := [][2]string{
		{"Invoice #:", order.OrderNumber},
		{"Date:", order.CreatedAt.Format("Jan 02, 2006")},
		{"Order ID:", order.ID},
	}
	for _, row := range meta {
		setFont(pdf, fontBold, 9)
		put(pdf, rightX, ry, row[0])
		setFont(pdf, fontRegular, 9)
		put(pdf, rightX+72, ry, row[1])
		ry += 14
	}

	if y < ry {
		return ry
	}
	return y
}

func drawAddressBlock(pdf *gopdf.GoPdf, order models.Order, y float64) float64 {
	setFont(pdf, fontBold, 9)
	put(pdf, mLeft, y, "BILL TO / SHIP TO")
	y += 14

	addr := order.ShippingAddress
	lines := []string{order.CustomerName, order.CustomerEmail, addr.AddressLine1}
	if addr.AddressLine2 != "" {
		lines = append(lines, addr.AddressLine2)
	}
	lines = append(lines,
		fmt.Sprintf("%s, %s %s", addr.City, addr.State, addr.PostalCode),
		addr.Country,
	)

	setFont(pdf, fontRegular, 9)
	for _, l := range lines {
		put(pdf, mLeft, y, l)
		y += 13
	}
	return y
}

func drawItemsTable(pdf *gopdf.GoPdf, order models.Order, y float64) float64 {
	rowH := 18.0

	// Header row — dark background, white text
	pdf.SetFillColor(45, 45, 45)
	pdf.Rectangle(mLeft, y, colEnd, y+rowH, "F", 0, 0)
	pdf.SetTextColor(255, 255, 255)
	setFont(pdf, fontBold, 9)
	put(pdf, colItem+6, y+5, "ITEM")
	put(pdf, colQty+6, y+5, "QTY")
	put(pdf, colUnit+6, y+5, "UNIT PRICE")
	put(pdf, colTotal+6, y+5, "TOTAL")
	pdf.SetTextColor(0, 0, 0)
	y += rowH

	// Item rows — alternating shading
	setFont(pdf, fontRegular, 9)
	for i, item := range order.Items {
		if i%2 == 0 {
			pdf.SetFillColor(248, 248, 248)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.Rectangle(mLeft, y, colEnd, y+rowH, "F", 0, 0)

		put(pdf, colItem+6, y+5, item.Name)
		put(pdf, colQty+6, y+5, fmt.Sprintf("%d", item.Quantity))
		put(pdf, colUnit+6, y+5, money(order.Currency, item.UnitPrice))
		put(pdf, colTotal+6, y+5, money(order.Currency, item.TotalPrice))
		y += rowH
	}

	pdf.SetFillColor(0, 0, 0) // reset fill
	return y
}

func drawTotals(pdf *gopdf.GoPdf, order models.Order, y float64) float64 {
	labelX := rightX
	valueX := 480.0
	rowH := 16.0

	rows := [][2]string{
		{"Subtotal:", money(order.Currency, order.Subtotal)},
		{"Shipping:", money(order.Currency, order.ShippingCost)},
		{"Tax:", money(order.Currency, order.Tax)},
	}
	if order.Discount > 0 {
		rows = append(rows, [2]string{"Discount:", "- " + money(order.Currency, order.Discount)})
	}

	for _, row := range rows {
		setFont(pdf, fontBold, 9) // label bold
		put(pdf, labelX, y, row[0])
		setFont(pdf, fontRegular, 9) // value regular
		put(pdf, valueX, y, row[1])
		y += rowH
	}

	// Grand total — dark banner
	y += 6
	bannerH := 22.0
	pdf.SetFillColor(30, 30, 30)
	pdf.Rectangle(labelX-6, y-4, colEnd, y-4+bannerH, "F", 0, 0)
	pdf.SetTextColor(255, 255, 255)
	setFont(pdf, fontBold, 11)
	put(pdf, labelX, y+3, "TOTAL:")
	put(pdf, valueX, y+3, money(order.Currency, order.Total))
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(0, 0, 0) // reset fill

	return y + bannerH
}

func drawPaymentSection(pdf *gopdf.GoPdf, order models.Order, y float64) {
	setFont(pdf, fontBold, 9)
	put(pdf, mLeft, y, "PAYMENT METHOD")
	setFont(pdf, fontRegular, 9)
	put(pdf, mLeft, y+14, order.PaymentMethod)
}

func drawFooter(pdf *gopdf.GoPdf, company *domain.CompanyInfo) {
	y := pageH - 50.0
	hline(pdf, y, 0.5)
	y += 10

	pdf.SetTextColor(130, 130, 130)
	setFont(pdf, fontRegular, 8)
	put(pdf, mLeft, y, "Thank you for your business! Questions? Contact "+company.Email)
	put(pdf, mLeft, y+12, fmt.Sprintf("%s  ·  %s  ·  Tax ID: %s", company.Phone, company.Website, company.TaxID))
	pdf.SetTextColor(0, 0, 0)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// divider draws a horizontal rule, applying padding above and below it,
// and returns the new Y position.
func divider(pdf *gopdf.GoPdf, y, padding, thickness float64) float64 {
	y += padding
	hline(pdf, y, thickness)
	return y + padding
}

func setFont(pdf *gopdf.GoPdf, name string, size float64) {
	_ = pdf.SetFont(name, "", size)
}

func put(pdf *gopdf.GoPdf, x, y float64, text string) {
	pdf.SetX(x)
	pdf.SetY(y)
	_ = pdf.Cell(nil, text)
}

func hline(pdf *gopdf.GoPdf, y, thickness float64) {
	pdf.SetLineWidth(thickness)
	pdf.Line(mLeft, y, pageW-mRight, y)
}

func money(currency string, amount float64) string {
	return fmt.Sprintf("%s%.2f", currency, amount)
}

func companyLines(company *domain.CompanyInfo) []string {
	lines := []string{company.AddressLine1}
	if company.AddressLine2 != "" {
		lines = append(lines, company.AddressLine2)
	}
	return append(lines,
		fmt.Sprintf("%s, %s %s", company.City, company.State, company.PostalCode),
		company.Country,
		company.Phone,
		company.Email,
		company.Website,
		"Tax ID: "+company.TaxID,
	)
}
