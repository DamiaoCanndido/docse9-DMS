import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Sidebar } from '../Sidebar'
import * as AuthContext from '@/context/AuthContext'
import { User } from '@/types'

vi.mock('next/navigation', () => ({
  usePathname: () => '/',
}))

vi.mock('next/image', () => ({
  default: ({ src, alt, ...rest }: React.ImgHTMLAttributes<HTMLImageElement> & { priority?: boolean }) => (
    <img src={typeof src === 'string' ? src : ''} alt={alt || ''} {...rest} />
  ),
}))

describe('Sidebar Component (RBAC Navigation)', () => {
  const mockLogout = vi.fn()

  const renderComponent = (role: 'ADMIN' | 'MOD' | 'COMMON') => {
    const user: User = {
      id: 'user-1',
      username: 'testuser',
      email: 'test@example.com',
      role,
      municipalityId: 'mun-1',
      municipality: {
        id: 'mun-1',
        name: 'Passagem',
        uf: 'PB',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    vi.spyOn(AuthContext, 'useAuth').mockReturnValue({
      user,
      loading: false,
      login: vi.fn(),
      logout: mockLogout,
      changePassword: vi.fn(),
      updateUser: vi.fn(),
      isAuthenticated: true,
    })

    return render(<Sidebar />)
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('deve exibir Documentos, Municípios e Usuários para usuário ADMIN', () => {
    renderComponent('ADMIN')

    expect(screen.getByText('Documentos Oficiais')).toBeInTheDocument()
    expect(screen.getByText('Municípios')).toBeInTheDocument()
    expect(screen.getByText('Usuários')).toBeInTheDocument()
    expect(screen.getByText('Meu Perfil')).toBeInTheDocument()
  })

  it('deve exibir Documentos e Usuários, mas NÃO Municípios para usuário MOD', () => {
    renderComponent('MOD')

    expect(screen.getByText('Documentos Oficiais')).toBeInTheDocument()
    expect(screen.getByText('Usuários')).toBeInTheDocument()
    expect(screen.queryByText('Municípios')).not.toBeInTheDocument()
    expect(screen.getByText('Meu Perfil')).toBeInTheDocument()
  })

  it('deve exibir APENAS Documentos e Perfil para usuário COMMON', () => {
    renderComponent('COMMON')

    expect(screen.getByText('Documentos Oficiais')).toBeInTheDocument()
    expect(screen.queryByText('Municípios')).not.toBeInTheDocument()
    expect(screen.queryByText('Usuários')).not.toBeInTheDocument()
    expect(screen.getByText('Meu Perfil')).toBeInTheDocument()
  })

  it('deve acionar logout ao clicar no botão Sair', async () => {
    const user = userEvent.setup()
    renderComponent('COMMON')

    await user.click(screen.getByRole('button', { name: /Sair do Sistema/i }))
    expect(mockLogout).toHaveBeenCalled()
  })
})
