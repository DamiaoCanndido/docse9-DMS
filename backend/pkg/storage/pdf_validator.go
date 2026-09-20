package storage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

var (
	ErrEmptyPDF       = errors.New("arquivo PDF vazio")
	ErrPDFMissingOCR  = errors.New("o arquivo PDF não possui camada de texto pesquisável (OCR)")
	ErrInvalidPDFFile = errors.New("arquivo inválido ou corrompido, não é um PDF válido")
)

// ValidatePDFText inspeciona um leitor de PDF e conta o número de caracteres alfanuméricos
// de texto pesquisável extraíveis. Retorna true se a contagem for maior ou igual a minChars.
func ValidatePDFText(r io.ReaderAt, size int64, minChars int) (valid bool, chars int, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			valid = false
			chars = 0
			err = fmt.Errorf("falha ao interpretar estrutura interna do PDF: %v", rec)
		}
	}()

	if size <= 0 {
		return false, 0, ErrEmptyPDF
	}

	reader, err := pdf.NewReader(r, size)
	if err != nil {
		return false, 0, fmt.Errorf("%w: %v", ErrInvalidPDFFile, err)
	}

	numPages := reader.NumPage()
	if numPages <= 0 {
		return false, 0, ErrPDFMissingOCR
	}

	// Limita a inspeção às primeiras 50 páginas para mitigar DoS de CPU (PDF bomb)
	maxPagesToCheck := numPages
	if maxPagesToCheck > 50 {
		maxPagesToCheck = 50
	}

	totalTextChars := 0
	var b bytes.Buffer

	for pageIndex := 1; pageIndex <= maxPagesToCheck; pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		plainText, err := page.GetPlainText(nil)
		if err == nil && plainText != "" {
			b.WriteString(plainText)
		}

		// Conta caracteres alfanuméricos
		for _, r := range b.String() {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				totalTextChars++
			}
		}

		b.Reset()

		// Early exit se já atingiu o limite mínimo requerido
		if totalTextChars >= minChars {
			return true, totalTextChars, nil
		}
	}

	if totalTextChars < minChars {
		return false, totalTextChars, ErrPDFMissingOCR
	}

	return true, totalTextChars, nil
}

// IsPDFContent verifica os primeiros bytes mágicos do cabeçalho PDF (%PDF-).
func IsPDFHeader(header []byte) bool {
	return len(header) >= 5 && strings.HasPrefix(string(header), "%PDF-")
}
