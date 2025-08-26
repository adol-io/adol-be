package services

import (
	"context"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/services"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
)

// PrintService implements the domain PrintService interface
type PrintService struct {
	pdfService     services.InvoicePDFService
	defaultPrinter string
	logger         logger.Logger
}

// NewPrintService creates a new print service
func NewPrintService(pdfService services.InvoicePDFService, defaultPrinter string, logger logger.Logger) services.PrintService {
	return &PrintService{
		pdfService:     pdfService,
		defaultPrinter: defaultPrinter,
		logger:         logger,
	}
}

// PrintInvoice prints an invoice to a printer
func (s *PrintService) PrintInvoice(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate, printerName string) error {
	if invoice == nil {
		return errors.NewValidationError("invoice is required", "invoice cannot be nil")
	}
	if template == nil {
		return errors.NewValidationError("template is required", "template cannot be nil")
	}

	// TODO: Implement actual printing functionality
	s.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"printer_name":   printerName,
	}).Info("Invoice print request received (not implemented)")

	return nil
}

// PrintReceipt prints a receipt to a thermal printer
func (s *PrintService) PrintReceipt(ctx context.Context, invoice *entities.Invoice, template *entities.InvoiceTemplate, printerName string) error {
	if invoice == nil {
		return errors.NewValidationError("invoice is required", "invoice cannot be nil")
	}
	if template == nil {
		return errors.NewValidationError("template is required", "template cannot be nil")
	}

	// TODO: Implement actual thermal printing functionality
	s.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"printer_name":   printerName,
	}).Info("Receipt print request received (not implemented)")

	return nil
}

// GetAvailablePrinters returns list of available printers
func (s *PrintService) GetAvailablePrinters() ([]services.PrinterInfo, error) {
	// Return mock printer data for now
	return []services.PrinterInfo{
		{
			Name:        "Default Printer",
			Description: "Default system printer",
			Status:      "ready",
			IsDefault:   true,
			SupportsA4:  true,
			SupportsA5:  true,
			Isthermal:   false,
		},
		{
			Name:        "Thermal Receipt Printer",
			Description: "80mm thermal receipt printer",
			Status:      "ready",
			IsDefault:   false,
			SupportsA4:  false,
			SupportsA5:  false,
			Isthermal:   true,
		},
	}, nil
}

// GetDefaultPrinter returns the default printer
func (s *PrintService) GetDefaultPrinter() (*services.PrinterInfo, error) {
	printers, err := s.GetAvailablePrinters()
	if err != nil {
		return nil, err
	}

	for _, printer := range printers {
		if printer.IsDefault {
			return &printer, nil
		}
	}

	// If no default printer found, return the first one
	if len(printers) > 0 {
		return &printers[0], nil
	}

	return nil, errors.NewNotFoundError("no printers available")
}

// SetDefaultPrinter sets the default printer
func (s *PrintService) SetDefaultPrinter(printerName string) error {
	if printerName == "" {
		return errors.NewValidationError("printer name is required", "printer name cannot be empty")
	}

	// Verify printer exists
	printers, err := s.GetAvailablePrinters()
	if err != nil {
		return err
	}

	found := false
	for _, printer := range printers {
		if printer.Name == printerName {
			found = true
			break
		}
	}

	if !found {
		return errors.NewNotFoundError("printer not found")
	}

	s.defaultPrinter = printerName
	s.logger.WithFields(map[string]interface{}{
		"printer_name": printerName,
	}).Info("Default printer set")

	return nil
}