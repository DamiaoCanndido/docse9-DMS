'use client';

import React, { useState, useEffect, useTransition } from 'react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { useDebounce } from 'use-debounce';
import { Document, DocumentType, ContractType, User, PaginatedResponse, UserPermission, PermissionLevel } from '@/types';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import {
  createDocument,
  updateDocument,
  deleteDocument,
  restoreDocument,
  hardDeleteDocument,
} from '@/app/api/documents';
import { getUserPermissions } from '@/app/api/users';
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
  FileCheck,
  Plus,
  Edit2,
  RotateCcw,
  Trash
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

interface DocumentsContentProps {
  initialData: PaginatedResponse<Document>;
  currentUser: User;
  viewTrash?: boolean;
}

export const DocumentsContent: React.FC<DocumentsContentProps> = ({ 
  initialData, 
  currentUser,
  viewTrash = false
}) => {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();

  // Search input state
  const [searchVal, setSearchVal] = useState(searchParams.get('search') || '');
  const [debouncedSearch] = useDebounce(searchVal, 400);

  // User permissions (for COMMON role)
  const [userPermissions, setUserPermissions] = useState<UserPermission[]>([]);

  useEffect(() => {
    if (currentUser.role === 'COMMON') {
      getUserPermissions(currentUser.id)
        .then((perms) => setUserPermissions(perms))
        .catch((err) => console.error('Erro ao buscar permissões do usuário:', err));
    }
  }, [currentUser]);

  // Sync debounced search to URL searchParams
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (debouncedSearch) {
      params.set('search', debouncedSearch);
    } else {
      params.delete('search');
    }
    params.set('page', '1');
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  }, [debouncedSearch, router]);

  // Read filters from searchParams
  const docTypeFilter = (searchParams.get('type') as DocumentType) || '';
  const contractTypeFilter = (searchParams.get('contractType') as ContractType) || '';
  const currentPage = Number(searchParams.get('page')) || 1;

  // Modal form states
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingDocument, setEditingDocument] = useState<Document | null>(null);
  
  const [type, setType] = useState<DocumentType>('NOTICE');
  const [description, setDescription] = useState('');
  const [contractType, setContractType] = useState<ContractType>('service');
  const [value, setValue] = useState('');
  const [duration, setDuration] = useState('');
  const [startIn, setStartIn] = useState('');
  const [formError, setFormError] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  // Permissions helpers
  const canCreate = (t: DocumentType) => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      const p = userPermissions.find((perm) => perm.documentType === t);
      return p && (p.level === 'WRITE' || p.level === 'DELETE');
    }
    return false;
  };

  const canEdit = (t: DocumentType) => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      const p = userPermissions.find((perm) => perm.documentType === t);
      return p && (p.level === 'WRITE' || p.level === 'DELETE');
    }
    return false;
  };

  const canDelete = (t: DocumentType) => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      const p = userPermissions.find((perm) => perm.documentType === t);
      return p && p.level === 'DELETE';
    }
    return false;
  };

  const hasAnyWritePermission = 
    currentUser.role === 'MOD' || 
    userPermissions.some((p) => p.level === 'WRITE' || p.level === 'DELETE');

  const docTypesList: { value: DocumentType; label: string }[] = [
    { value: 'NOTICE', label: 'Ofício (Notice)' },
    { value: 'DECREE', label: 'Decreto (Decree)' },
    { value: 'ORDINANCE', label: 'Portaria (Ordinance)' },
    { value: 'LAW', label: 'Lei (Law)' },
    { value: 'CONTRACT', label: 'Contrato (Contract)' },
  ];

  // Filters
  const handleFilterChange = (key: string, value: string) => {
    const params = new URLSearchParams(window.location.search);
    if (value) {
      params.set(key, value);
    } else {
      params.delete(key);
    }
    params.set('page', '1');
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
    const params = new URLSearchParams();
    if (viewTrash) {
      params.set('trash', 'true');
    }
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  };

  const handleToggleTrash = () => {
    const params = new URLSearchParams(window.location.search);
    if (viewTrash) {
      params.delete('trash');
    } else {
      params.set('trash', 'true');
    }
    params.set('page', '1');
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  };

  // Modals actions
  const openCreateModal = () => {
    // Encontra o primeiro tipo de documento que o usuário tem permissão para criar
    const allowedType = docTypesList.find((t) => canCreate(t.value));
    setType(allowedType ? allowedType.value : 'NOTICE');
    setEditingDocument(null);
    setDescription('');
    setContractType('service');
    setValue('');
    setDuration('');
    setStartIn('');
    setFormError('');
    setIsModalOpen(true);
  };

  const openEditModal = (doc: Document) => {
    setEditingDocument(doc);
    setType(doc.type);
    setDescription(doc.description);
    setContractType(doc.contractType || 'service');
    setValue(doc.value ? doc.value.toString() : '');
    setDuration(doc.duration ? doc.duration.toString() : '');
    setStartIn(doc.startIn ? new Date(doc.startIn).toISOString().split('T')[0] : '');
    setFormError('');
    setIsModalOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!description.trim()) {
      setFormError('A descrição do documento é obrigatória.');
      return;
    }

    if (type === 'CONTRACT') {
      if (!value.trim() || isNaN(Number(value)) || Number(value) <= 0) {
        setFormError('O valor do contrato é obrigatório e deve ser maior que zero.');
        return;
      }
      if (!duration.trim() || isNaN(Number(duration)) || Number(duration) <= 0) {
        setFormError('A duração em meses do contrato é obrigatória e deve ser maior que zero.');
        return;
      }
      if (!startIn) {
        setFormError('A data de início do contrato é obrigatória.');
        return;
      }
    }

    setIsSaving(true);
    setFormError('');

    try {
      if (editingDocument) {
        const input: any = {
          description,
        };

        if (type === 'CONTRACT') {
          input.contractType = contractType;
          input.value = Number(value);
          input.duration = Number(duration);
          input.startIn = new Date(startIn).toISOString();
        }

        await updateDocument({
          id: editingDocument.id,
          input,
          path: pathname,
        });
        toast.success('Documento atualizado com sucesso!');
      } else {
        const input: any = {
          type,
          description,
          creatorId: currentUser.id,
          municipalityId: currentUser.municipalityId,
        };

        if (type === 'CONTRACT') {
          input.contractType = contractType;
          input.value = Number(value);
          input.duration = Number(duration);
          input.startIn = new Date(startIn).toISOString();
        }

        await createDocument({
          input,
          path: pathname,
        });
        toast.success('Documento criado com sucesso!');
      }
      setIsModalOpen(false);
    } catch (err: any) {
      setFormError(err.response?.data?.error || 'Erro ao salvar o documento.');
      toast.error('Ocorreu um erro.');
    } finally {
      setIsSaving(false);
    }
  };

  // Trash actions
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

  const handleRestore = async (id: string) => {
    try {
      await restoreDocument({ id, path: pathname });
      toast.success('Documento restaurado com sucesso!');
    } catch {
      toast.error('Erro ao restaurar documento.');
    }
  };

  const handleHardDelete = async (id: string) => {
    const confirmDelete = window.confirm(
      'ATENÇÃO: Esta ação é definitiva e apagará permanentemente o arquivo e seus registros. Confirmar?'
    );
    if (!confirmDelete) return;

    try {
      await hardDeleteDocument({ id, path: pathname });
      toast.success('Documento deletado permanentemente.');
    } catch {
      toast.error('Erro ao excluir definitivamente.');
    }
  };

  return (
    <div className="flex flex-col gap-8">
      {/* Welcome banner */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-2 text-violet-400 text-xs font-semibold uppercase tracking-wider mb-1.5">
            <Sparkles className="w-4 h-4" />
            Painel de Documentos Municipais
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white flex items-center gap-3">
            <FileText className="w-8 h-8 text-violet-500" />
            {viewTrash ? 'Documentos Excluídos (Lixeira)' : 'Documentos Oficiais'}
          </h1>
          <p className="text-zinc-400 mt-1.5 text-sm max-w-xl">
            {viewTrash 
              ? 'Visualize, restaure ou delete em definitivo os documentos excluídos de seu município.' 
              : `Gerencie, pesquise e acesse os decretos, portarias, ofícios e contratos de ${
                  currentUser.municipality?.name || 'sua prefeitura'
                }.`}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            className="border-zinc-800 hover:border-zinc-700 hover:bg-zinc-900 text-zinc-300 font-semibold px-4.5 rounded-xl h-11"
            onClick={handleToggleTrash}
          >
            {viewTrash ? 'Ver Ativos' : 'Ver Lixeira'}
          </Button>

          {!viewTrash && hasAnyWritePermission && (
            <Button
              variant="primary"
              className="flex items-center gap-2 font-semibold px-5 rounded-xl h-11"
              onClick={openCreateModal}
            >
              <Plus className="w-4.5 h-4.5" />
              Novo Documento
            </Button>
          )}
        </div>
      </div>

      {/* Filter panel */}
      {!viewTrash && (
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
      )}

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
                  <th className="px-6 py-4.5 text-center">Ações</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-900 text-sm text-zinc-300">
                {initialData.data.map((doc) => {
                  const showEdit = !viewTrash && canEdit(doc.type);
                  const showDelete = !viewTrash && canDelete(doc.type);
                  const showRestore = viewTrash && (currentUser.role === 'MOD');
                  const showHardDelete = viewTrash && (currentUser.role === 'MOD');

                  return (
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
                          <span className="text-zinc-650">—</span>
                        )}
                      </td>
                      <td className="px-6 py-4.5 text-right text-zinc-500 text-xs font-medium">
                        <div className="flex items-center justify-end gap-1.5">
                          <Calendar className="w-3.5 h-3.5 text-zinc-600" />
                          {new Date(doc.createdAt).toLocaleDateString('pt-BR')}
                        </div>
                      </td>
                      <td className="px-6 py-4.5 text-center">
                        <div className="flex items-center justify-center gap-2 opacity-50 group-hover:opacity-100 transition-opacity">
                          {showEdit && (
                            <button
                              onClick={() => openEditModal(doc)}
                              className="p-1.5 text-zinc-400 hover:text-violet-400 hover:bg-violet-500/10 rounded-lg transition-all"
                              title="Editar"
                            >
                              <Edit2 className="w-4.5 h-4.5" />
                            </button>
                          )}
                          {showDelete && (
                            <button
                              onClick={() => handleDelete(doc.id)}
                              className="p-1.5 text-zinc-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                              title="Mover para Lixeira"
                            >
                              <Trash2 className="w-4.5 h-4.5" />
                            </button>
                          )}
                          {showRestore && (
                            <button
                              onClick={() => handleRestore(doc.id)}
                              className="p-1.5 text-zinc-400 hover:text-emerald-400 hover:bg-emerald-500/10 rounded-lg transition-all"
                              title="Restaurar"
                            >
                              <RotateCcw className="w-4.5 h-4.5" />
                            </button>
                          )}
                          {showHardDelete && (
                            <button
                              onClick={() => handleHardDelete(doc.id)}
                              className="p-1.5 text-zinc-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                              title="Excluir Permanentemente"
                            >
                              <Trash className="w-4.5 h-4.5" />
                            </button>
                          )}
                          {!showEdit && !showDelete && !showRestore && !showHardDelete && (
                            <span className="text-zinc-600 italic text-xs">Sem ações</span>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
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

      {/* Create/Edit Document Modal */}
      <AnimatePresence>
        {isModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="absolute inset-0 bg-black/80 backdrop-blur-sm"
              onClick={() => setIsModalOpen(false)}
            />

            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: 10 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 10 }}
              className="relative w-full max-w-xl bg-zinc-950 border border-zinc-800/80 rounded-2xl shadow-2xl p-6 overflow-hidden"
            >
              <div className="absolute top-0 right-0 p-4">
                <button
                  onClick={() => setIsModalOpen(false)}
                  className="text-zinc-500 hover:text-white transition-colors p-1 hover:bg-zinc-900 rounded-lg"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <h2 className="text-xl font-bold text-white mb-6">
                {editingDocument ? 'Editar Documento' : 'Novo Documento Oficial'}
              </h2>

              <form onSubmit={handleSave} className="flex flex-col gap-5">
                {/* Type selection - read-only on edit */}
                <div className="w-full flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-zinc-400">Tipo de Documento</label>
                  {editingDocument ? (
                    <div className="w-full bg-zinc-900/50 border border-zinc-850 px-3.5 py-2.5 rounded-lg text-sm text-zinc-400 font-semibold uppercase">
                      {type}
                    </div>
                  ) : (
                    <select
                      className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3.5 py-2.5 rounded-lg text-sm transition-all focus:outline-none focus:border-violet-500 cursor-pointer"
                      value={type}
                      onChange={(e) => setType(e.target.value as DocumentType)}
                    >
                      {docTypesList.map((t) => {
                        const allowed = canCreate(t.value);
                        return (
                          <option key={t.value} value={t.value} disabled={!allowed}>
                            {t.label} {!allowed ? '(Sem permissão)' : ''}
                          </option>
                        );
                      })}
                    </select>
                  )}
                </div>

                {/* Description textarea */}
                <div className="w-full flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-zinc-400">Descrição / Ementa</label>
                  <textarea
                    className="w-full h-28 bg-zinc-900 border border-border text-foreground px-3.5 py-2.5 rounded-lg text-sm transition-all duration-200 placeholder-zinc-500 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                    placeholder="Descreva o conteúdo do documento ou sua ementa oficial..."
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    required
                  />
                </div>

                {/* Specific fields for CONTRACT */}
                {type === 'CONTRACT' && (
                  <motion.div
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="border border-zinc-900/80 rounded-xl p-4 bg-zinc-950/20 flex flex-col gap-4"
                  >
                    <h3 className="text-xs font-bold text-violet-400 uppercase tracking-widest flex items-center gap-1.5 mb-1">
                      <FileCheck className="w-4 h-4" />
                      Detalhes do Contrato
                    </h3>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {/* Contract Type */}
                      <div className="w-full flex flex-col gap-1.5">
                        <label className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                          Tipo de Contrato
                        </label>
                        <select
                          className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3 py-2 rounded-lg text-xs transition-all focus:outline-none focus:border-violet-500 cursor-pointer"
                          value={contractType}
                          onChange={(e) => setContractType(e.target.value as ContractType)}
                        >
                          <option value="service">Prestação de Serviço</option>
                          <option value="bidding">Licitação</option>
                          <option value="publicinterest">Interesse Público</option>
                        </select>
                      </div>

                      {/* Value */}
                      <Input
                        label="Valor do Contrato (R$)"
                        type="number"
                        step="0.01"
                        placeholder="0.00"
                        value={value}
                        onChange={(e) => setValue(e.target.value)}
                        className="text-xs py-2"
                        required
                      />

                      {/* Duration */}
                      <Input
                        label="Duração (em meses)"
                        type="number"
                        placeholder="Ex: 12"
                        value={duration}
                        onChange={(e) => setDuration(e.target.value)}
                        className="text-xs py-2"
                        required
                      />

                      {/* Start Date */}
                      <Input
                        label="Data de Início"
                        type="date"
                        value={startIn}
                        onChange={(e) => setStartIn(e.target.value)}
                        className="text-xs py-2"
                        required
                      />
                    </div>
                  </motion.div>
                )}

                {formError && (
                  <span className="text-xs text-red-500 font-semibold bg-red-500/10 border border-red-500/20 px-3.5 py-2.5 rounded-xl">
                    {formError}
                  </span>
                )}

                <div className="flex items-center justify-end gap-3 mt-4">
                  <Button
                    type="button"
                    variant="outline"
                    className="border-zinc-800 text-zinc-300 rounded-xl"
                    onClick={() => setIsModalOpen(false)}
                  >
                    Cancelar
                  </Button>
                  <Button
                    type="submit"
                    variant="primary"
                    className="rounded-xl font-bold px-6"
                    isLoading={isSaving}
                  >
                    Salvar
                  </Button>
                </div>
              </form>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </div>
  );
};
