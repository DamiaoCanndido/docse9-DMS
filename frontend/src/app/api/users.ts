'use server';

import { apiServer } from '@/lib/axios';
import { getToken, getUserOrNull, logoutUser } from './auth';
import { parseStringify } from '@/lib/utils';
import { revalidatePath } from 'next/cache';
import { redirect } from 'next/navigation';
import axios from 'axios';
import { User, UserPermission, DocumentType, PermissionLevel, PaginatedResponse } from '@/types';

async function getAuthHeader() {
  const token = await getToken();
  if (!token) {
    return {};
  }
  return { Authorization: `Bearer ${token}` };
}

export async function getUsers(
  page = 1,
  pageSize = 10,
  filter?: { municipalityId?: string; role?: string; search?: string }
): Promise<PaginatedResponse<User>> {
  try {
    const headers = await getAuthHeader();
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());

    if (filter?.municipalityId) params.append('municipalityId', filter.municipalityId);
    if (filter?.role) params.append('role', filter.role);
    if (filter?.search) params.append('search', filter.search);

    const response = await apiServer.get<{
      success: boolean;
      data: User[];
      pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
      };
    }>('/users', {
      headers,
      params,
    });
    
    const body = response.data;
    return {
      data: parseStringify(body.data || []),
      total: body.pagination?.total || 0,
      page: body.pagination?.page || 1,
      pageSize: body.pagination?.pageSize || 10,
    };
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function getUsersTrash(page = 1, pageSize = 10): Promise<PaginatedResponse<User>> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<{
      success: boolean;
      data: User[];
      pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
      };
    }>('/users/trash', {
      headers,
      params: { page, pageSize },
    });
    
    const body = response.data;
    return {
      data: parseStringify(body.data || []),
      total: body.pagination?.total || 0,
      page: body.pagination?.page || 1,
      pageSize: body.pagination?.pageSize || 10,
    };
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function getUserByID(id: string): Promise<User> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<{ success: boolean; data: User }>(`/users/${id}`, { headers });
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export interface CreateUserResponse {
  user: User;
  randomPassword?: string;
}

export async function createUser({
  input,
  path,
}: {
  input: { username: string; email: string; password?: string; role: string; municipalityId: string };
  path: string;
}): Promise<CreateUserResponse> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.post<{ success: boolean; data: User | CreateUserResponse }>('/users', input, { headers });
    revalidatePath(path);
    const data = parseStringify(response.data.data);
    if (data && typeof data === 'object' && 'user' in data) {
      return data as CreateUserResponse;
    }
    return { user: data as User };
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function updateUser({
  id,
  input,
  path,
}: {
  id: string;
  input: { username?: string; email?: string; password?: string; role?: string; municipalityId?: string };
  path: string;
}): Promise<{ user: User; randomPassword?: string }> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.patch<{ success: boolean; data: User | { user: User; randomPassword?: string } }>(
      `/users/${id}`,
      input,
      { headers }
    );
    
    // Se o usuário atualizou a própria senha, desloga e redireciona
    if (input.password && input.password.trim() !== '') {
      const currentUser = await getUserOrNull();
      if (currentUser && currentUser.id === id) {
        await logoutUser();
        return redirect('/login');
      }
    }

    revalidatePath(path);
    const data = parseStringify(response.data.data);
    if (data && typeof data === 'object' && 'user' in data) {
      return data as { user: User; randomPassword?: string };
    }
    return { user: data as User };
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function deleteUser({ id, path }: { id: string; path: string }): Promise<void> {
  try {
    const headers = await getAuthHeader();
    await apiServer.delete(`/users/${id}`, { headers });
    revalidatePath(path);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function restoreUser({ id, path }: { id: string; path: string }): Promise<User> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.patch<{ success: boolean; data: User }>(`/users/${id}/restore`, {}, { headers });
    revalidatePath(path);
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function hardDeleteUser({ id, path }: { id: string; path: string }): Promise<void> {
  try {
    const headers = await getAuthHeader();
    await apiServer.delete(`/users/${id}/hard`, { headers });
    revalidatePath(path);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function getUserPermissions(id: string): Promise<UserPermission[]> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<UserPermission[] | { success: boolean; data: UserPermission[] }>(
      `/users/${id}/permissions`,
      { headers }
    );
    // Handle wrapped or unwrapped response
    const data = Array.isArray(response.data) ? response.data : response.data.data;
    return parseStringify(data || []);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function updateUserPermissions({
  id,
  permissions,
  path,
}: {
  id: string;
  permissions: { documentType: DocumentType; level: PermissionLevel }[];
  path: string;
}): Promise<UserPermission[]> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.put<UserPermission[] | { success: boolean; data: UserPermission[] }>(
      `/users/${id}/permissions`,
      { permissions },
      { headers }
    );
    revalidatePath(path);
    const data = Array.isArray(response.data) ? response.data : response.data.data;
    return parseStringify(data || []);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}
