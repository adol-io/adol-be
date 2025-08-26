package services

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/nicklaros/adol/internal/domain/entities"
	"github.com/nicklaros/adol/internal/domain/services"
	"github.com/nicklaros/adol/pkg/errors"
	"github.com/nicklaros/adol/pkg/logger"
)

// EmailService implements the domain EmailService interface
type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
	logger       logger.Logger
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig, logger logger.Logger) services.EmailService {
	return &EmailService{
		smtpHost:     config.SMTPHost,
		smtpPort:     config.SMTPPort,
		smtpUsername: config.SMTPUsername,
		smtpPassword: config.SMTPPassword,
		fromEmail:    config.FromEmail,
		fromName:     config.FromName,
		logger:       logger,
	}
}

// SendInvoiceEmail sends an invoice via email
func (s *EmailService) SendInvoiceEmail(ctx context.Context, invoice *entities.Invoice, recipient string, pdfData []byte) error {
	if invoice == nil {
		return errors.NewValidationError("invoice is required", "invoice cannot be nil")
	}
	if recipient == "" {
		return errors.NewValidationError("recipient is required", "recipient email cannot be empty")
	}
	if len(pdfData) == 0 {
		return errors.NewValidationError("PDF data is required", "PDF data cannot be empty")
	}

	// Validate email configuration
	if err := s.validateConfig(); err != nil {
		return err
	}

	subject := fmt.Sprintf("Invoice %s - %s", invoice.InvoiceNumber, invoice.CustomerName)

	// Create email body
	body := s.createInvoiceEmailBody(invoice)

	// Create email message with attachment
	message := s.createEmailMessage(recipient, subject, body, pdfData, fmt.Sprintf("invoice_%s.pdf", invoice.InvoiceNumber))

	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(addr, auth, s.fromEmail, []string{recipient}, []byte(message))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"invoice_id": invoice.ID,
			"recipient":  recipient,
			"error":      err.Error(),
		}).Error("Failed to send invoice email")
		return errors.NewInternalError("failed to send email", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"recipient":      recipient,
	}).Info("Invoice email sent successfully")

	return nil
}

// SendReceiptEmail sends a receipt via email
func (s *EmailService) SendReceiptEmail(ctx context.Context, invoice *entities.Invoice, recipient string, pdfData []byte) error {
	if invoice == nil {
		return errors.NewValidationError("invoice is required", "invoice cannot be nil")
	}
	if recipient == "" {
		return errors.NewValidationError("recipient is required", "recipient email cannot be empty")
	}
	if len(pdfData) == 0 {
		return errors.NewValidationError("PDF data is required", "PDF data cannot be empty")
	}

	// Validate email configuration
	if err := s.validateConfig(); err != nil {
		return err
	}

	subject := fmt.Sprintf("Receipt - Invoice #%s", invoice.InvoiceNumber)
	
	// Create receipt email body
	body := s.createReceiptEmailBody(invoice)
	
	// Create email message with PDF attachment
	message := s.createEmailMessage(recipient, subject, body, pdfData, fmt.Sprintf("receipt_%s.pdf", invoice.InvoiceNumber))
	
	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{recipient}, []byte(message))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"invoice_id": invoice.ID,
			"recipient": recipient,
			"error": err.Error(),
		}).Error("Failed to send receipt email")
		return errors.NewInternalError("failed to send receipt email", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"invoice_id": invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"recipient": recipient,
	}).Info("Receipt email sent successfully")

	return nil
}

// SendPaymentConfirmation sends payment confirmation email
func (s *EmailService) SendPaymentConfirmation(ctx context.Context, invoice *entities.Invoice, recipient string) error {
	if invoice == nil {
		return errors.NewValidationError("invoice is required", "invoice cannot be nil")
	}
	if recipient == "" {
		return errors.NewValidationError("recipient is required", "recipient email cannot be empty")
	}

	// Validate email configuration
	if err := s.validateConfig(); err != nil {
		return err
	}

	subject := fmt.Sprintf("Payment Confirmation - Invoice %s", invoice.InvoiceNumber)
	
	// Create payment confirmation email body
	body := s.createPaymentConfirmationEmailBody(invoice)
	
	// Create simple email message (no attachment for confirmation)
	message := s.createSimpleEmailMessage(recipient, subject, body)
	
	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{recipient}, []byte(message))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"invoice_id": invoice.ID,
			"recipient":  recipient,
			"error":      err.Error(),
		}).Error("Failed to send payment confirmation email")
		return errors.NewInternalError("failed to send payment confirmation email", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"recipient":      recipient,
	}).Info("Payment confirmation email sent successfully")

	return nil
}

