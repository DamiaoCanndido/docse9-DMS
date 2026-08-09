'use client';

import React, { useState, useEffect, useTransition } from 'react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { useDebounce } from 'use-debounce';
import { Document, DocumentType, ContractType, User, PaginatedResponse } from '@/types';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { deleteDocument } from '@/app/api/documents';
import { toast } from 'sonner';
import { 
  FileText, 
  Trash2, 
  Search, 
  X, 
  ChevronLeft, 
  ChevronRight, 
  Calendar, 
  DollarSign, 
  Clock,
  Sparkles,
  FileCheck
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

interface DocumentsContentProps {
  initialData: PaginatedResponse<Document>;
  currentUser: User;
}

export const DocumentsContent: React.FC<DocumentsContentProps> = ({ initialData, currentUser }) => {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();

  // Search input state (local to allow immediate typing)
  const [searchVal, setSearchVal] = useState(searchParams.get('search') || '');
  const [debouncedSearch] = useDebounce(searchVal, 400);

  // Sync debounced search to URL searchParams
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (debouncedSearch) {
      params.set('search', debouncedSearch);
    } else {
      params.delete('search');
    }
    params.set('page', '1'); // reset to page 1 on search
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  }, [debouncedSearch, router]);

  // Read filters from searchParams
  const docTypeFilter = (searchParams.get('type') as DocumentType) || '';
  const contractTypeFilter = (searchParams.get('contractType') as ContractType) || '';
  const currentPage = Number(searchParams.get('page')) || 1;

  const handleFilterChange = (key: string, value: string) => {
    const params = new URLSearchParams(window.location.search);
    if (value) {
      params.set(key, value);
    } else {
      params.delete(key);
    }
    params.set('page', '1'); // reset to page 1 on filter change
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  };

  const handlePageChange = (newPage: number) => {
    const params = new URLSearchParams(window.location.search);
    params.set('page', newPage.toString());
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  };

  const handleClearFilters = () => {
    setSearchVal('');
    startTransition(() => {
      router.replace(pathname);
    });
  };

  // CRUD actions
  const handleDelete = async (id: string) => {
    const confirmDelete = window.confirm('Deseja realmente enviar este documento para a lixeira?');
    if (!confirmDelete) return;

    try {
      await deleteDocument({ id, path: pathname });
      toast.success('Documento enviado para a lixeira com sucesso!');
    } catch {
      toast.error('Erro ao excluir documento.');
    }
  };



  const hasAccessToActions = currentUser.role === 'ADMIN' || currentUser.role === 'MOD';

  return (
    <div className="flex flex-col gap-8">
      {/* Welcome banner */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-2 text-violet-400 text-xs font-semibold uppercase tracking-wider mb-1.5">
            <Sparkles className="w-4 h-4" />
            Painel de Documentos Municipais
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white">
            Documentos Oficiais
          </h1>
          <p className="text-zinc-400 mt-1.5 text-sm max-w-xl">
            Gerencie, pesquise e acesse os decretos, portarias, editais e contratos de{' '}
            <span className="text-violet-300 font-semibold">{currentUser.municipality?.name || 'sua prefeitura'}</span>.
          </p>
        </div>
      </div>

      {/* Filter panel */}
      <div className="p-6 bg-zinc-950/60 border border-zinc-800/80 backdrop-blur-md rounded-2xl flex flex-col md:flex-row items-end gap-5 shadow-2xl relative overflow-hidden">
        {/* Glow effect */}
        <div className="absolute -top-24 -left-24 w-48 h-48 bg-violet-600/5 blur-3xl rounded-full pointer-events-none" />

        <div className="w-full md:flex-1 relative">
          <Input
            label="Pesquisa rápida"
            placeholder="Buscar por descrição ou palavra-chave..."
            value={searchVal}
            onChange={(e) => setSearchVal(e.target.value)}
            className="pl-10"
          />
          <Search className="w-4 h-4 text-zinc-500 absolute left-3.5 bottom-3.5" />
          {searchVal && (
            <button
              onClick={() => setSearchVal('')}
              className="absolute right-3.5 bottom-3.5 text-zinc-500 hover:text-white p-0.5 rounded-full hover:bg-zinc-800 transition-all"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        {/* Document Type Dropdown */}
        <div className="w-full md:w-52 flex flex-col gap-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-zinc-400">Tipo de Documento</label>
          <select
            className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 cursor-pointer"
            value={docTypeFilter}
            onChange={(e) => handleFilterChange('type', e.target.value)}
          >
            <option value="">Todos</option>
            <option value="NOTICE">Ofício</option>
            <option value="DECREE">Decreto</option>
            <option value="ORDINANCE">Portaria</option>
            <option value="LAW">Lei</option>
            <option value="CONTRACT">Contrato</option>
          </select>
        </div>

        {/* Contract Type Dropdown (visible only when type is CONTRACT) */}
        <AnimatePresence>
          {docTypeFilter === 'CONTRACT' && (
            <motion.div
              initial={{ opacity: 0, width: 0, scale: 0.95 }}
              animate={{ opacity: 1, width: 'auto', scale: 1 }}
              exit={{ opacity: 0, width: 0, scale: 0.95 }}
              className="w-full md:w-52 flex flex-col gap-1.5"
            >
              <label className="text-xs font-semibold uppercase tracking-wider text-zinc-400">Tipo de Contrato</label>
              <select
                className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 cursor-pointer"
                value={contractTypeFilter}
                onChange={(e) => handleFilterChange('contractType', e.target.value)}
              >
                <option value="">Todos</option>
                <option value="service">Prestação de Serviço</option>
                <option value="bidding">Licitação</option>
                <option value="publicinterest">Interesse Público</option>
              </select>
            </motion.div>
          )}
        </AnimatePresence>

        <Button
          variant="outline"
          className="w-full md:w-auto h-11 border-zinc-800 hover:border-zinc-700 hover:bg-zinc-900 text-zinc-300 font-semibold px-5 rounded-xl"
          onClick={handleClearFilters}
        >
          Limpar Filtros
        </Button>
      </div>

      {/* Documents list */}
      <div className="bg-zinc-950/40 border border-zinc-900 rounded-2xl overflow-hidden shadow-2xl relative">
        {/* Loading overlay */}
        {isPending && (
          <div className="absolute inset-0 bg-black/60 backdrop-blur-[2px] z-10 flex items-center justify-center transition-all duration-300">
            <div className="flex flex-col items-center gap-3">
              <div className="w-10 h-10 rounded-full border-2 border-violet-500/20 border-t-violet-500 animate-spin" />
              <span className="text-violet-400 text-xs font-bold tracking-widest uppercase">Atualizando dados...</span>
            </div>
          </div>
        )}

        {initialData.data.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24 px-4 text-center">
            <div className="w-14 h-14 rounded-2xl bg-zinc-900/50 flex items-center justify-center text-zinc-500 mb-5 border border-zinc-800">
              <FileText className="w-7 h-7 text-zinc-600" />
            </div>
            <h3 className="text-xl font-bold text-white">Nenhum documento encontrado</h3>
            <p className="text-zinc-500 text-sm mt-2 max-w-sm">
              Tente redefinir seus termos de busca ou mude os filtros aplicados para listar mais registros.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-zinc-900 bg-zinc-950/70 text-zinc-400 text-[11px] font-bold uppercase tracking-wider">
                  <th className="px-6 py-4.5">Número</th>
                  <th className="px-6 py-4.5">Tipo</th>
                  <th className="px-6 py-4.5">Descrição</th>
                  <th className="px-6 py-4.5">Detalhes / Tipo</th>
                  <th className="px-6 py-4.5 text-right">Data de Registro</th>
                  {hasAccessToActions && <th className="px-6 py-4.5 text-center">Ações</th>}
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-900 text-sm text-zinc-300">
                {initialData.data.map((doc) => (
                  <tr key={doc.id} className="hover:bg-zinc-900/10 transition-colors group">
                    <td className="px-6 py-4.5 font-bold text-white group-hover:text-violet-400 transition-colors">
                      #{doc.order}
                    </td>
                    <td className="px-6 py-4.5">
                      <span className={`inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold border ${
                        doc.type === 'CONTRACT' 
                          ? 'bg-violet-500/5 text-violet-400 border-violet-500/20' 
                          : doc.type === 'LAW'
                          ? 'bg-emerald-500/5 text-emerald-400 border-emerald-500/20'
                          : 'bg-zinc-900/80 text-zinc-400 border-zinc-800'
                      }`}>
                        {doc.type}
                      </span>
                    </td>
                    <td className="px-6 py-4.5 max-w-xs md:max-w-md truncate font-medium">
                      {doc.description}
                    </td>
                    <td className="px-6 py-4.5 text-xs">
                      {doc.type === 'CONTRACT' && doc.contractType ? (
                        <div className="flex flex-col gap-1 text-zinc-400">
                          <span className="capitalize font-semibold text-zinc-300 flex items-center gap-1.5">
                            <FileCheck className="w-3.5 h-3.5 text-violet-400" />
                            {doc.contractType}
                          </span>
                          {doc.value && (
                            <span className="flex items-center gap-1 text-emerald-400 font-semibold">
                              <DollarSign className="w-3 h-3" />
                              {doc.value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}
                            </span>
                          )}
                          {doc.duration && (
                            <span className="flex items-center gap-1 text-zinc-500 font-medium">
                              <Clock className="w-3 h-3" />
                              {doc.duration} meses
                            </span>
                          )}
                        </div>
                      ) : (
                        <span className="text-zinc-600">—</span>
                      )}
                    </td>
                    <td className="px-6 py-4.5 text-right text-zinc-500 text-xs font-medium">
                      <div className="flex items-center justify-end gap-1.5">
                        <Calendar className="w-3.5 h-3.5 text-zinc-600" />
                        {new Date(doc.createdAt).toLocaleDateString('pt-BR')}
                      </div>
                    </td>
                    {hasAccessToActions && (
                      <td className="px-6 py-4.5 text-center">
                        <div className="flex items-center justify-center gap-2 opacity-50 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => handleDelete(doc.id)}
                            className="p-1.5 text-zinc-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                            title="Mover para Lixeira"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination bar */}
        {initialData.total > initialData.pageSize && (
          <div className="px-6 py-4.5 border-t border-zinc-900 bg-zinc-950/20 flex flex-col sm:flex-row items-center justify-between gap-4">
            <span className="text-xs text-zinc-500 font-semibold uppercase tracking-wider">
              Mostrando {initialData.data.length} de {initialData.total} documentos
            </span>
            <div className="flex items-center gap-3">
              <Button
                variant="outline"
                className="px-3.5 py-1.5 text-xs h-9 border-zinc-800 text-zinc-400 hover:text-white rounded-xl"
                onClick={() => handlePageChange(Math.max(1, currentPage - 1))}
                disabled={currentPage === 1 || isPending}
              >
                <ChevronLeft className="w-4 h-4 mr-1" />
                Anterior
              </Button>
              <span className="text-xs text-zinc-400 font-bold px-3 py-1.5 rounded-lg bg-zinc-900">
                Página {currentPage}
              </span>
              <Button
                variant="outline"
                className="px-3.5 py-1.5 text-xs h-9 border-zinc-800 text-zinc-400 hover:text-white rounded-xl"
                onClick={() => handlePageChange(currentPage + 1)}
                disabled={currentPage * initialData.pageSize >= initialData.total || isPending}
              >
                Próxima
                <ChevronRight className="w-4 h-4 ml-1" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
