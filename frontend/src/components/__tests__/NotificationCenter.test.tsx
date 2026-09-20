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
  today.setHours(0, 0, 0, 0);

  // Helper para criar data de início que resulta exatamente em X dias a partir de hoje
  const createContractStartDate = (daysFromToday: number, durationMonths: number = 1): string => {
    const start = new Date(today);
    start.setMonth(start.getMonth() - durationMonths);
    start.setDate(start.getDate() + daysFromToday);
    return start.toISOString();
  };

  // Contrato 1: Vence em 4 dias (dentro da janela de 7 dias) -> Deve aparecer
  const expiring4DaysStart = createContractStartDate(4);

  // Contrato 2: Expirou há 3 dias (dentro da janela de 7 dias) -> Deve aparecer
  const expired3DaysStart = createContractStartDate(-3);

  // Contrato 3: Vence em 20 dias (fora da janela de 7 dias) -> NÃO deve aparecer
  const expiring20DaysStart = createContractStartDate(20);

  // Contrato 4: Expirou há 15 dias (fora da janela de 7 dias) -> NÃO deve aparecer
  const expired15DaysStart = createContractStartDate(-15);

  // Contrato 5: Contrato antigo de ano anterior -> Deve ser ignorado
  const pastYearStartDate = new Date(today);
  pastYearStartDate.setFullYear(pastYearStartDate.getFullYear() - 2);

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
      duration: 1,
      startIn: expiring4DaysStart,
      createdAt: expiring4DaysStart,
      updatedAt: expiring4DaysStart,
    },
    {
      id: 'contract-expired-2',
      type: 'CONTRACT',
      order: 102,
      description: 'Contrato de Manutenção de Software',
      fileKey: '',
      creatorId: mockUser.id,
      municipalityId: mockUser.municipalityId,
      contractType: 'bidding',
      value: 45000,
      duration: 1,
      startIn: expired3DaysStart,
      createdAt: expired3DaysStart,
      updatedAt: expired3DaysStart,
    },
    {
      id: 'contract-far-future-3',
      type: 'CONTRACT',
      order: 103,
      description: 'Contrato Vencendo em 20 dias',
      fileKey: '',
      creatorId: mockUser.id,
      municipalityId: mockUser.municipalityId,
      contractType: 'service',
      value: 80000,
      duration: 1,
      startIn: expiring20DaysStart,
      createdAt: expiring20DaysStart,
      updatedAt: expiring20DaysStart,
    },
    {
      id: 'contract-far-past-4',
      type: 'CONTRACT',
      order: 104,
      description: 'Contrato Expirado há 15 dias',
      fileKey: '',
      creatorId: mockUser.id,
      municipalityId: mockUser.municipalityId,
      contractType: 'service',
      value: 50000,
      duration: 1,
      startIn: expired15DaysStart,
      createdAt: expired15DaysStart,
      updatedAt: expired15DaysStart,
    },
    {
      id: 'contract-past-year-5',
      type: 'CONTRACT',
      order: 99,
      description: 'Contrato Antigo de Ano Anterior',
      fileKey: '',
      creatorId: mockUser.id,
      municipalityId: mockUser.municipalityId,
      contractType: 'service',
      value: 30000,
      duration: 12,
      startIn: pastYearStartDate.toISOString(),
      createdAt: pastYearStartDate.toISOString(),
      updatedAt: pastYearStartDate.toISOString(),
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

    // Deve exibir o badge com a contagem: 5 novidades do changelog + 2 contratos elegíveis (janela de 7 dias) = 7
    await waitFor(() => {
      expect(screen.getByText('7')).toBeInTheDocument();
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

  it('deve exibir apenas contratos dentro da janela de 7 dias e ignorar os demais', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    // Contratos na janela de 7 dias devem aparecer
    await waitFor(() => {
      expect(screen.getByText(/Contrato #101/i)).toBeInTheDocument();
      expect(screen.getByText(/Contrato #102/i)).toBeInTheDocument();
    });

    // Contratos fora da janela de 7 dias não devem estar no feed
    expect(screen.queryByText(/Contrato #103/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/Contrato #104/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/Contrato #99/i)).not.toBeInTheDocument();
  });

  it('deve alternar para a aba "Novidades ✨" e exibir as notas de versão do sistema', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    const changelogTab = screen.getByRole('button', { name: /Novidades ✨/i });
    await user.click(changelogTab);

    expect(screen.getByText('Sessão Estendida de 7 Dias')).toBeInTheDocument();
    expect(screen.getByText('Central de Notificações & Alertas de Vigência')).toBeInTheDocument();
    expect(screen.getByText('Gestão Completa de Contratos Públicos')).toBeInTheDocument();
    expect(screen.getByText('Lançamento do Docseq')).toBeInTheDocument();
  });

  it('deve remover novidade do feed assim que for marcada como lida individualmente', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    const changelogTab = screen.getByRole('button', { name: /Novidades ✨/i });
    await user.click(changelogTab);

    // A novidade está visível
    const changelogItem = screen.getByText('Central de Notificações & Alertas de Vigência');
    expect(changelogItem).toBeInTheDocument();

    // Clica no botão de marcar como lida da primeira novidade
    const markReadBtn = screen.getByRole('button', { name: /Marcar "Central de Notificações & Alertas de Vigência" como lida/i });
    await user.click(markReadBtn);

    // Deve ser imediatamente removida do feed
    await waitFor(() => {
      expect(screen.queryByText('Central de Notificações & Alertas de Vigência')).not.toBeInTheDocument();
    });

    // As outras novidades não lidas continuam visíveis
    expect(screen.getByText('Gestão Completa de Contratos Públicos')).toBeInTheDocument();
    expect(screen.getByText('Lançamento do Docseq')).toBeInTheDocument();
  });

  it('deve remover todas as novidades do feed e desativar badge ao clicar em "Ler todas"', async () => {
    const user = userEvent.setup();
    render(<NotificationCenter />);

    const button = screen.getByRole('button', { name: /Notificações e Avisos de Vencimento/i });
    await user.click(button);

    const markAllBtn = await screen.findByTitle('Marcar todas como lidas');
    expect(markAllBtn).toBeInTheDocument();

    await user.click(markAllBtn);

    // O badge de contagem desaparece
    await waitFor(() => {
      expect(screen.queryByText('5')).not.toBeInTheDocument();
    });

    // Ao ir para a aba de novidades, todas as novidades foram removidas do feed
    const changelogTab = screen.getByRole('button', { name: /Novidades ✨/i });
    await user.click(changelogTab);

    expect(screen.queryByText('Central de Notificações & Alertas de Vigência')).not.toBeInTheDocument();
    expect(screen.queryByText('Gestão Completa de Contratos Públicos')).not.toBeInTheDocument();
    expect(screen.queryByText('Lançamento do Docseq')).not.toBeInTheDocument();
    expect(screen.getByText('Tudo em dia por aqui!')).toBeInTheDocument();
  });
});
