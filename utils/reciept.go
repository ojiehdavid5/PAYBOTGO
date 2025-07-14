package utils

import (
	"fmt"
	"time"

	"github.com/phpdave11/gofpdf"
)

func GenerateReceiptPDF(to string, amount int, txnID string) (string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(40, 10, "💰 STUDENT PAYBOT RECEIPT")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("To: %s", to))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Amount: ₦%.2f", float64(amount)/100))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Date: %s", time.Now().Format("January 2, 2006")))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Transaction ID: %s", txnID))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Status: SUCCESSFUL")
	pdf.Ln(12)
	pdf.Cell(40, 10, "Thank you for using StudentPay!")

	filename := fmt.Sprintf("receipt_%s.pdf", txnID)
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}
