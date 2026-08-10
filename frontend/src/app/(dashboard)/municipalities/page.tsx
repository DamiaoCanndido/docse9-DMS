import React from 'react';
import { getMunicipalities, getMunicipalitiesTrash } from '@/app/api/municipalities';
import { getMe } from '@/app/api/auth';
import { MunicipalitiesContent } from '@/components/MunicipalitiesContent';
import { redirect } from 'next/navigation';

interface PageProps {
  searchParams: Promise<{
    page?: string;
    trash?: string;
  }>;
}

export default async function MunicipalitiesPage({ searchParams }: PageProps) {
  const params = await searchParams;
  const page = Number(params.page) || 1;
  const viewTrash = params.trash === 'true';

  const user = await getMe();
  
  // Apenas ADMIN pode gerenciar municípios
  if (user.role !== 'ADMIN') {
    return redirect('/');
  }

  // Busca dados dependendo se o usuário está visualizando ativos ou excluídos (lixeira)
  const municipalitiesData = viewTrash 
    ? await getMunicipalitiesTrash(page, 10)
    : await getMunicipalities(page, 10);

  return (
    <MunicipalitiesContent
      initialData={municipalitiesData}
      currentUser={user}
      viewTrash={viewTrash}
    />
  );
}
