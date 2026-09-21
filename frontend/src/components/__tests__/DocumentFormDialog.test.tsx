import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DocumentFormDialog, DocumentTypeOption } from '../documents/DocumentFormDialog';
import { ContractType } from '@/types';

// Mock validatePDFClientSide to avoid pdfjs worker issues in node/jsdom
vi.mock('@/lib/pdf-validator', () => ({
  validatePDFClientSide: vi.fn().mockResolvedValue({ valid: true }),
}));

describe('DocumentFormDialog Component', () => {
  const mockDocTypesList: DocumentTypeOption[] = [
    { value: 'NOTICE', label: 'Ofícios', singleLabel: 'Ofício' },
    { value: 'DECREE', label: 'Decretos', singleLabel: 'Decreto' },
    { value: 'ORDINANCE', label: 'Portarias', singleLabel: 'Portaria' },
    { value: 'LAW', label: 'Leis', singleLabel: 'Lei' },
    { value: 'CONTRACT', label: 'Contratos', singleLabel: 'Contrato' },
  ];

  const mockContractTypeLabels: Record<ContractType, string> = {
    service: 'Prestação de Serviço',
    bidding: 'Licitação',
    publicinterest: 'Interesse Público',
  };

  it('deve renderizar o diálogo com layout responsivo para criação de Ofício', () => {
    render(
      <DocumentFormDialog
        isOpen={true}
        onClose={vi.fn()}
        editingDocument={null}
        activeTab="NOTICE"
        canCreate={() => true}
        docTypesList={mockDocTypesList}
        contractTypeLabels={mockContractTypeLabels}
        onSave={vi.fn()}
        creatorId="user-1"
        municipalityId="mun-1"
      />
    );

    expect(screen.getByText('Novo Documento Oficial')).toBeInTheDocument();
    expect(screen.getByText('Tipo de Documento')).toBeInTheDocument();
    expect(screen.getByText('Anexo Oficial (PDF com OCR)')).toBeInTheDocument();
    expect(screen.getByText('Descrição / Ementa')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Cancelar/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Salvar/i })).toBeInTheDocument();
  });

  it('deve renderizar os campos específicos de Contrato quando o tipo for CONTRACT', () => {
    render(
      <DocumentFormDialog
        isOpen={true}
        onClose={vi.fn()}
        editingDocument={null}
        activeTab="CONTRACT"
        canCreate={() => true}
        docTypesList={mockDocTypesList}
        contractTypeLabels={mockContractTypeLabels}
        onSave={vi.fn()}
        creatorId="user-1"
        municipalityId="mun-1"
      />
    );

    expect(screen.getByText('Detalhes do Contrato')).toBeInTheDocument();
    expect(screen.getByText('Tipo de Contrato')).toBeInTheDocument();
    expect(screen.getByText('Valor (R$)')).toBeInTheDocument();
    expect(screen.getByText('Duração (meses)')).toBeInTheDocument();
    expect(screen.getByText('Data de Início')).toBeInTheDocument();
    expect(screen.getByText('Hora de Início')).toBeInTheDocument();
  });
});
