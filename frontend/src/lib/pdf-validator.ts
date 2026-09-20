/**
 * Validador client-side de arquivos PDF e detecção preliminar de OCR (camada de texto pesquisável).
 */

export interface PDFValidationResult {
  valid: boolean;
  error?: string;
}

const MAX_FILE_SIZE_BYTES = 25 * 1024 * 1024; // 25 MB

/**
 * Valida se o arquivo selecionado é um PDF válido e se possui indícios claros de camada de texto (OCR).
 */
export async function validatePDFClientSide(file: File): Promise<PDFValidationResult> {
  // 1. Validação de extensão e tipo MIME
  const isPDFMime = file.type === 'application/pdf' || file.type === '';
  const hasPDFExtension = file.name.toLowerCase().endsWith('.pdf');

  if (!isPDFMime || !hasPDFExtension) {
    return {
      valid: false,
      error: 'Formato inválido. Somente arquivos PDF (.pdf) são permitidos.',
    };
  }

  // 2. Validação de tamanho máximo
  if (file.size <= 0) {
    return {
      valid: false,
      error: 'O arquivo selecionado está vazio.',
    };
  }

  if (file.size > MAX_FILE_SIZE_BYTES) {
    return {
      valid: false,
      error: 'O arquivo excede o limite máximo permitido de 25 MB.',
    };
  }

  // 3. Inspeção de cabeçalho mágico (%PDF-)
  try {
    const buffer = await file.slice(0, 1024).arrayBuffer();
    const header = new TextDecoder('ascii').decode(buffer);
    if (!header.startsWith('%PDF-')) {
      return {
        valid: false,
        error: 'O arquivo não possui uma estrutura válida de PDF.',
      };
    }
  } catch {
    return {
      valid: false,
      error: 'Não foi possível ler o cabeçalho do arquivo.',
    };
  }

  // 4. Verificação preliminar de presença de texto / OCR
  // Inspeciona os blocos do arquivo procurando recursos de fonte (/Font) e operadores de texto (BT, Tj, TJ)
  try {
    const sampleSize = Math.min(file.size, 2 * 1024 * 1024); // primeiros 2MB
    const sampleBuffer = await file.slice(0, sampleSize).arrayBuffer();
    const sampleText = new TextDecoder('latin1').decode(sampleBuffer);

    const hasFontDefinition = sampleText.includes('/Font') || sampleText.includes('/Type /Font') || sampleText.includes('/Subtype /Type1') || sampleText.includes('/Subtype /TrueType') || sampleText.includes('/ToUnicode');
    const hasTextBlocks = sampleText.includes('BT') && (sampleText.includes('Tj') || sampleText.includes('TJ') || sampleText.includes('ET'));

    if (!hasFontDefinition && !hasTextBlocks) {
      return {
        valid: false,
        error: 'O arquivo PDF selecionado parece ser apenas uma imagem digitalizada sem camada de texto pesquisável (OCR). Por favor, realize a digitalização com OCR habilitado antes de anexar.',
      };
    }
  } catch {
    // Se a leitura da amostra falhar por limitação de memória, deixa a validação autoritativa para o backend
  }

  return { valid: true };
}
