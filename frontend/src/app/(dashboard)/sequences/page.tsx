import React from 'react';
import { getMe } from '@/app/api/auth';
import { getMunicipalitySequences } from '@/app/api/documents';
import { SequencesContent } from '@/components/SequencesContent';
import { SequenceItemResponse } from '@/types';
import { redirect } from 'next/navigation';

export default async function SequencesPage() {
  const user = await getMe();

  // Apenas MOD pode acessar configuração de sequências numéricas e marcos iniciais
  if (user.role !== 'MOD') {
    return redirect('/');
  }

  const currentYear = new Date().getFullYear();
  let initialSequences: SequenceItemResponse[] = [];
  try {
    initialSequences = await getMunicipalitySequences(user.municipalityId, currentYear);
  } catch {
    // Se a requisição inicial falhar no servidor, o cliente refaz a consulta
    initialSequences = [];
  }

  return (
    <SequencesContent
      currentUser={user}
      initialData={initialSequences}
    />
  );
}
