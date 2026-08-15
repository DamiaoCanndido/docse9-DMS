'use server';

import { cookies } from 'next/headers';
import { apiServer } from '@/lib/axios';
import { parseStringify } from '@/lib/utils';
import { redirect } from 'next/navigation';
import axios from 'axios';
import { User, LoginResponse } from '@/types';

export async function getToken() {
  const cookieStore = await cookies();
  return cookieStore.get('token')?.value;
}

export async function getMe(): Promise<User> {
  try {
    const cookieStore = await cookies();
    const userCookie = cookieStore.get('user')?.value;
    if (!userCookie) {
      return redirect('/login');
    }
    return parseStringify(JSON.parse(userCookie));
  } catch {
    return redirect('/login');
  }
}

export async function getUserOrNull(): Promise<User | null> {
  try {
    const cookieStore = await cookies();
    const userCookie = cookieStore.get('user')?.value;
    if (!userCookie) return null;
    return parseStringify(JSON.parse(userCookie));
  } catch {
    return null;
  }
}

export type LoginActionResult = 
  | { success: true; token: string; user: User }
  | { success: false; error: string };

export async function loginUser({ form }: { form: { username: string; password: string } }): Promise<LoginActionResult> {
  try {
    const response = await apiServer.post<{ success: boolean; data: LoginResponse }>('/auth/login', {
      username: form.username,
      password: form.password,
    });

    const loginData = response.data.data;

    if (!loginData || !loginData.token) {
      return { success: false, error: 'Resposta inválida do servidor.' };
    }

    const cookieStore = await cookies();
    
    // Set token cookie
    cookieStore.set('token', loginData.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 24 * 60 * 60, // 1 day
      path: '/',
    });

    // Set user cookie
    cookieStore.set('user', JSON.stringify(loginData.user), {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 24 * 60 * 60, // 1 day
      path: '/',
    });

    return {
      success: true,
      token: loginData.token,
      user: parseStringify(loginData.user),
    };
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const msg = error.response?.data?.error || 'Credenciais inválidas.';
      return { success: false, error: msg };
    }
    return { success: false, error: 'Falha na conexão com o servidor de API.' };
  }
}

export async function logoutUser() {
  try {
    const cookieStore = await cookies();
    cookieStore.delete('token');
    cookieStore.delete('user');
  } catch {
    throw new Error('Falha ao deslogar o usuário.');
  }
}

export async function changePassword(form: { currentPassword: string; newPassword: string; confirmPassword: string }) {
  try {
    const token = await getToken();
    if (!token) {
      return { success: false, error: 'Usuário não autenticado.' };
    }

    const response = await apiServer.post<{ success: boolean; data: { token: string; user: User } }>(
      '/users/me/change-password',
      form,
      {
        headers: { Authorization: `Bearer ${token}` },
      }
    );

    const data = response.data.data;
    if (!data || !data.token) {
      return { success: false, error: 'Resposta inválida do servidor.' };
    }

    const cookieStore = await cookies();
    
    // Update token cookie
    cookieStore.set('token', data.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 24 * 60 * 60, // 1 day
      path: '/',
    });

    // Update user cookie
    cookieStore.set('user', JSON.stringify(data.user), {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 24 * 60 * 60, // 1 day
      path: '/',
    });

    return { success: true, user: parseStringify(data.user) };
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const msg = error.response?.data?.error || 'Senha atual incorreta ou dados inválidos.';
      return { success: false, error: msg };
    }
    return { success: false, error: 'Falha na conexão com o servidor de API.' };
  }
}

export async function updateUserProfile(form: { username: string; email: string }): Promise<{ success: boolean; user?: User; error?: string }> {
  try {
    const token = await getToken();
    const currentUser = await getUserOrNull();
    if (!token || !currentUser) {
      return { success: false, error: 'Usuário não autenticado.' };
    }

    const response = await apiServer.patch<{ success: boolean; data: User | { user: User } }>(
      `/users/${currentUser.id}`,
      {
        username: form.username,
        email: form.email,
      },
      {
        headers: { Authorization: `Bearer ${token}` },
      }
    );

    const rawData = response.data.data;
    const updatedUser: User = (rawData && typeof rawData === 'object' && 'user' in rawData)
      ? (rawData as { user: User }).user
      : (rawData as User);

    const cookieStore = await cookies();
    cookieStore.set('user', JSON.stringify(updatedUser), {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 24 * 60 * 60, // 1 day
      path: '/',
    });

    return { success: true, user: parseStringify(updatedUser) };
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const msg = error.response?.data?.error || 'Erro ao atualizar dados do perfil.';
      return { success: false, error: msg };
    }
    return { success: false, error: 'Falha na conexão com o servidor de API.' };
  }
}
