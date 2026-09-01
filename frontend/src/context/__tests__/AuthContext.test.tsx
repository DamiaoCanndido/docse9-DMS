import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { AuthProvider, useAuth } from '../AuthContext'
import * as authApi from '@/app/api/auth'
import { User } from '@/types'

const mockPush = vi.fn()
const mockRefresh = vi.fn()

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    refresh: mockRefresh,
  }),
}))

vi.mock('@/app/api/auth', () => ({
  loginUser: vi.fn(),
  logoutUser: vi.fn(),
  changePassword: vi.fn(),
}))

describe('AuthContext', () => {
  const mockUser: User = {
    id: 'user-1',
    username: 'admin',
    email: 'admin@example.com',
    role: 'ADMIN',
    municipalityId: 'mun-1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('deve inicializar com o usuário padrão ou nulo', () => {
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider initialUser={mockUser}>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    expect(result.current.user).toEqual(mockUser)
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.loading).toBe(false)
  })

  it('deve realizar login com sucesso e atualizar o usuário', async () => {
    vi.mocked(authApi.loginUser).mockResolvedValueOnce({
      success: true,
      token: 'token-123',
      user: mockUser,
    })

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await act(async () => {
      await result.current.login('admin', 'password')
    })

    expect(result.current.user).toEqual(mockUser)
    expect(result.current.isAuthenticated).toBe(true)
  })

  it('deve realizar logout com sucesso e limpar o estado do usuário', async () => {
    vi.mocked(authApi.logoutUser).mockResolvedValueOnce(undefined)

    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <AuthProvider initialUser={mockUser}>{children}</AuthProvider>
    )

    const { result } = renderHook(() => useAuth(), { wrapper })

    await act(async () => {
      await result.current.logout()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(mockPush).toHaveBeenCalledWith('/login')
  })

  it('deve lançar erro se useAuth for chamado fora do AuthProvider', () => {
    expect(() => renderHook(() => useAuth())).toThrow(
      'useAuth deve ser utilizado dentro de um AuthProvider'
    )
  })
})
