import api from './api';
import { Document, CreateDocumentInput, UpdateDocumentInput, DocumentFilter, PaginatedResponse } from '../types';

export const documentService = {
  getAll: async (filter: DocumentFilter, page = 1, pageSize = 10): Promise<PaginatedResponse<Document>> => {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('pageSize', pageSize.toString());

    if (filter.type) params.append('type', filter.type);
    if (filter.contractType) params.append('contractType', filter.contractType);
    if (filter.search) params.append('search', filter.search);
    if (filter.allowedTypes) {
      filter.allowedTypes.forEach(t => params.append('allowedTypes', t));
    }

    const response = await api.get<PaginatedResponse<Document>>('/documents', { params });
    return response.data;
  },

  getByID: async (id: string): Promise<Document> => {
    const response = await api.get<Document>(`/documents/${id}`);
    return response.data;
  },

  create: async (input: CreateDocumentInput): Promise<Document> => {
    const response = await api.post<Document>('/documents', input);
    return response.data;
  },

  update: async (id: string, input: UpdateDocumentInput): Promise<Document> => {
    const response = await api.patch<Document>(`/documents/${id}`, input);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await api.delete(`/documents/${id}`);
  },

  restore: async (id: string): Promise<Document> => {
    const response = await api.patch<Document>(`/documents/${id}/restore`);
    return response.data;
  },
};
export default documentService;
