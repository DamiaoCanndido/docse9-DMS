package storage_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/storage"
	"github.com/stretchr/testify/assert"
)

// buildValidPDF monta em memória um PDF estruturalmente válido com tabela de referências (xref) exata.
func buildValidPDF(textContent string) []byte {
	var body bytes.Buffer
	body.WriteString("%PDF-1.4\n")

	var offsets []int

	// Objeto 1: Catalog
	offsets = append(offsets, body.Len())
	body.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Objeto 2: Pages
	offsets = append(offsets, body.Len())
	body.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	if textContent != "" {
		// Objeto 3: Page com Font e Contents
		offsets = append(offsets, body.Len())
		body.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n")

		// Objeto 4: Font Helvetica
		offsets = append(offsets, body.Len())
		body.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

		// Objeto 5: Stream de Texto
		streamContent := fmt.Sprintf("BT\n/F1 12 Tf\n72 712 Td\n(%s) Tj\nET\n", textContent)
		offsets = append(offsets, body.Len())
		body.WriteString(fmt.Sprintf("5 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(streamContent), streamContent))
	} else {
		// Objeto 3: Page sem Font e Contents vazios (imagem raster / sem texto)
		offsets = append(offsets, body.Len())
		body.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n")

		// Objeto 4: Stream Vazio
		offsets = append(offsets, body.Len())
		body.WriteString("4 0 obj\n<< /Length 0 >>\nstream\n\nendstream\nendobj\n")
	}

	startXref := body.Len()
	numObjs := len(offsets) + 1

	body.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", numObjs))
	for _, off := range offsets {
		body.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}

	body.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", numObjs, startXref))

	return body.Bytes()
}

func TestValidatePDFText_Success(t *testing.T) {
	text := "Prefeitura Municipal de Aracaju Estado de Sergipe Decreto Oficial 2026 Publicacao Integrada"
	pdfBytes := buildValidPDF(text)
	reader := bytes.NewReader(pdfBytes)

	valid, count, err := storage.ValidatePDFText(reader, int64(len(pdfBytes)), 30)
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.GreaterOrEqual(t, count, 30)
}

func TestValidatePDFText_MissingOCR(t *testing.T) {
	pdfBytes := buildValidPDF("") // PDF válido porém sem camada de texto (scanned image placeholder)
	reader := bytes.NewReader(pdfBytes)

	valid, count, err := storage.ValidatePDFText(reader, int64(len(pdfBytes)), 50)
	assert.Error(t, err)
	assert.ErrorIs(t, err, storage.ErrPDFMissingOCR)
	assert.False(t, valid)
	assert.Equal(t, 0, count)
}

func TestValidatePDFText_InsufficientChars(t *testing.T) {
	pdfBytes := buildValidPDF("Ola") // Apenas 3 caracteres
	reader := bytes.NewReader(pdfBytes)

	valid, count, err := storage.ValidatePDFText(reader, int64(len(pdfBytes)), 50)
	assert.Error(t, err)
	assert.ErrorIs(t, err, storage.ErrPDFMissingOCR)
	assert.False(t, valid)
	assert.Equal(t, 3, count)
}

func TestValidatePDFText_InvalidOrCorrupt(t *testing.T) {
	data := []byte("Not a valid PDF at all")
	reader := bytes.NewReader(data)

	valid, count, err := storage.ValidatePDFText(reader, int64(len(data)), 50)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Equal(t, 0, count)
}

func TestValidatePDFText_EmptyFile(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	valid, count, err := storage.ValidatePDFText(reader, 0, 50)
	assert.ErrorIs(t, err, storage.ErrEmptyPDF)
	assert.False(t, valid)
	assert.Equal(t, 0, count)
}

func TestIsPDFHeader(t *testing.T) {
	assert.True(t, storage.IsPDFHeader([]byte("%PDF-1.4")))
	assert.True(t, storage.IsPDFHeader([]byte("%PDF-2.0 rest of content")))
	assert.False(t, storage.IsPDFHeader([]byte("GIF89a")))
	assert.False(t, storage.IsPDFHeader([]byte("PK\x03\x04")))
	assert.False(t, storage.IsPDFHeader([]byte("")))
	assert.False(t, storage.IsPDFHeader([]byte(strings.Repeat("a", 3))))
}

func TestValidatePDFText_CorruptXrefDoesNotPanic(t *testing.T) {
	corruptData := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\nxref\n0 99999\nstartxref\n10\n%%EOF")
	reader := bytes.NewReader(corruptData)

	assert.NotPanics(t, func() {
		valid, count, err := storage.ValidatePDFText(reader, int64(len(corruptData)), 50)
		assert.False(t, valid)
		assert.Equal(t, 0, count)
		assert.Error(t, err)
	})
}
