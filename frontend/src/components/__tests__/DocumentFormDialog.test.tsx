import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

  it('deve renderizar campos de Data e Hora de Registro Oficial na criação de documento não-contrato', () => {
    render(
      <DocumentFormDialog
        isOpen={true}
        onClose={vi.fn()}
        editingDocument={null}
        activeTab="DECREE"
        canCreate={() => true}
        docTypesList={mockDocTypesList}
        contractTypeLabels={mockContractTypeLabels}
        onSave={vi.fn()}
        creatorId="user-1"
        municipalityId="mun-1"
      />
    );

    expect(screen.getByText('Data e Hora de Registro Oficial')).toBeInTheDocument();
    expect(screen.getByText('Data do Ato')).toBeInTheDocument();
    expect(screen.getByText('Hora do Registro')).toBeInTheDocument();
  });

  it('deve exibir opção de número manual exclusivamente para usuário MOD', () => {
    const modUser = {
      id: 'mod-1',
      username: 'mod_user',
      email: 'mod@example.com',
      role: 'MOD' as const,
      municipalityId: 'mun-1',
      createdAt: '',
      updatedAt: '',
    };

    render(
      <DocumentFormDialog
        isOpen={true}
        onClose={vi.fn()}
        editingDocument={null}
        activeTab="DECREE"
        canCreate={() => true}
        docTypesList={mockDocTypesList}
        contractTypeLabels={mockContractTypeLabels}
        onSave={vi.fn()}
        creatorId="mod-1"
        municipalityId="mun-1"
        currentUser={modUser}
      />
    );

    expect(
      screen.getByText('Lançamento de documento de acervo físico / Número manual')
    ).toBeInTheDocument();
    expect(
      screen.getByText('O número sequencial será gerado automaticamente pelo sistema.')
    ).toBeInTheDocument();
  });

  it('NÃO deve exibir opção de número manual para usuário COMMON', () => {
    const commonUser = {
      id: 'common-1',
      username: 'common_user',
      email: 'common@example.com',
      role: 'COMMON' as const,
      municipalityId: 'mun-1',
      createdAt: '',
      updatedAt: '',
    };

    render(
      <DocumentFormDialog
        isOpen={true}
        onClose={vi.fn()}
        editingDocument={null}
        activeTab="DECREE"
        canCreate={() => true}
        docTypesList={mockDocTypesList}
        contractTypeLabels={mockContractTypeLabels}
        onSave={vi.fn()}
        creatorId="common-1"
        municipalityId="mun-1"
        currentUser={commonUser}
      />
    );

    expect(
      screen.queryByText('Lançamento de documento de acervo físico / Número manual')
    ).not.toBeInTheDocument();
  });

  it('deve permitir digitar data retroativa antiga (ex: 15/03/1980) diretamente no campo Data do Ato', async () => {
    const user = userEvent.setup();
    const handleSave = vi.fn().mockResolvedValue(undefined);

    render(
      <DocumentFormDialog
        isOpen={true}
        onClose={vi.fn()}
        editingDocument={null}
        activeTab="LAW"
        canCreate={() => true}
        docTypesList={mockDocTypesList}
        contractTypeLabels={mockContractTypeLabels}
        onSave={handleSave}
        creatorId="user-1"
        municipalityId="mun-1"
      />
    );

    const dateInput = screen.getByLabelText('Data do Ato');
    await user.clear(dateInput);
    await user.type(dateInput, '15/03/1980');
    expect(dateInput).toHaveValue('15/03/1980');

    // Preenche descrição e submete
    const descInput = screen.getByPlaceholderText(/Descreva o conteúdo do documento/i);
    await user.type(descInput, 'Lei Municipal Histórica nº 10 de 1980');

    const submitBtn = screen.getByRole('button', { name: /Salvar/i });
    await user.click(submitBtn);

    expect(handleSave).toHaveBeenCalledWith(
      expect.objectContaining({
        createInput: expect.objectContaining({
          type: 'LAW',
          description: 'Lei Municipal Histórica nº 10 de 1980',
          createdAt: expect.stringContaining('1980-03-15'),
        }),
      })
    );
  });

  it('deve exibir opção de número manual para usuário MOD quando o tipo for CONTRACT', () => {
    const modUser = {
      id: 'mod-1',
      username: 'mod_user',
      email: 'mod@example.com',
      role: 'MOD' as const,
      municipalityId: 'mun-1',
      createdAt: '',
      updatedAt: '',
    };

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
        creatorId="mod-1"
        municipalityId="mun-1"
        currentUser={modUser}
      />
    );

    expect(
      screen.getByText('Lançamento de documento de acervo físico / Número manual')
    ).toBeInTheDocument();
  });

  it('NÃO deve exibir opção de número manual para usuário COMMON quando o tipo for CONTRACT', () => {
    const commonUser = {
      id: 'common-1',
      username: 'common_user',
      email: 'common@example.com',
      role: 'COMMON' as const,
      municipalityId: 'mun-1',
      createdAt: '',
      updatedAt: '',
    };

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
        creatorId="common-1"
        municipalityId="mun-1"
        currentUser={commonUser}
      />
    );

    expect(
      screen.queryByText('Lançamento de documento de acervo físico / Número manual')
    ).not.toBeInTheDocument();
  });

  it('deve submeter contrato com número manual e createdAt sincronizado com startIn', async () => {
    const user = userEvent.setup();
    const handleSave = vi.fn().mockResolvedValue(undefined);
    const modUser = {
      id: 'mod-1',
      username: 'mod_user',
      email: 'mod@example.com',
      role: 'MOD' as const,
      municipalityId: 'mun-1',
      createdAt: '',
      updatedAt: '',
    };

    render(
      <DocumentFormDialog
        isOpen={true}
        onClose={vi.fn()}
        editingDocument={null}
        activeTab="CONTRACT"
        canCreate={() => true}
        docTypesList={mockDocTypesList}
        contractTypeLabels={mockContractTypeLabels}
        onSave={handleSave}
        creatorId="mod-1"
        municipalityId="mun-1"
        currentUser={modUser}
      />
    );

    // Marca o checkbox de acervo físico
    const manualCheckbox = screen.getByTestId('manual-order-checkbox');
    await user.click(manualCheckbox);

    // Preenche o número oficial do ato
    const manualOrderInput = screen.getByLabelText(/Número Oficial do Ato/i);
    await user.type(manualOrderInput, '42');

    // Preenche a descrição
    const descInput = screen.getByPlaceholderText(/Descreva o conteúdo do documento/i);
    await user.type(descInput, 'Contrato Físico Legado nº 42 de 2024');

    // Preenche valor e duração
    const valueInput = screen.getByLabelText(/Valor \(R\$\)/i);
    await user.type(valueInput, '50000.00');

    const durationInput = screen.getByLabelText(/Duração \(meses\)/i);
    await user.type(durationInput, '12');

    // Preenche data de início retroativa (10/02/2024)
    const startDateInput = screen.getByLabelText(/Data de Início/i);
    await user.clear(startDateInput);
    await user.type(startDateInput, '10/02/2024');

    // Submete o formulário
    const submitBtn = screen.getByRole('button', { name: /Salvar/i });
    await user.click(submitBtn);

    expect(handleSave).toHaveBeenCalledWith(
      expect.objectContaining({
        createInput: expect.objectContaining({
          type: 'CONTRACT',
          description: 'Contrato Físico Legado nº 42 de 2024',
          manualOrder: 42,
          value: 50000,
          duration: 12,
          startIn: expect.stringContaining('2024-02-10'),
          createdAt: expect.stringContaining('2024-02-10'),
        }),
      })
    );
  });
});
