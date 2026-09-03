import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { WhatsNewAlertDialog } from '../WhatsNewAlertDialog';
import * as AuthContext from '@/context/AuthContext';
import { APP_VERSION } from '@/lib/version';
import { User } from '@/types';

describe('WhatsNewAlertDialog Component', () => {
  const mockUser: User = {
    id: 'user-whats-new-1',
    username: 'gestor.teste',
    email: 'gestor@docseq.local',
    role: 'MOD',
    municipalityId: 'mun-1',
    municipality: {
      id: 'mun-1',
      name: 'Passagem',
      uf: 'PB',
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    },
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  };

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
  });

  it('deve abrir o AlertDialog na primeira vez que o usuário acessa o dashboard', async () => {
    render(<WhatsNewAlertDialog />);

    await waitFor(() => {
      expect(screen.getByText('Bem-vindo às novidades do Docseq!')).toBeInTheDocument();
    });

    expect(screen.getByText('Central de Notificações no Topo')).toBeInTheDocument();
    expect(screen.getByText('Alertas de Vigência de Contratos')).toBeInTheDocument();
  });

  it('deve fechar o AlertDialog e salvar no localStorage ao clicar em "Entendi, vamos começar"', async () => {
    const user = userEvent.setup();
    render(<WhatsNewAlertDialog />);

    await waitFor(() => {
      expect(screen.getByText('Bem-vindo às novidades do Docseq!')).toBeInTheDocument();
    });

    const actionButton = screen.getByRole('button', { name: /Entendi, vamos começar/i });
    await user.click(actionButton);

    await waitFor(() => {
      expect(screen.queryByText('Bem-vindo às novidades do Docseq!')).not.toBeInTheDocument();
    });

    const storageKey = `docseq_whats_new_seen_${mockUser.id}_${APP_VERSION}`;
    expect(localStorage.getItem(storageKey)).toBe('true');
  });

  it('NÃO deve abrir o AlertDialog se o usuário já tiver visto a versão atual', async () => {
    const storageKey = `docseq_whats_new_seen_${mockUser.id}_${APP_VERSION}`;
    localStorage.setItem(storageKey, 'true');

    render(<WhatsNewAlertDialog />);

    // Aguarda o timeout inicial
    await new Promise((r) => setTimeout(r, 500));

    expect(screen.queryByText('Bem-vindo às novidades do Docseq!')).not.toBeInTheDocument();
  });
});
