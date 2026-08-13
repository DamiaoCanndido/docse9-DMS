import React from 'react';
import { getDocuments, getDocumentsTrash } from '@/app/api/documents';
import { getMe } from '@/app/api/auth';
import { DocumentsContent } from '@/components/DocumentsContent';
import { DocumentType, ContractType } from '@/types';
import { redirect } from 'next/navigation';

interface PageProps {
  searchParams: Promise<{
    search?: string;
    type?: string;
    contractType?: string;
    year?: string;
    page?: string;
    pageSize?: string;
    trash?: string;
  }>;
}

export default async function DashboardPage({ searchParams }: PageProps) {
  // Resolve params
  const params = await searchParams;
  const search = params.search || '';
  const type = (params.type as DocumentType) || 'NOTICE';
  const contractType = (params.contractType as ContractType) || undefined;
  
  const currentYear = new Date().getFullYear();
  const yearParam = params.year;
  const year = yearParam === 'all' ? undefined : (yearParam ? Number(yearParam) : currentYear);

  const page = Number(params.page) || 1;
  const pageSize = Number(params.pageSize) || 10;
  const viewTrash = params.trash === 'true';

  const user = await getMe();

  // Redireciona admins para a tela de municípios, pois eles não gerenciam documentos
  if (user.role === 'ADMIN') {
    return redirect('/municipalities');
  }

  // Busca os documentos (ativos ou excluídos na lixeira)
  const documentsData = viewTrash
    ? await getDocumentsTrash({ search, type, contractType, year }, page, pageSize)
    : await getDocuments({ search, type, contractType, year }, page, pageSize);

  return (
    <DocumentsContent
      initialData={documentsData}
      currentUser={user}
      viewTrash={viewTrash}
    />
  );
}


