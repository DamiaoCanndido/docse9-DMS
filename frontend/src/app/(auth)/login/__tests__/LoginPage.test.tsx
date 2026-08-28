import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import LoginPage from '../page'
import * as AuthContext from '@/context/AuthContext'

const mockPush = vi.fn()
const mockRefresh = vi.fn()

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    refresh: mockRefresh,
  }),
}))

vi.mock('next/image', () => ({
  default: ({ src, alt, ...rest }: React.ImgHTMLAttributes<HTMLImageElement> & { priority?: boolean }) => (
    <img src={typeof src === 'string' ? src : ''} alt={alt || ''} {...rest} />
  ),
}))

describe('LoginPage Component', () => {
  const mockLogin = vi.fn()

  const renderComponent = () => {
    vi.spyOn(AuthContext, 'useAuth').mockReturnValue({
      user: null,
      loading: false,
      login: mockLogin,
      logout: vi.fn(),
      changePassword: vi.fn(),
      updateUser: vi.fn(),
      isAuthenticated: false,
    })

    return render(<LoginPage />)
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('deve renderizar os campos de login e botão de acesso', () => {
    renderComponent()

    expect(screen.getByText('docseq')).toBeInTheDocument()
    expect(screen.getByLabelText('Usuário')).toBeInTheDocument()
    expect(screen.getByLabelText('Senha')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Acessar Sistema/i })).toBeInTheDocument()
  })

  it('deve exibir erros de validação com campos vazios ou curtos', async () => {
    const user = userEvent.setup()
    renderComponent()

    await user.click(screen.getByRole('button', { name: /Acessar Sistema/i }))

    await waitFor(() => {
      expect(screen.getByText('O usuário deve ter pelo menos 3 caracteres')).toBeInTheDocument()
      expect(screen.getByText('A senha deve ter pelo menos 6 caracteres')).toBeInTheDocument()
    })
    expect(mockLogin).not.toHaveBeenCalled()
  })

  it('deve chamar login e redirecionar em caso de sucesso', async () => {
    const user = userEvent.setup()
    mockLogin.mockResolvedValueOnce(undefined)
    renderComponent()

    await user.type(screen.getByLabelText('Usuário'), 'admin')
    await user.type(screen.getByLabelText('Senha'), 'password123')
    await user.click(screen.getByRole('button', { name: /Acessar Sistema/i }))

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('admin', 'password123')
      expect(mockRefresh).toHaveBeenCalled()
      expect(mockPush).toHaveBeenCalledWith('/')
    })
  })

  it('deve exibir mensagem de erro quando o login falhar', async () => {
    const user = userEvent.setup()
    mockLogin.mockRejectedValueOnce(new Error('Credenciais inválidas'))
    renderComponent()

    await user.type(screen.getByLabelText('Usuário'), 'admin')
    await user.type(screen.getByLabelText('Senha'), 'wrongpass')
    await user.click(screen.getByRole('button', { name: /Acessar Sistema/i }))


    await waitFor(() => {
      expect(screen.getByText('Credenciais inválidas')).toBeInTheDocument()
    })
    expect(mockPush).not.toHaveBeenCalled()
  })
})
