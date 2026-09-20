import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DocumentTable } from '../documents/DocumentTable';
import { Document, User } from '@/types';

describe('DocumentTable Attachment Actions', () => {
  const mockUser: User = {
    id: 'user-1',
    username: 'mod_user',
    email: 'mod@example.com',
    role: 'MOD',
    municipalityId: 'mun-1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  const sampleDocWithFile: Document = {
    id: 'doc-1',
    type: 'NOTICE',
    order: 10,
    description: 'Oficio de Solicitacao com Anexo',
    fileKey: 'tenants/mun-1/NOTICE/2026/doc-1/file.pdf',
    creatorId: 'user-1',
    municipalityId: 'mun-1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  const sampleDocWithoutFile: Document = {
    id: 'doc-2',
    type: 'NOTICE',
    order: 11,
    description: 'Oficio sem Arquivo Anexado',
    fileKey: '',
    creatorId: 'user-1',
    municipalityId: 'mun-1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  it('deve exibir indicador de PDF com OCR para documentos com fileKey', () => {
    render(
      <DocumentTable
        documents={[sampleDocWithFile, sampleDocWithoutFile]}
        activeTab="NOTICE"
        activeTabLabel="Ofícios"
        viewTrash={false}
        currentUser={mockUser}
        isPending={false}
        canEdit={() => true}
        canDelete={() => true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onRestore={vi.fn()}
        onHardDelete={vi.fn()}
        pagination={{ page: 1, pageSize: 10, total: 2 }}
        onPageChange={vi.fn()}
        onPageSizeChange={vi.fn()}
      />
    );

    expect(screen.getByText('PDF Anexado (OCR)')).toBeInTheDocument();
    expect(screen.getByText('Sem anexo')).toBeInTheDocument();
  });

  it('deve chamar onViewAttachment ao clicar no botão de anexo', async () => {
    const user = userEvent.setup();
    const handleView = vi.fn();

    render(
      <DocumentTable
        documents={[sampleDocWithFile]}
        activeTab="NOTICE"
        activeTabLabel="Ofícios"
        viewTrash={false}
        currentUser={mockUser}
        isPending={false}
        canEdit={() => true}
        canDelete={() => true}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        onRestore={vi.fn()}
        onHardDelete={vi.fn()}
        onViewAttachment={handleView}
        pagination={{ page: 1, pageSize: 10, total: 1 }}
        onPageChange={vi.fn()}
        onPageSizeChange={vi.fn()}
      />
    );

    const button = screen.getByText('PDF Anexado (OCR)');
    await user.click(button);

    expect(handleView).toHaveBeenCalledTimes(1);
    expect(handleView).toHaveBeenCalledWith(sampleDocWithFile);
  });
});
