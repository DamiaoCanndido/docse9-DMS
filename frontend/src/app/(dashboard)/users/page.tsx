import React from 'react';
import { getUsers, getUsersTrash } from '@/app/api/users';
import { getMunicipalities } from '@/app/api/municipalities';
import { getMe } from '@/app/api/auth';
import { UsersContent } from '@/components/UsersContent';
import { redirect } from 'next/navigation';

interface PageProps {
  searchParams: Promise<{
    page?: string;
    pageSize?: string;
    trash?: string;
    role?: string;
    municipalityId?: string;
    search?: string;
  }>;
}

export default async function UsersPage({ searchParams }: PageProps) {
  const params = await searchParams;
  const page = Number(params.page) || 1;
  const pageSize = Number(params.pageSize) || 10;
  const viewTrash = params.trash === 'true';
  const roleFilter = params.role || undefined;
  const municipalityIdFilter = params.municipalityId || undefined;
  const searchFilter = params.search || undefined;

  const user = await getMe();
  
  // Apenas ADMIN e MOD podem acessar gerenciamento de usuários
  if (user.role !== 'ADMIN' && user.role !== 'MOD') {
    return redirect('/');
  }

  // Se for MOD, filtra automaticamente apenas os usuários do próprio município
  const finalMunicipalityId = user.role === 'MOD' ? user.municipalityId : municipalityIdFilter;

  const userFilter = {
    municipalityId: finalMunicipalityId,
    role: roleFilter,
    search: searchFilter,
  };

  // Busca lista de usuários
  const usersData = viewTrash 
    ? await getUsersTrash(page, pageSize, userFilter)
    : await getUsers(page, pageSize, userFilter);

  // Busca municípios para o dropdown de criação/filtragem (apenas ADMIN precisa da lista completa)
  const municipalitiesData = user.role === 'ADMIN' 
    ? await getMunicipalities(1, 100)
    : null;

  return (
    <UsersContent
      initialData={usersData}
      currentUser={user}
      viewTrash={viewTrash}
      municipalities={municipalitiesData?.data || []}
    />
  );
}
