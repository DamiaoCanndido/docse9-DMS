'use server';

import { apiServer } from '@/lib/axios';
import { getToken } from './auth';
import { parseStringify } from '@/lib/utils';
import { revalidatePath } from 'next/cache';
import { redirect } from 'next/navigation';
import axios from 'axios';
import { Document, CreateDocumentInput, UpdateDocumentInput, DocumentFilter, PaginatedResponse } from '@/types';

async function getAuthHeader() {
  const token = await getToken();
  if (!token) {
    return {};
  }
  return { Authorization: `Bearer ${token}` };
}

export async function getDocuments(filter: DocumentFilter, page = 1, pageSize = 10): Promise<PaginatedResponse<Document>> {
  try {
    const headers = await getAuthHeader();
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());

    if (filter.type) params.append('type', filter.type);
    if (filter.contractType) params.append('contractType', filter.contractType);
    if (filter.search) params.append('search', filter.search);
    if (filter.allowedTypes) {
      filter.allowedTypes.forEach(t => params.append('allowedTypes', t));
    }

    const response = await apiServer.get<{
      success: boolean;
      data: Document[];
      pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
      };
    }>('/documents', {
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

export async function getDocumentByID(id: string): Promise<Document> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<{ success: boolean; data: Document }>(`/documents/${id}`, { headers });
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function createDocument({ input, path }: { input: CreateDocumentInput; path: string }): Promise<Document> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.post<{ success: boolean; data: Document }>('/documents', input, { headers });
    revalidatePath(path);
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function updateDocument({ id, input, path }: { id: string; input: UpdateDocumentInput; path: string }): Promise<Document> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.patch<{ success: boolean; data: Document }>(`/documents/${id}`, input, { headers });
    revalidatePath(path);
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function deleteDocument({ id, path }: { id: string; path: string }): Promise<void> {
  try {
    const headers = await getAuthHeader();
    await apiServer.delete(`/documents/${id}`, { headers });
    revalidatePath(path);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function restoreDocument({ id, path }: { id: string; path: string }): Promise<Document> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.patch<{ success: boolean; data: Document }>(`/documents/${id}/restore`, {}, { headers });
    revalidatePath(path);
    return parseStringify(response.data.data);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

export async function getDocumentsTrash(page = 1, pageSize = 10): Promise<PaginatedResponse<Document>> {
  try {
    const headers = await getAuthHeader();
    const response = await apiServer.get<{
      success: boolean;
      data: Document[];
      pagination: {
        page: number;
        pageSize: number;
        total: number;
        totalPages: number;
      };
    }>('/documents/trash', {
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

export async function hardDeleteDocument({ id, path }: { id: string; path: string }): Promise<void> {
  try {
    const headers = await getAuthHeader();
    await apiServer.delete(`/documents/${id}/hard`, { headers });
    revalidatePath(path);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return redirect('/login');
    }
    throw error;
  }
}

