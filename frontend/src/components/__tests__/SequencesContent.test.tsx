import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SequencesContent } from '../SequencesContent';
import * as DocumentApi from '@/app/api/documents';
import { User, SequenceItemResponse } from '@/types';

vi.mock('@/app/api/documents', () => ({
  getMunicipalitySequences: vi.fn(),
  setMunicipalitySequenceOffset: vi.fn(),
}));

describe('SequencesContent Component', () => {
  const mockUser: User = {
    id: 'user-mod-1',
    username: 'gestor_mod',
    email: 'mod@passagem.pb.gov.br',
    role: 'MOD',
    municipalityId: 'mun-1',
    municipality: {
      id: 'mun-1',
      name: 'Passagem',
      uf: 'PB',
      createdAt: '',
      updatedAt: '',
    },
    createdAt: '',
    updatedAt: '',
  };

  const mockSequences: SequenceItemResponse[] = [
    {
      type: 'NOTICE',
      initialOrder: 1,
      currentOrder: 10,
      nextOrder: 11,
      year: 2026,
    },
    {
      type: 'DECREE',
      initialOrder: 85,
      currentOrder: 0,
      nextOrder: 85,
      year: 2026,
    },
    {
      type: 'LAW',
      initialOrder: 50,
      currentOrder: 52,
      nextOrder: 53,
      year: null,
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(DocumentApi.getMunicipalitySequences).mockResolvedValue(mockSequences);
    vi.mocked(DocumentApi.setMunicipalitySequenceOffset).mockResolvedValue({
      id: 'seq-1',
      municipalityId: 'mun-1',
      type: 'DECREE',
      year: 2026,
      initialOrder: 90,
    });
  });

  it('deve renderizar o cabeçalho e informações do município', () => {
    render(<SequencesContent currentUser={mockUser} initialData={mockSequences} />);

    expect(screen.getByText('Sequências & Marcos Iniciais')).toBeInTheDocument();
    expect(screen.getByText('Passagem (PB)')).toBeInTheDocument();
  });

  it('deve listar as séries documentais com marco inicial, último emitido e próximo a ser gerado', () => {
    render(<SequencesContent currentUser={mockUser} initialData={mockSequences} />);

    expect(screen.getByText('Ofício')).toBeInTheDocument();
    expect(screen.getByText('Decreto')).toBeInTheDocument();
    expect(screen.getByText('Lei Municipal')).toBeInTheDocument();

    // Próximos números
    expect(screen.getByText('11')).toBeInTheDocument();
    expect(screen.getByText('85')).toBeInTheDocument();
    expect(screen.getByText('53')).toBeInTheDocument();

    // Lei é perpétua
    expect(screen.getByText('Perpétuo (Contínuo)')).toBeInTheDocument();
  });

  it('deve chamar setMunicipalitySequenceOffset ao alterar o marco inicial e clicar em Salvar', async () => {
    const user = userEvent.setup();
    render(<SequencesContent currentUser={mockUser} initialData={mockSequences} />);

    // Localiza input de marco inicial para Decreto (valor inicial 85)
    const decreeInput = screen.getByLabelText('Marco inicial para Decreto');
    expect(decreeInput).toHaveValue(85);

    await user.clear(decreeInput);
    await user.type(decreeInput, '90');

    // Clica no botão Salvar correspondente
    const saveButtons = screen.getAllByRole('button', { name: /Salvar/i });
    // Botão 1 é para NOTICE, 2 para DECREE, 3 para LAW
    await user.click(saveButtons[1]);

    await waitFor(() => {
      expect(DocumentApi.setMunicipalitySequenceOffset).toHaveBeenCalledWith(
        'mun-1',
        expect.objectContaining({
          type: 'DECREE',
          initialOrder: 90,
        })
      );
    });
  });
});
