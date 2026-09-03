import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { NotificationCenter } from '../NotificationCenter';
import * as AuthContext from '@/context/AuthContext';
import * as DocumentsApi from '@/app/api/documents';
import { User, Document } from '@/types';

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    refresh: vi.fn(),
  }),
}));

describe('NotificationCenter Component', () => {
  const mockUser: User = {
    id: 'user-test-123',
    username: 'gestor.municipal',
    email: 'gestor@passagem.pb.gov.br',
    role: 'MOD',
    municipalityId: 'mun-test-1',
    municipality: {
      id: 'mun-test-1',
      name: 'Passagem',
      uf: 'PB',
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    },
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  };

  const today = new Date();
  
  // Contrato 1: Inicia há 11 meses e duração 12 meses -> vence em ~30 dias (crítico/urgente)
  const expiringStartDate = new Date(today);
  expiringStartDate.setMonth(expiringStartDate.getMonth() - 11);

  // Contrato 2: Inicia há 10 meses e duração 12 meses -> vence em ~60 dias (atenção)
  const warningStartDate = new Date(today);
  warningStartDate.setMonth(warningStartDate.getMonth() - 10);

  const mockContracts: Document[] = [
    {
      id: 'contract-urgent-1',
      type: 'CONTRACT',
      order: 101,
      description: 'Contrato de Coleta de Resíduos Sólidos',
      fileKey: '',
      creatorId: mockUser.id,
      municipalityId: mockUser.municipalityId,
      contractType: 'service',
      value: 120000,
      duration: 12,
      startIn: expiringStartDate.toISOString(),
      createdAt: expiringStartDate.toISOString(),
      updatedAt: expiringStartDate.toISOString(),
    },
    {
      id: 'contract-warning-2',
      type: 'CONTRACT',
      order: 102,
      description: 'Contrato de Manutenção de Software',
      fileKey: '',
      creatorId: mockUser.id,
      municipalityId: mockUser.municipalityId,
      contractType: 'bidding',
      value: 45000,
      duration: 12,
      startIn: warningStartDate.toISOString(),
      createdAt: warningStartDate.toISOString(),
      updatedAt: warningStartDate.toISOString(),
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();

    vi.spyOn(AuthContext, 'useAuth').mockReturnValue({
      user: mockUser,
      loading: false,
      login: vi.fn(),
      logout: vi.fn(),
      changePassword: vi.fn(),
      updateUser: vi.fn(),
      isAuthenticated: true,
    });

    vi.spyOn(DocumentsApi, 'getExpiringContracts').mockResolvedValue(mockContracts);
  });

  it('deve renderizar o botão de notificações com badge de contagem inicial', async () => {
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    expect(button).toBeInTheDocument();

    // Deve exibir o badge com a contagem de novidades + contratos calculados
    await waitFor(() => {
      // 3 do changelog + 2 dos contratos
      expect(screen.getByText('5')).toBeInTheDocument();
    });
  });

  it('deve abrir o popover e exibir o cabeçalho e abas ao clicar no sino', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    expect(screen.getByText('Notificações')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^Todas/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Vencimentos/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Novidades ✨/i })).toBeInTheDocument();
  });

  it('deve exibir alertas de contratos a vencer com numeração e tags de urgência', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    // Deve exibir os títulos dos contratos a vencer
    await waitFor(() => {
      expect(screen.getByText(/Contrato #101/i)).toBeInTheDocument();
      expect(screen.getByText(/Contrato #102/i)).toBeInTheDocument();
    });
  });

  it('deve alternar para a aba "Novidades ✨" e exibir as notas de versão do sistema', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    const changelogTab = screen.getByRole('button', { name: /Novidades ✨/i });
    await user.click(changelogTab);

    expect(screen.getByText('Central de Notificações & Alertas de Vigência')).toBeInTheDocument();
    expect(screen.getByText('Gestão Completa de Contratos Públicos')).toBeInTheDocument();
    expect(screen.getByText('Lançamento do Docseq')).toBeInTheDocument();
  });

  it('deve marcar todas como lidas ao clicar em "Ler todas"', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    const markAllBtn = await screen.findByTitle('Marcar todas como lidas');
    expect(markAllBtn).toBeInTheDocument();

    await user.click(markAllBtn);

    // O badge deve desaparecer
    await waitFor(() => {
      expect(screen.queryByText('5')).not.toBeInTheDocument();
    });
  });
});
