'use server';

import { apiServer } from '@/lib/axios';
import { getToken } from './auth';
import { parseStringify } from '@/lib/utils';
import { revalidatePath } from 'next/cache';
import { redirect } from 'next/navigation';
import axios from 'axios';
import { Municipality, PaginatedResponse } from '@/types';

async function getAuthHeader() {
  const token = await getToken();
  if (!token) {
    return {};
  }
  return { Authorization: `Bearer ${token}` };
}

export async function getMunicipalities(page = 1, pageSize = 10): Promise<PaginatedResponse<Municipality>> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<{
      success: boolean;
      data: Municipality[];
      pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
      };
    }>('/municipalities', {
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
    if (axios.isAxiosError(error)) {
      if (error.response?.status === 401) {
        return redirect('/login');
      }
      if (error.response?.status === 403) {
        return {
          data: [],
          total: 0,
          page: 1,
          pageSize: 10,
        };
      }
    }
    throw error;
  }
}

export async function getMunicipalitiesTrash(page = 1, pageSize = 10): Promise<PaginatedResponse<Municipality>> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<{
      success: boolean;
      data: Municipality[];
      pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
      };
    }>('/municipalities/trash', {
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
    if (axios.isAxiosError(error)) {
      if (error.response?.status === 401) {
        return redirect('/login');
      }
      if (error.response?.status === 403) {
        return {
          data: [],
          total: 0,
          page: 1,
          pageSize: 10,
        };
      }
    }
    throw error;
  }
}

export async function getMunicipalityByID(id: string): Promise<Municipality> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<{ success: boolean; data: Municipality }>(`/municipalities/${id}`, { headers });
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function createMunicipality({ input, path }: { input: { name: string; uf: string; imageUrl?: string }; path: string }): Promise<Municipality> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.post<{ success: boolean; data: Municipality }>('/municipalities', input, { headers });
    revalidatePath(path);
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function updateMunicipality({ id, input, path }: { id: string; input: { name?: string; uf?: string; imageUrl?: string }; path: string }): Promise<Municipality> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.patch<{ success: boolean; data: Municipality }>(`/municipalities/${id}`, input, { headers });
    revalidatePath(path);
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function deleteMunicipality({ id, path }: { id: string; path: string }): Promise<void> {
  try {
    const headers = await getAuthHeader();
    await apiServer.delete(`/municipalities/${id}`, { headers });
    revalidatePath(path);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function restoreMunicipality({ id, path }: { id: string; path: string }): Promise<Municipality> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.patch<{ success: boolean; data: Municipality }>(`/municipalities/${id}/restore`, {}, { headers });
    revalidatePath(path);
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function hardDeleteMunicipality({ id, path }: { id: string; path: string }): Promise<void> {
  try {
    const headers = await getAuthHeader();
    await apiServer.delete(`/municipalities/${id}/hard`, { headers });
    revalidatePath(path);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}