// SendInvoiceReminder sends an invoice reminder email
func (s *EmailService) SendInvoiceReminder(ctx context.Context, invoice *entities.Invoice, recipient string) error {
	if invoice == nil {
		return errors.NewValidationError("invoice is required", "invoice cannot be nil")
	}
	if recipient == "" {
		return errors.NewValidationError("recipient is required", "recipient email cannot be empty")
	}

	// Validate email configuration
	if err := s.validateConfig(); err != nil {
		return err
	}

	subject := fmt.Sprintf("Payment Reminder - Invoice %s", invoice.InvoiceNumber)

	// Create reminder email body
	body := s.createReminderEmailBody(invoice)

	// Create simple email message (no attachment for reminder)
	message := s.createSimpleEmailMessage(recipient, subject, body)

	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(addr, auth, s.fromEmail, []string{recipient}, []byte(message))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"invoice_id": invoice.ID,
			"recipient":  recipient,
			"error":      err.Error(),
		}).Error("Failed to send invoice reminder email")
		return errors.NewInternalError("failed to send reminder email", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"recipient":      recipient,
	}).Info("Invoice reminder email sent successfully")

	return nil
}

// SendOverdueNotice sends an overdue payment notice
func (s *EmailService) SendOverdueNotice(ctx context.Context, invoice *entities.Invoice, recipient string) error {
	if invoice == nil {
		return errors.NewValidationError("invoice is required", "invoice cannot be nil")
	}
	if recipient == "" {
		return errors.NewValidationError("recipient is required", "recipient email cannot be empty")
	}

	// Validate email configuration
	if err := s.validateConfig(); err != nil {
		return err
	}

	subject := fmt.Sprintf("OVERDUE PAYMENT NOTICE - Invoice %s", invoice.InvoiceNumber)
	
	// Create overdue notice email body
	body := s.createOverdueNoticeEmailBody(invoice)
	
	// Create simple email message (no attachment for overdue notice)
	message := s.createSimpleEmailMessage(recipient, subject, body)
	
	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{recipient}, []byte(message))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"invoice_id": invoice.ID,
			"recipient":  recipient,
			"error":      err.Error(),
		}).Error("Failed to send overdue notice email")
		return errors.NewInternalError("failed to send overdue notice email", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"invoice_id":     invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"recipient":      recipient,
	}).Info("Overdue notice email sent successfully")

	return nil
}

