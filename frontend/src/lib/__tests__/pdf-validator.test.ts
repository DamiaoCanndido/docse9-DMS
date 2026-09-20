import { describe, it, expect } from 'vitest';
import { validatePDFClientSide } from '../pdf-validator';

describe('validatePDFClientSide', () => {
  it('deve rejeitar arquivos com extensão diferente de .pdf', async () => {
    const file = new File(['dummy content'], 'documento.docx', { type: 'application/msword' });
    const result = await validatePDFClientSide(file);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('Somente arquivos PDF (.pdf) são permitidos');
  });

  it('deve rejeitar arquivo vazio', async () => {
    const file = new File([], 'vazio.pdf', { type: 'application/pdf' });
    const result = await validatePDFClientSide(file);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('vazio');
  });

  it('deve rejeitar arquivo sem cabeçalho %PDF-', async () => {
    const file = new File(['GIF89a corrupted file'], 'fake.pdf', { type: 'application/pdf' });
    const result = await validatePDFClientSide(file);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('estrutura válida de PDF');
  });

  it('deve rejeitar PDF sem camada de texto/OCR', async () => {
    // Cabeçalho PDF mas sem blocos de texto ou fontes
    const rawData = '%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n';
    const file = new File([rawData], 'scan_imagem.pdf', { type: 'application/pdf' });
    const result = await validatePDFClientSide(file);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('OCR');
  });

  it('deve aprovar PDF válido com fontes e operadores de texto', async () => {
    const validData = '%PDF-1.4\n/Font << /F1 >>\nBT\n(Prefeitura Municipal de Aracaju) Tj\nET\n';
    const file = new File([validData], 'decreto_com_ocr.pdf', { type: 'application/pdf' });
    const result = await validatePDFClientSide(file);
    expect(result.valid).toBe(true);
    expect(result.error).toBeUndefined();
  });
});
