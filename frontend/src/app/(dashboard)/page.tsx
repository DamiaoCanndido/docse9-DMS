import React from 'react';
import { getDocuments } from '@/app/api/documents';
import { getMe } from '@/app/api/auth';
import { DocumentsContent } from '@/components/DocumentsContent';
import { DocumentType, ContractType } from '@/types';

interface PageProps {
  searchParams: Promise<{
    search?: string;
    type?: string;
    contractType?: string;
    page?: string;
  }>;
}

export default async function DashboardPage({ searchParams }: PageProps) {
  // Resolve params (necessário no Next.js 15+)
  const params = await searchParams;
  const search = params.search || '';
  const type = (params.type as DocumentType) || undefined;
  const contractType = (params.contractType as ContractType) || undefined;
  const page = Number(params.page) || 1;

  // Busca do usuário logado e dados de documentos executadas de forma paralela no servidor (SSR)
  const [user, documentsData] = await Promise.all([
    getMe(),
    getDocuments({ search, type, contractType }, page, 10),
  ]);

  return (
    <DocumentsContent
      initialData={documentsData}
      currentUser={user}
    />
  );
}