// ValidateEmailAddress validates an email address
func (s *EmailService) ValidateEmailAddress(email string) bool {
	// Simple email validation - in production, use a proper library
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// Helper methods

func (s *EmailService) validateConfig() error {
	if s.smtpHost == "" {
		return errors.NewValidationError("SMTP host is required", "SMTP host cannot be empty")
	}
	if s.smtpPort == "" {
		return errors.NewValidationError("SMTP port is required", "SMTP port cannot be empty")
	}
	if s.smtpUsername == "" {
		return errors.NewValidationError("SMTP username is required", "SMTP username cannot be empty")
	}
	if s.smtpPassword == "" {
		return errors.NewValidationError("SMTP password is required", "SMTP password cannot be empty")
	}
	if s.fromEmail == "" {
		return errors.NewValidationError("From email is required", "From email cannot be empty")
	}
	return nil
}

func (s *EmailService) createInvoiceEmailBody(invoice *entities.Invoice) string {
	var body strings.Builder

	body.WriteString("Dear ")
	body.WriteString(invoice.CustomerName)
	body.WriteString(",\n\n")

	body.WriteString("Thank you for your business! Please find attached your invoice ")
	body.WriteString(invoice.InvoiceNumber)
	body.WriteString(".\n\n")

	body.WriteString("Invoice Details:\n")
	body.WriteString(fmt.Sprintf("Invoice Number: %s\n", invoice.InvoiceNumber))
	body.WriteString(fmt.Sprintf("Invoice Date: %s\n", invoice.CreatedAt.Format("January 2, 2006")))
	body.WriteString(fmt.Sprintf("Total Amount: $%.2f\n", invoice.TotalAmount.InexactFloat64()))

	if invoice.DueDate != nil {
		body.WriteString(fmt.Sprintf("Due Date: %s\n", invoice.DueDate.Format("January 2, 2006")))
	}

	body.WriteString("\n")

	if invoice.Status == entities.InvoiceStatusPaid {
		body.WriteString("This invoice has been paid. Thank you!\n\n")
	} else {
		body.WriteString("Please process payment by the due date to avoid any late fees.\n\n")
	}

	body.WriteString("If you have any questions about this invoice, please contact us.\n\n")
	body.WriteString("Best regards,\n")
	body.WriteString("ADOL Point of Sale Team")

	return body.String()
}

func (s *EmailService) createReceiptEmailBody(invoice *entities.Invoice) string {
	var body strings.Builder
	
	body.WriteString("Dear ")
	body.WriteString(invoice.CustomerName)
	body.WriteString(",\n\n")
	
	body.WriteString("Thank you for your purchase! Please find your receipt details below.\n\n")
	
	body.WriteString("Receipt Details:\n")
	body.WriteString(fmt.Sprintf("Invoice Number: %s\n", invoice.InvoiceNumber))
	body.WriteString(fmt.Sprintf("Invoice Date: %s\n", invoice.CreatedAt.Format("January 2, 2006")))
	body.WriteString(fmt.Sprintf("Total Amount: $%.2f\n", invoice.TotalAmount.InexactFloat64()))
	if invoice.PaymentMethod != "" {
		body.WriteString(fmt.Sprintf("Payment Method: %s\n", invoice.PaymentMethod))
	}
	
	body.WriteString("\n")
	body.WriteString("Items Purchased:\n")
	for _, item := range invoice.Items {
		body.WriteString(fmt.Sprintf("- %s x%d: $%.2f\n", item.ProductName, item.Quantity, item.TotalPrice.InexactFloat64()))
	}
	
	body.WriteString("\n")
	body.WriteString("Thank you for shopping with us!\n\n")
	body.WriteString("Best regards,\n")
	body.WriteString("ADOL Point of Sale Team")
	
	return body.String()
}

func (s *EmailService) createReminderEmailBody(invoice *entities.Invoice) string {
	var body strings.Builder

	body.WriteString("Dear ")
	body.WriteString(invoice.CustomerName)
	body.WriteString(",\n\n")

	body.WriteString("This is a friendly reminder that your invoice ")
	body.WriteString(invoice.InvoiceNumber)
	body.WriteString(" is pending payment.\n\n")

	body.WriteString("Invoice Details:\n")
	body.WriteString(fmt.Sprintf("Invoice Number: %s\n", invoice.InvoiceNumber))
	body.WriteString(fmt.Sprintf("Invoice Date: %s\n", invoice.CreatedAt.Format("January 2, 2006")))
	body.WriteString(fmt.Sprintf("Total Amount: $%.2f\n", invoice.TotalAmount.InexactFloat64()))

	if invoice.DueDate != nil {
		body.WriteString(fmt.Sprintf("Due Date: %s\n", invoice.DueDate.Format("January 2, 2006")))
	}

	body.WriteString("\n")
	body.WriteString("Please process payment at your earliest convenience.\n\n")
	body.WriteString("If you have already made payment, please disregard this reminder.\n")
	body.WriteString("If you have any questions, please contact us.\n\n")
	body.WriteString("Best regards,\n")
	body.WriteString("ADOL Point of Sale Team")

	return body.String()
}

func (s *EmailService) createEmailMessage(to, subject, body string, attachment []byte, filename string) string {
	var msg strings.Builder

	// Email headers
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.fromEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/mixed; boundary=\"boundary123\"\r\n")
	msg.WriteString("\r\n")

	// Email body
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	msg.WriteString("\r\n\r\n")

	// PDF attachment
	msg.WriteString("--boundary123\r\n")
	msg.WriteString("Content-Type: application/pdf\r\n")
	msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
	msg.WriteString("\r\n")

	// Convert attachment to base64
	encoded := s.encodeBase64(attachment)
	msg.WriteString(encoded)
	msg.WriteString("\r\n")

	msg.WriteString("--boundary123--\r\n")

	return msg.String()
}

