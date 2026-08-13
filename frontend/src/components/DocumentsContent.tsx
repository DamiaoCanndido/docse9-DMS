'use client';

import React, { useState, useEffect, useTransition } from 'react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { useDebounce } from 'use-debounce';
import { Document, DocumentType, ContractType, User, PaginatedResponse, UserPermission, PermissionLevel } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { PaginationBar } from '@/components/ui/PaginationBar';
import { motion, AnimatePresence } from 'framer-motion';
import {
  createDocument,
  updateDocument,
  deleteDocument,
  restoreDocument,
  hardDeleteDocument,
} from '@/app/api/documents';
import { getUserPermissions } from '@/app/api/users';
import { toast } from 'sonner';
import { isRedirectError } from '@/lib/utils';
import { 
  FileText, 
  Trash2, 
  Search, 
  X, 
  ChevronLeft, 
  ChevronRight, 
  Calendar as CalendarIcon, 
  DollarSign, 
  Clock,
  Sparkles,
  FileCheck,
  Plus,
  Edit2,
  RotateCcw,
  Trash,
  MoreHorizontal,
  Scale,
  FileSignature,
  FolderOpen,
  ShieldAlert,
  Building2
} from 'lucide-react';
import { format } from 'date-fns';
import { ptBR } from 'date-fns/locale';
import { cn } from '@/lib/utils';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Calendar } from '@/components/ui/calendar';

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
  const [isLoadedPermissions, setIsLoadedPermissions] = useState(false);

  useEffect(() => {
    if (currentUser.role === 'COMMON') {
      getUserPermissions(currentUser.id)
        .then((perms) => {
          setUserPermissions(perms);
          setIsLoadedPermissions(true);
        })
        .catch((err) => {
          console.error('Erro ao buscar permissões do usuário:', err);
          setIsLoadedPermissions(true);
        });
    } else {
      setIsLoadedPermissions(true);
    }
  }, [currentUser]);

  const hasNoPermissions = 
    currentUser.role === 'COMMON' && 
    isLoadedPermissions && 
    !userPermissions.some((p) => p.level === 'READ' || p.level === 'WRITE' || p.level === 'DELETE');

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
  const activeTab = (searchParams.get('type') as DocumentType) || 'NOTICE';
  const contractTypeFilter = (searchParams.get('contractType') as ContractType) || '';
  const currentYear = new Date().getFullYear();
  const yearFilter = searchParams.get('year') || currentYear.toString();
  const currentPage = Number(searchParams.get('page')) || 1;

  const availableYears = [
    currentYear.toString(),
    (currentYear - 1).toString(),
    (currentYear - 2).toString(),
    (currentYear - 3).toString(),
    'all',
  ];

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

  // Date parsing for shadcn Calendar
  const dateValue = startIn ? new Date(startIn) : undefined;

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

  const canViewTab = (t: DocumentType) => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      if (userPermissions.length === 0) return true;
      const p = userPermissions.find((perm) => perm.documentType === t);
      return p && p.level !== 'NONE';
    }
    return false;
  };

  const hasAnyWritePermission = 
    currentUser.role === 'MOD' || 
    userPermissions.some((p) => p.level === 'WRITE' || p.level === 'DELETE');

  const docTypesList: { value: DocumentType; label: string; singleLabel: string; icon: React.ComponentType<{ className?: string }> }[] = [
    { value: 'NOTICE', label: 'Ofícios', singleLabel: 'Ofício', icon: FileText },
    { value: 'DECREE', label: 'Decretos', singleLabel: 'Decreto', icon: FileCheck },
    { value: 'ORDINANCE', label: 'Portarias', singleLabel: 'Portaria', icon: Clock },
    { value: 'LAW', label: 'Leis', singleLabel: 'Lei', icon: Scale },
    { value: 'CONTRACT', label: 'Contratos', singleLabel: 'Contrato', icon: FileSignature },
  ];

  const contractTypeLabels: Record<ContractType, string> = {
    service: 'Prestação de Serviço',
    bidding: 'Licitação',
    publicinterest: 'Interesse Público',
  };

  // Filters
  const handleTabChange = (newType: DocumentType) => {
    if (newType === activeTab) return;
    const params = new URLSearchParams(window.location.search);
    params.set('type', newType);
    params.set('page', '1');
    if (newType !== 'CONTRACT') {
      params.delete('contractType');
    }
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  };

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

  const handlePageSizeChange = (newPageSize: number) => {
    const params = new URLSearchParams(window.location.search);
    params.set('pageSize', newPageSize.toString());
    params.set('page', '1');
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  };

  const handleClearFilters = () => {
    setSearchVal('');
    const params = new URLSearchParams();
    params.set('type', activeTab);
    params.set('year', currentYear.toString());
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
    const defaultType = canCreate(activeTab) ? activeTab : (docTypesList.find((t) => canCreate(t.value))?.value || 'NOTICE');
    setType(defaultType);
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
    setStartIn(doc.startIn ? new Date(doc.startIn).toISOString() : '');
    setFormError('');
    setIsModalOpen(true);
  };

  const handleSave = async (e: React.SubmitEvent<HTMLFormElement>) => {
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
      if (isRedirectError(err)) {
        throw err;
      }
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

  const activeTabMeta = docTypesList.find(t => t.value === activeTab) || docTypesList[0];

  if (hasNoPermissions) {
    return (
      <div className="flex flex-col items-center justify-center py-20 px-4 text-center max-w-2xl mx-auto">
        <div className="w-20 h-20 rounded-3xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-400 mb-6 shadow-xl shadow-violet-500/5">
          <ShieldAlert className="w-10 h-10" />
        </div>
        <h2 className="text-2xl font-extrabold text-white tracking-tight">
          Acesso Pendente de Liberação
        </h2>
        <p className="text-zinc-400 text-sm mt-3 leading-relaxed">
          Olá, <strong className="text-zinc-200">{currentUser.username}</strong>! Sua conta foi registrada com sucesso no sistema, porém você ainda não possui permissões atribuídas para visualizar ou gerenciar documentos oficiais.
        </p>
        <div className="mt-6 p-5 bg-zinc-950/80 border border-zinc-800 rounded-2xl text-xs text-zinc-400 text-left flex flex-col gap-2.5 w-full shadow-2xl">
          <span className="font-semibold text-zinc-300 flex items-center gap-2 text-sm">
            <Building2 className="w-4 h-4 text-violet-400" />
            Como solicitar acesso?
          </span>
          <p className="leading-relaxed">
            Entre em contato com um <strong className="text-violet-300">Moderador</strong> ou <strong className="text-violet-300">Administrador</strong> do município de <strong className="text-zinc-200">{currentUser.municipality?.name || 'sua prefeitura'}</strong> para que sejam concedidas as permissões de leitura/escrita nos tipos de documentos desejados (Ofícios, Decretos, Portarias, Leis ou Contratos).
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
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
              : `Gerencie e pesquise os documentos oficiais de ${
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
              variant="default"
              className="flex items-center gap-2 font-semibold px-5 rounded-xl h-11"
              onClick={openCreateModal}
            >
              <Plus className="w-4.5 h-4.5" />
              Novo Documento
            </Button>
          )}
        </div>
      </div>

      {/* Navigation Tabs Bar by Document Type */}
      <div className="flex items-center gap-1.5 p-1.5 bg-zinc-950/80 border border-zinc-800/90 rounded-2xl overflow-x-auto scrollbar-none shadow-xl">
        {docTypesList.map((tab) => {
          const isActive = activeTab === tab.value;
          const Icon = tab.icon;
          const allowed = canViewTab(tab.value);
          if (!allowed) return null;

          return (
            <button
              key={tab.value}
              onClick={() => handleTabChange(tab.value)}
              className={cn(
                "flex items-center gap-2.5 px-5 py-3 rounded-xl text-sm font-semibold transition-all duration-200 cursor-pointer whitespace-nowrap select-none relative",
                isActive
                  ? "bg-violet-600 text-white shadow-lg shadow-violet-600/25 font-bold"
                  : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900/60"
              )}
            >
              <Icon className={cn("w-4 h-4", isActive ? "text-white" : "text-zinc-400")} />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Filter panel */}
      {!viewTrash && (
        <div className="p-5 bg-zinc-950/60 border border-zinc-800/80 backdrop-blur-md rounded-2xl flex flex-col md:flex-row items-end gap-4 shadow-2xl relative overflow-hidden">
          {/* Glow effect */}
          <div className="absolute -top-24 -left-24 w-48 h-48 bg-violet-600/5 blur-3xl rounded-full pointer-events-none" />

          <div className="w-full md:flex-1 relative">
            <Input
              label={`Pesquisa em ${activeTabMeta.label}`}
              placeholder={`Buscar por descrição em ${activeTabMeta.label.toLowerCase()}...`}
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

          {/* Year Dropdown */}
          <div className="w-full md:w-44 flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-wider text-zinc-400">Ano</label>
            <select
              className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 cursor-pointer"
              value={yearFilter}
              onChange={(e) => handleFilterChange('year', e.target.value)}
            >
              {availableYears.map((y) => (
                <option key={y} value={y}>
                  {y === 'all' ? 'Todos os Anos' : y}
                </option>
              ))}
            </select>
          </div>

          {/* Contract Type Dropdown (visible ONLY on CONTRACT tab) */}
          <AnimatePresence>
            {activeTab === 'CONTRACT' && (
              <motion.div
                initial={{ opacity: 0, width: 0, scale: 0.95 }}
                animate={{ opacity: 1, width: 'auto', scale: 1 }}
                exit={{ opacity: 0, width: 0, scale: 0.95 }}
                className="w-full md:w-56 flex flex-col gap-1.5"
              >
                <label className="text-xs font-semibold uppercase tracking-wider text-zinc-400">Tipo de Contrato</label>
                <select
                  className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 cursor-pointer"
                  value={contractTypeFilter}
                  onChange={(e) => handleFilterChange('contractType', e.target.value)}
                >
                  <option value="">Todos os Contratos</option>
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
              <span className="text-violet-400 text-xs font-bold tracking-widest uppercase">Carregando {activeTabMeta.label.toLowerCase()}...</span>
            </div>
          </div>
        )}

        {initialData.data.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 px-4 text-center">
            <div className="w-14 h-14 rounded-2xl bg-zinc-900/60 flex items-center justify-center text-zinc-500 mb-4 border border-zinc-800">
              <FolderOpen className="w-7 h-7 text-zinc-500" />
            </div>
            <h3 className="text-xl font-bold text-white">Nenhum registro encontrado</h3>
            <p className="text-zinc-400 text-sm mt-1.5 max-w-sm">
              Não há {activeTabMeta.label.toLowerCase()} cadastrados{searchVal ? ' com os termos buscados' : ''}.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-zinc-900 bg-zinc-950/70 text-zinc-400 text-[11px] font-bold uppercase tracking-wider">
                  <th className="px-6 py-4.5">Número</th>
                  <th className="px-6 py-4.5">Descrição / Ementa</th>
                  {activeTab === 'CONTRACT' ? (
                    <>
                      <th className="px-6 py-4.5">Tipo de Contrato</th>
                      <th className="px-6 py-4.5">Valor (R$)</th>
                      <th className="px-6 py-4.5">Duração</th>
                      <th className="px-6 py-4.5 text-right">Data de Início</th>
                    </>
                  ) : (
                    <>
                      <th className="px-6 py-4.5">Criado Por</th>
                      <th className="px-6 py-4.5 text-right">Data de Registro</th>
                    </>
                  )}
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
                    <tr key={doc.id} className="hover:bg-zinc-900/20 transition-colors group">
                      <td className="px-6 py-4.5 font-bold text-white group-hover:text-violet-400 transition-colors">
                        #{doc.order}
                      </td>
                      <td className="px-6 py-4.5 max-w-xs md:max-w-md font-medium">
                        {doc.description}
                      </td>
                      {activeTab === 'CONTRACT' ? (
                        <>
                          <td className="px-6 py-4.5 text-xs">
                            <span className="font-semibold text-zinc-300 flex items-center gap-1.5">
                              <FileCheck className="w-3.5 h-3.5 text-violet-400" />
                              {doc.contractType ? (contractTypeLabels[doc.contractType] || doc.contractType) : '—'}
                            </span>
                          </td>
                          <td className="px-6 py-4.5 text-xs font-semibold text-emerald-400">
                            {doc.value ? (
                              doc.value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
                            ) : '—'}
                          </td>
                          <td className="px-6 py-4.5 text-xs text-zinc-400 font-medium">
                            {doc.duration ? `${doc.duration} meses` : '—'}
                          </td>
                          <td className="px-6 py-4.5 text-right text-zinc-400 text-xs font-medium">
                            <div className="flex items-center justify-end gap-1.5">
                              <CalendarIcon className="w-3.5 h-3.5 text-zinc-500" />
                              {doc.startIn ? format(new Date(doc.startIn), 'dd/MM/yyyy HH:mm', { locale: ptBR }) : '—'}
                            </div>
                          </td>
                        </>
                      ) : (
                        <>
                          <td className="px-6 py-4.5 text-xs text-zinc-400">
                            {doc.createdBy?.username || doc.createdBy?.email || 'Sistema'}
                          </td>
                          <td className="px-6 py-4.5 text-right text-zinc-400 text-xs font-medium">
                            <div className="flex items-center justify-end gap-1.5">
                              <CalendarIcon className="w-3.5 h-3.5 text-zinc-500" />
                              {format(new Date(doc.createdAt), 'dd/MM/yyyy HH:mm', { locale: ptBR })}
                            </div>
                          </td>
                        </>
                      )}
                      <td className="px-6 py-4.5 text-center">
                        <DropdownMenu>
                          <DropdownMenuTrigger render={
                            <button className="p-1.5 text-zinc-400 hover:text-white hover:bg-zinc-900 rounded-lg transition-all cursor-pointer">
                              <MoreHorizontal className="w-4.5 h-4.5" />
                            </button>
                          } />
                          <DropdownMenuContent align="end" className="bg-zinc-950 border-zinc-800 text-zinc-300 min-w-[170px]">
                            {showEdit && (
                              <DropdownMenuItem
                                onClick={() => openEditModal(doc)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-violet-400 cursor-pointer focus:bg-zinc-900 focus:text-violet-400 p-2 text-xs font-medium"
                              >
                                <Edit2 className="w-4 h-4" />
                                Editar
                              </DropdownMenuItem>
                            )}
                            {showDelete && (
                              <DropdownMenuItem
                                onClick={() => handleDelete(doc.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-red-400 cursor-pointer focus:bg-zinc-900 focus:text-red-400 p-2 text-xs font-medium"
                              >
                                <Trash2 className="w-4 h-4" />
                                Mover para Lixeira
                              </DropdownMenuItem>
                            )}
                            {showRestore && (
                              <DropdownMenuItem
                                onClick={() => handleRestore(doc.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-emerald-400 cursor-pointer focus:bg-zinc-900 focus:text-emerald-400 p-2 text-xs font-medium"
                              >
                                <RotateCcw className="w-4.5 h-4.5" />
                                Restaurar
                              </DropdownMenuItem>
                            )}
                            {showHardDelete && (
                              <DropdownMenuItem
                                onClick={() => handleHardDelete(doc.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-red-400 cursor-pointer focus:bg-zinc-900 focus:text-red-400 p-2 text-xs font-medium"
                              >
                                <Trash className="w-4.5 h-4.5" />
                                Excluir Definitivamente
                              </DropdownMenuItem>
                            )}
                            {!showEdit && !showDelete && !showRestore && !showHardDelete && (
                              <span className="p-2 text-xs text-zinc-500 italic block text-center">Sem permissões</span>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination bar */}
        <PaginationBar
          currentPage={initialData.page}
          pageSize={initialData.pageSize}
          total={initialData.total}
          itemLabel="documentos"
          isPending={isPending}
          onPageChange={handlePageChange}
          onPageSizeChange={handlePageSizeChange}
        />
      </div>

      {/* Create/Edit Document Modal via shadcn/ui Dialog */}
      <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
        <DialogContent className="bg-zinc-950 border-zinc-800/80 text-zinc-300 max-w-xl rounded-2xl shadow-2xl p-6">
          <DialogHeader>
            <DialogTitle className="text-xl font-bold text-white">
              {editingDocument ? 'Editar Documento' : 'Novo Documento Oficial'}
            </DialogTitle>
            <DialogDescription className="text-zinc-500 text-xs mt-1">
              {editingDocument
                ? 'Modifique os metadados do documento selecionado.'
                : 'Cadastre um novo documento oficial no isolamento do seu município.'}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSave} className="flex flex-col gap-5 mt-4">
            {/* Type selection - read-only on edit */}
            <div className="w-full flex flex-col gap-1.5">
              <label className="text-sm font-medium text-zinc-400">Tipo de Documento</label>
              {editingDocument ? (
                <div className="w-full bg-zinc-900/50 border border-zinc-850 px-3.5 py-2.5 rounded-lg text-sm text-zinc-300 font-semibold">
                  {docTypesList.find((t) => t.value === type)?.singleLabel || type}
                </div>
              ) : (
                <Select value={type} onValueChange={(val) => setType(val as DocumentType)}>
                  <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-300 text-sm h-10">
                    <SelectValue placeholder="Selecione o tipo" />
                  </SelectTrigger>
                  <SelectContent className="bg-zinc-950 border-zinc-800 text-zinc-300">
                    {docTypesList.map((t) => {
                      const allowed = canCreate(t.value);
                      return (
                        <SelectItem key={t.value} value={t.value} disabled={!allowed}>
                          {t.singleLabel} {!allowed ? '(Sem permissão)' : ''}
                        </SelectItem>
                      );
                    })}
                  </SelectContent>
                </Select>
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

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-end">
                  {/* Contract Type */}
                  <div className="w-full flex flex-col gap-1.5">
                    <label className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                      Tipo de Contrato
                    </label>
                    <Select value={contractType} onValueChange={(val) => setContractType(val as ContractType)}>
                      <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-300 text-xs h-10">
                        <SelectValue placeholder="Selecione o tipo de contrato" />
                      </SelectTrigger>
                      <SelectContent className="bg-zinc-950 border-zinc-800 text-zinc-300">
                        <SelectItem value="service">Prestação de Serviço</SelectItem>
                        <SelectItem value="bidding">Licitação</SelectItem>
                        <SelectItem value="publicinterest">Interesse Público</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  {/* Value */}
                  <Input
                    label="Valor do Contrato (R$)"
                    type="number"
                    step="0.01"
                    placeholder="0.00"
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    className="text-xs py-2 h-10"
                    required
                  />

                  {/* Duration */}
                  <Input
                    label="Duração (em meses)"
                    type="number"
                    placeholder="Ex: 12"
                    value={duration}
                    onChange={(e) => setDuration(e.target.value)}
                    className="text-xs py-2 h-10"
                    required
                  />

                  {/* Start Date via Popover + Calendar (DatePicker) */}
                  <div className="w-full flex flex-col gap-1.5">
                    <label className="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
                      Data de Início
                    </label>
                    <Popover>
                      <PopoverTrigger render={
                        <Button
                          variant="outline"
                          className={cn(
                            "w-full justify-start text-left font-normal bg-zinc-900 border-zinc-850 text-zinc-300 text-xs h-10 rounded-lg",
                            !startIn && "text-zinc-500"
                          )}
                        >
                          <CalendarIcon className="mr-2 h-4 w-4 text-zinc-500" />
                          {dateValue ? format(dateValue, "PPP", { locale: ptBR }) : <span>Selecione a data</span>}
                        </Button>
                      } />
                      <PopoverContent className="w-auto p-0 bg-zinc-950 border-zinc-800" align="start">
                        <Calendar
                          mode="single"
                          selected={dateValue}
                          onSelect={(date) => setStartIn(date ? date.toISOString() : '')}
                          locale={ptBR}
                        />
                      </PopoverContent>
                    </Popover>
                  </div>
                </div>
              </motion.div>
            )}

            {formError && (
              <span className="text-xs text-red-500 font-semibold bg-red-500/10 border border-red-500/20 px-3.5 py-2.5 rounded-xl">
                {formError}
              </span>
            )}

            <DialogFooter className="mt-4 gap-2 flex flex-row justify-end">
              <DialogClose render={
                <Button
                  type="button"
                  variant="outline"
                  className="border-zinc-800 text-zinc-300 rounded-xl"
                  onClick={() => setIsModalOpen(false)}
                >
                  Cancelar
                </Button>
              } />
              <Button
                type="submit"
                variant="default"
                className="rounded-xl font-bold px-6"
                isLoading={isSaving}
              >
                Salvar
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
};
