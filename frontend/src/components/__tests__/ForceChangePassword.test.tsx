import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ForceChangePassword } from '../ForceChangePassword'
import * as AuthContext from '@/context/AuthContext'
import { User } from '@/types'
import { toast } from 'sonner'

// Mocks
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    refresh: vi.fn(),
  }),
}))

vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

describe('ForceChangePassword Component', () => {
  const mockChangePassword = vi.fn()
  const mockLogout = vi.fn()

  const defaultUser: User = {
    id: 'user-123',
    username: 'testuser',
    email: 'test@example.com',
    role: 'COMMON',
    municipalityId: 'mun-123',
    mustChangePassword: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }

  const renderComponent = (user: User | null = defaultUser) => {
    vi.spyOn(AuthContext, 'useAuth').mockReturnValue({
      user,
      loading: false,
      login: vi.fn(),
      logout: mockLogout,
      changePassword: mockChangePassword,
      updateUser: vi.fn(),
      isAuthenticated: !!user,
    })

    return render(<ForceChangePassword />)
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('não deve renderizar nada se o usuário não tiver mustChangePassword ativo', () => {
    const normalUser = { ...defaultUser, mustChangePassword: false }
    const { container } = renderComponent(normalUser)
    expect(container).toBeEmptyDOMElement()
  })

  it('não deve renderizar nada se o usuário for nulo', () => {
    const { container } = renderComponent(null)
    expect(container).toBeEmptyDOMElement()
  })

  it('deve renderizar o modal de troca de senha quando mustChangePassword for true', () => {
    renderComponent()

    expect(screen.getByText('Troca de Senha Obrigatória')).toBeInTheDocument()
    expect(screen.getByLabelText(/Senha Temporária/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^Nova Senha/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/Confirmar Nova Senha/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Atualizar Senha/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Sair/i })).toBeInTheDocument()
  })

  it('deve exibir erro de validação quando as senhas não coincidirem', async () => {
    const user = userEvent.setup()
    renderComponent()

    await user.type(screen.getByLabelText(/Senha Temporária/i), 'temp123')
    await user.type(screen.getByLabelText(/^Nova Senha/i), 'novaSenha123')
    await user.type(screen.getByLabelText(/Confirmar Nova Senha/i), 'outraSenha123')

    await user.click(screen.getByRole('button', { name: /Atualizar Senha/i }))

    await waitFor(() => {
      expect(screen.getByText('As novas senhas não coincidem')).toBeInTheDocument()
    })
    expect(mockChangePassword).not.toHaveBeenCalled()
  })

  it('deve chamar changePassword com sucesso e exibir toast', async () => {
    const user = userEvent.setup()
    mockChangePassword.mockResolvedValueOnce(undefined)
    renderComponent()

    await user.type(screen.getByLabelText(/Senha Temporária/i), 'temp123')
    await user.type(screen.getByLabelText(/^Nova Senha/i), 'novaSenha123')
    await user.type(screen.getByLabelText(/Confirmar Nova Senha/i), 'novaSenha123')

    await user.click(screen.getByRole('button', { name: /Atualizar Senha/i }))

    await waitFor(() => {
      expect(mockChangePassword).toHaveBeenCalledWith({
        currentPassword: 'temp123',
        newPassword: 'novaSenha123',
        confirmPassword: 'novaSenha123',
      })
      expect(toast.success).toHaveBeenCalledWith('Senha alterada com sucesso! Acesso liberado.')
    })
  })

  it('deve exibir mensagem de erro quando a API falhar', async () => {
    const user = userEvent.setup()
    mockChangePassword.mockRejectedValueOnce(new Error('A senha temporária está incorreta'))
    renderComponent()

    await user.type(screen.getByLabelText(/Senha Temporária/i), 'senhaErrada')
    await user.type(screen.getByLabelText(/^Nova Senha/i), 'novaSenha123')
    await user.type(screen.getByLabelText(/Confirmar Nova Senha/i), 'novaSenha123')

    await user.click(screen.getByRole('button', { name: /Atualizar Senha/i }))

    await waitFor(() => {
      expect(screen.getByText('A senha temporária está incorreta')).toBeInTheDocument()
      expect(toast.error).toHaveBeenCalledWith('Falha ao alterar senha.')
    })
  })

  it('deve chamar logout ao clicar no botão Sair', async () => {
    const user = userEvent.setup()
    renderComponent()

    await user.click(screen.getByRole('button', { name: /Sair/i }))
    expect(mockLogout).toHaveBeenCalled()
  })
})
