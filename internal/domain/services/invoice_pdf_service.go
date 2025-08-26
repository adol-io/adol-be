package services

import (
	"context"
	"io"

	"github.com/nicklaros/adol/internal/domain/entities"
)

// InvoicePDFService defines the interface for invoice PDF generation
type InvoicePDFService interface {
	// GenerateInvoicePDF generates a PDF invoice
	GenerateInvoicePDF(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate) ([]byte, error)

	// GenerateInvoicePDFToWriter generates a PDF invoice and writes to a writer
	GenerateInvoicePDFToWriter(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate, writer io.Writer) error

	// GenerateReceiptPDF generates a thermal receipt PDF (80mm width)
	GenerateReceiptPDF(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate) ([]byte, error)

	// ValidateTemplate validates an invoice template
	ValidateTemplate(template *entities.InvoiceTemplate) error

	// GetDefaultTemplate returns the default invoice template for a paper size
	GetDefaultTemplate(paperSize entities.PaperSize) *entities.InvoiceTemplate

	// PreviewInvoice generates a preview image of the invoice
	PreviewInvoice(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate) ([]byte, error)
}

// EmailService defines the interface for email operations
type EmailService interface {
	// SendInvoiceEmail sends an invoice via email
	SendInvoiceEmail(ctx context.Context, invoice *entities.Invoice, recipient string, pdfData []byte) error

	// SendReceiptEmail sends a receipt via email
	SendReceiptEmail(ctx context.Context, invoice *entities.Invoice, recipient string, pdfData []byte) error

	// SendPaymentConfirmation sends payment confirmation email
	SendPaymentConfirmation(ctx context.Context, invoice *entities.Invoice, recipient string) error

	// SendInvoiceReminder sends payment reminder email
	SendInvoiceReminder(ctx context.Context, invoice *entities.Invoice, recipient string) error

	// SendOverdueNotice sends overdue payment notice
	SendOverdueNotice(ctx context.Context, invoice *entities.Invoice, recipient string) error

	// ValidateEmailAddress validates an email address
	ValidateEmailAddress(email string) bool
}

// PrintService defines the interface for printing operations
type PrintService interface {
	// PrintInvoice prints an invoice to a printer
	PrintInvoice(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate, printerName string) error

	// PrintReceipt prints a receipt to a thermal printer
	PrintReceipt(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate, printerName string) error

	// GetAvailablePrinters returns list of available printers
	GetAvailablePrinters() ([]PrinterInfo, error)

	// GetDefaultPrinter returns the default printer
	GetDefaultPrinter() (*PrinterInfo, error)

	// SetDefaultPrinter sets the default printer
	SetDefaultPrinter(printerName string) error
}

// PrinterInfo represents printer information
type PrinterInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	IsDefault   bool   `json:"is_default"`
	SupportsA4  bool   `json:"supports_a4"`
	SupportsA5  bool   `json:"supports_a5"`
	Isthermal   bool   `json:"is_thermal"`
}

// PDFOptions represents options for PDF generation
type PDFOptions struct {
	PaperSize   entities.PaperSize `json:"paper_size"`
	Orientation string             `json:"orientation"` // "portrait" or "landscape"
	Margins     PDFMargins         `json:"margins"`
	Font        PDFFont            `json:"font"`
	Colors      PDFColors          `json:"colors"`
	ShowLogo    bool               `json:"show_logo"`
	LogoPath    string             `json:"logo_path,omitempty"`
	Watermark   string             `json:"watermark,omitempty"`
}

// PDFMargins represents PDF margin settings
type PDFMargins struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// PDFFont represents PDF font settings
type PDFFont struct {
	Family string  `json:"family"`
	Size   float64 `json:"size"`
	Style  string  `json:"style"` // "normal", "bold", "italic"
}

// PDFColors represents PDF color scheme
type PDFColors struct {
	Primary    string `json:"primary"`    // Primary color (hex)
	Secondary  string `json:"secondary"`  // Secondary color (hex)
	Text       string `json:"text"`       // Text color (hex)
	Background string `json:"background"` // Background color (hex)
}

// EmailTemplate represents email template
type EmailTemplate struct {
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	IsHTML      bool              `json:"is_html"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
}

// EmailAttachment represents email attachment
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

// DefaultPDFOptions returns default PDF options for different paper sizes
func GetDefaultPDFOptions(paperSize entities.PaperSize) PDFOptions {
	switch paperSize {
	case entities.PaperSizeA4:
		return PDFOptions{
			PaperSize:   entities.PaperSizeA4,
			Orientation: "portrait",
			Margins: PDFMargins{
				Top:    72, // 1 inch
				Bottom: 72, // 1 inch
				Left:   72, // 1 inch
				Right:  72, // 1 inch
			},
			Font: PDFFont{
				Family: "Arial",
				Size:   10,
				Style:  "normal",
			},
			Colors: PDFColors{
				Primary:    "#2E86AB",
				Secondary:  "#A23B72",
				Text:       "#333333",
				Background: "#FFFFFF",
			},
			ShowLogo: true,
		}
	case entities.PaperSizeA5:
		return PDFOptions{
			PaperSize:   entities.PaperSizeA5,
			Orientation: "portrait",
			Margins: PDFMargins{
				Top:    36, // 0.5 inch
				Bottom: 36, // 0.5 inch
				Left:   36, // 0.5 inch
				Right:  36, // 0.5 inch
			},
			Font: PDFFont{
				Family: "Arial",
				Size:   8,
				Style:  "normal",
			},
			Colors: PDFColors{
				Primary:    "#2E86AB",
				Secondary:  "#A23B72",
				Text:       "#333333",
				Background: "#FFFFFF",
			},
			ShowLogo: true,
		}
	case entities.PaperSizeReceipt:
		return PDFOptions{
			PaperSize:   entities.PaperSizeReceipt,
			Orientation: "portrait",
			Margins: PDFMargins{
				Top:    5, // Minimal margin for receipt
				Bottom: 5, // Minimal margin for receipt
				Left:   5, // Minimal margin for receipt
				Right:  5, // Minimal margin for receipt
			},
			Font: PDFFont{
				Family: "Courier",
				Size:   8,
				Style:  "normal",
			},
			Colors: PDFColors{
				Primary:    "#000000",
				Secondary:  "#000000",
				Text:       "#000000",
				Background: "#FFFFFF",
			},
			ShowLogo: false, // Usually no logo on thermal receipts
		}
	default:
		return GetDefaultPDFOptions(entities.PaperSizeA4)
	}
}
