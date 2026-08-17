export type Role = 'ADMIN' | 'MOD' | 'COMMON';

export interface Municipality {
  id: string;
  name: string;
  uf: string;
  imageUrl?: string;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  username: string;
  email: string;
  role: Role;
  municipalityId: string;
  municipality?: Municipality;
  mustChangePassword?: boolean;
  lastLogin?: string;
  createdAt: string;
  updatedAt: string;
}

export type DocumentType = 'NOTICE' | 'DECREE' | 'ORDINANCE' | 'LAW' | 'CONTRACT';
export type ContractType = 'publicinterest' | 'bidding' | 'service';

export interface Document {
  id: string;
  type: DocumentType;
  order: number;
  description: string;
  fileKey: string;
  creatorId: string;
  createdBy?: User;
  municipalityId: string;
  municipality?: Municipality;
  
  // Specific fields for CONTRACT type
  duration?: number;
  contractType?: ContractType;
  value?: number;
  startIn?: string;

  createdAt: string;
  updatedAt: string;
}

export type PermissionLevel = 'NONE' | 'READ' | 'WRITE' | 'DELETE';

export interface UserPermission {
  id: string;
  userId: string;
  documentType: DocumentType;
  level: PermissionLevel;
  createdAt: string;
  updatedAt: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface CreateDocumentInput {
  type: DocumentType;
  description: string;
  creatorId: string;
  municipalityId: string;
  duration?: number;
  contractType?: ContractType;
  value?: number;
  startIn?: string;
}

export interface UpdateDocumentInput {
  description?: string;
  fileKey?: string;
  createdAt?: string;
  duration?: number;
  contractType?: ContractType;
  value?: number;
  startIn?: string;
}

export interface DocumentFilter {
  type?: DocumentType;
  contractType?: ContractType;
  search?: string;
  year?: number;
  allowedTypes?: DocumentType[];
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