func (s *EmailService) createSimpleEmailMessage(to, subject, body string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.fromEmail))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return msg.String()
}

func (s *EmailService) encodeBase64(data []byte) string {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	encoded := make([]byte, (len(data)+2)/3*4)

	for i := 0; i < len(data); i += 3 {
		var chunk uint32
		chunkSize := 3

		if i+2 >= len(data) {
			chunkSize = len(data) - i
		}

		for j := 0; j < chunkSize; j++ {
			chunk = (chunk << 8) | uint32(data[i+j])
		}

		for j := chunkSize; j < 3; j++ {
			chunk = chunk << 8
		}

		for j := 0; j < 4; j++ {
			if chunkSize == 1 && j >= 2 {
				encoded[i/3*4+j] = '='
			} else if chunkSize == 2 && j >= 3 {
				encoded[i/3*4+j] = '='
			} else {
				encoded[i/3*4+j] = base64Table[(chunk>>(18-j*6))&0x3F]
			}
		}
	}

	// Add line breaks every 76 characters
	var result strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		result.Write(encoded[i:end])
		result.WriteString("\r\n")
	}

	return result.String()
}

func (s *EmailService) createPaymentConfirmationEmailBody(invoice *entities.Invoice) string {
	var body strings.Builder
	
	body.WriteString("Dear ")
	body.WriteString(invoice.CustomerName)
	body.WriteString(",\n\n")
	
	body.WriteString("Thank you! We have received your payment for invoice ")
	body.WriteString(invoice.InvoiceNumber)
	body.WriteString(".\n\n")
	
	body.WriteString("Payment Details:\n")
	body.WriteString(fmt.Sprintf("Invoice Number: %s\n", invoice.InvoiceNumber))
	body.WriteString(fmt.Sprintf("Invoice Date: %s\n", invoice.CreatedAt.Format("January 2, 2006")))
	body.WriteString(fmt.Sprintf("Total Amount: $%.2f\n", invoice.TotalAmount.InexactFloat64()))
	if invoice.PaidAt != nil {
		body.WriteString(fmt.Sprintf("Payment Date: %s\n", invoice.PaidAt.Format("January 2, 2006")))
	}
	
	body.WriteString("\n")
	body.WriteString("Your payment has been processed successfully and your account is now up to date.\n\n")
	body.WriteString("Thank you for your business!\n\n")
	body.WriteString("Best regards,\n")
	body.WriteString("ADOL Point of Sale Team")
	
	return body.String()
}

func (s *EmailService) createOverdueNoticeEmailBody(invoice *entities.Invoice) string {
	var body strings.Builder
	
	body.WriteString("Dear ")
	body.WriteString(invoice.CustomerName)
	body.WriteString(",\n\n")
	
	body.WriteString("URGENT: Your invoice ")
	body.WriteString(invoice.InvoiceNumber)
	body.WriteString(" is now OVERDUE and requires immediate payment.\n\n")
	
	body.WriteString("Invoice Details:\n")
	body.WriteString(fmt.Sprintf("Invoice Number: %s\n", invoice.InvoiceNumber))
	body.WriteString(fmt.Sprintf("Invoice Date: %s\n", invoice.CreatedAt.Format("January 2, 2006")))
	if invoice.DueDate != nil {
		body.WriteString(fmt.Sprintf("Due Date: %s\n", invoice.DueDate.Format("January 2, 2006")))
	}
	body.WriteString(fmt.Sprintf("Total Amount: $%.2f\n", invoice.TotalAmount.InexactFloat64()))
	
	body.WriteString("\n")
	body.WriteString("Please make payment immediately to avoid additional late fees or collection actions.\n\n")
	body.WriteString("If you have any questions or need to arrange a payment plan, please contact us urgently.\n\n")
	body.WriteString("Regards,\n")
	body.WriteString("ADOL Point of Sale Accounts Department")
	
	return body.String()
}