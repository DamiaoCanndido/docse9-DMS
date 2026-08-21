'use client';

import React, { useState, useEffect, useTransition, useMemo } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useDebounce } from 'use-debounce';
import { Document, DocumentType, ContractType, User, PaginatedResponse, UserPermission } from '@/types';
import { Button } from '@/components/ui/button';
import { getUserPermissions } from '@/app/api/users';
import { useDocumentActions } from '@/hooks/useDocumentActions';
import { DocumentTable } from './documents/DocumentTable';
import { DocumentFilterBar } from './documents/DocumentFilterBar';
import { DocumentFormDialog, DocumentTypeOption } from './documents/DocumentFormDialog';
import { DocumentDeleteDialog } from './documents/DocumentDeleteDialog';
import { 
  FileText, 
  Sparkles,
  FileCheck,
  Plus,
  Clock,
  Scale,
  FileSignature,
  ShieldAlert,
  Building2
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface DocumentsContentProps {
  initialData: PaginatedResponse<Document>;
  currentUser: User;
  viewTrash?: boolean;
}

const docTypesList: DocumentTypeOption[] = [
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

export const DocumentsContent: React.FC<DocumentsContentProps> = ({ 
  initialData, 
  currentUser,
  viewTrash = false
}) => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();

  // Search input state with debounce
  const [searchVal, setSearchVal] = useState(searchParams.get('search') || '');
  const [debouncedSearch] = useDebounce(searchVal, 400);

  // User permissions (for COMMON role)
  const [userPermissions, setUserPermissions] = useState<UserPermission[]>([]);
  const [isLoadedPermissions, setIsLoadedPermissions] = useState(currentUser.role !== 'COMMON');

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
  const availableYears = useMemo(
    () => Array.from({ length: currentYear - 2022 + 1 }, (_, i) => (currentYear - i).toString()).concat('all'),
    [currentYear]
  );

  // Document actions hook
  const {
    isMutating,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
    confirmDeleteDialogAction,
    handleCreate,
    handleUpdate,
  } = useDocumentActions();

  // Form modal state
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingDocument, setEditingDocument] = useState<Document | null>(null);

  // Permission helpers
  const canCreate = (t: DocumentType): boolean => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      const p = userPermissions.find((perm) => perm.documentType === t);
      return !!p && (p.level === 'WRITE' || p.level === 'DELETE');
    }
    return false;
  };

  const canEdit = (t: DocumentType): boolean => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      const p = userPermissions.find((perm) => perm.documentType === t);
      return !!p && (p.level === 'WRITE' || p.level === 'DELETE');
    }
    return false;
  };

  const canDelete = (t: DocumentType): boolean => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      const p = userPermissions.find((perm) => perm.documentType === t);
      return !!p && p.level === 'DELETE';
    }
    return false;
  };

  const canViewTab = (t: DocumentType): boolean => {
    if (currentUser.role === 'MOD') return true;
    if (currentUser.role === 'COMMON') {
      if (userPermissions.length === 0) return true;
      const p = userPermissions.find((perm) => perm.documentType === t);
      return !!p && p.level !== 'NONE';
    }
    return false;
  };

  const hasAnyWritePermission = 
    currentUser.role === 'MOD' || 
    userPermissions.some((p) => p.level === 'WRITE' || p.level === 'DELETE');

  const activeTabMeta = docTypesList.find((t) => t.value === activeTab) || docTypesList[0];

  // URL Filter dispatchers
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

  const openCreateModal = () => {
    setEditingDocument(null);
    setIsModalOpen(true);
  };

  const openEditModal = (doc: Document) => {
    setEditingDocument(doc);
    setIsModalOpen(true);
  };

  const handleFormSave = async ({
    isEdit,
    id,
    createInput,
    updateInput,
  }: {
    isEdit: boolean;
    id?: string;
    createInput?: Parameters<typeof handleCreate>[0];
    updateInput?: Parameters<typeof handleUpdate>[1];
  }) => {
    if (isEdit && id && updateInput) {
      await handleUpdate(id, updateInput);
    } else if (createInput) {
      await handleCreate(createInput);
    }
  };

  if (hasNoPermissions) {
    return (
      <div className="flex flex-col items-center justify-center py-20 px-4 text-center max-w-2xl mx-auto">
        <div className="w-20 h-20 rounded-3xl bg-teal-600/10 border border-teal-500/20 flex items-center justify-center text-teal-600 dark:text-teal-400 mb-6 shadow-xl shadow-teal-500/5">
          <ShieldAlert className="w-10 h-10" />
        </div>
        <h2 className="text-2xl font-extrabold text-foreground tracking-tight">
          Acesso Pendente de Liberação
        </h2>
        <p className="text-muted-foreground text-sm mt-3 leading-relaxed">
          Olá, <strong className="text-foreground">{currentUser.username}</strong>! Sua conta foi registrada com sucesso no sistema, porém você ainda não possui permissões atribuídas para visualizar ou gerenciar documentos oficiais.
        </p>
        <div className="mt-6 p-5 bg-card border border-border rounded-2xl text-xs text-muted-foreground text-left flex flex-col gap-2.5 w-full shadow-xl">
          <span className="font-semibold text-foreground flex items-center gap-2 text-sm">
            <Building2 className="w-4 h-4 text-teal-600 dark:text-teal-400" />
            Como solicitar acesso?
          </span>
          <p className="leading-relaxed">
            Entre em contato com um <strong className="text-teal-600 dark:text-teal-400">Moderador</strong> ou <strong className="text-teal-600 dark:text-teal-400">Administrador</strong> do município de <strong className="text-foreground">{currentUser.municipality?.name || 'sua prefeitura'}</strong> para que sejam concedidas as permissões de leitura/escrita nos tipos de documentos desejados (Ofícios, Decretos, Portarias, Leis ou Contratos).
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
          <div className="flex items-center gap-2 text-teal-600 dark:text-teal-400 text-xs font-semibold uppercase tracking-wider mb-1.5">
            <Sparkles className="w-4 h-4" />
            Painel de Documentos Municipais
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-foreground flex items-center gap-3">
            <FileText className="w-8 h-8 text-teal-600 dark:text-teal-500" />
            {viewTrash ? 'Documentos Excluídos (Lixeira)' : 'Documentos Oficiais'}
          </h1>
          <p className="text-muted-foreground mt-1.5 text-sm max-w-xl">
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
            className="border-border hover:bg-muted text-foreground font-semibold px-4.5 rounded-xl h-11"
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
      <div className="flex items-center gap-1.5 p-1.5 bg-card/80 border border-border rounded-2xl overflow-x-auto scrollbar-none shadow-md">
        {docTypesList.map((tab) => {
          const isActive = activeTab === tab.value;
          const Icon = tab.icon || FileText;
          const allowed = canViewTab(tab.value);
          if (!allowed) return null;

          return (
            <button
              key={tab.value}
              onClick={() => handleTabChange(tab.value)}
              className={cn(
                "flex items-center gap-2.5 px-5 py-3 rounded-xl text-sm font-semibold transition-all duration-200 cursor-pointer whitespace-nowrap select-none relative",
                isActive
                  ? "bg-teal-600 text-white shadow-md shadow-teal-600/25 font-bold"
                  : "text-muted-foreground hover:text-foreground hover:bg-muted/60"
              )}
            >
              <Icon className={cn("w-4 h-4", isActive ? "text-white" : "text-muted-foreground")} />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Filter panel */}
      {!viewTrash && (
        <DocumentFilterBar
          activeTab={activeTab}
          activeTabLabel={activeTabMeta.label}
          searchValue={searchVal}
          onSearchChange={setSearchVal}
          yearFilter={yearFilter}
          onYearChange={(val) => handleFilterChange('year', val)}
          availableYears={availableYears}
          contractTypeFilter={contractTypeFilter}
          onContractTypeChange={(val) => handleFilterChange('contractType', val)}
          onClearFilters={handleClearFilters}
        />
      )}

      {/* Documents table */}
      <DocumentTable
        documents={initialData.data}
        activeTab={activeTab}
        activeTabLabel={activeTabMeta.label}
        viewTrash={viewTrash}
        currentUser={currentUser}
        isPending={isPending}
        canEdit={canEdit}
        canDelete={canDelete}
        onEdit={openEditModal}
        onDelete={(doc) => openDeleteDialog(doc, 'delete')}
        onRestore={(doc) => openDeleteDialog(doc, 'restore')}
        onHardDelete={(doc) => openDeleteDialog(doc, 'hardDelete')}
        pagination={{
          page: initialData.page,
          pageSize: initialData.pageSize,
          total: initialData.total,
        }}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        searchVal={searchVal}
      />

      {/* Create / Edit Document Dialog */}
      <DocumentFormDialog
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        editingDocument={editingDocument}
        activeTab={activeTab}
        canCreate={canCreate}
        docTypesList={docTypesList}
        contractTypeLabels={contractTypeLabels}
        onSave={handleFormSave}
        creatorId={currentUser.id}
        municipalityId={currentUser.municipalityId}
      />

      {/* Confirmation Delete / Restore / HardDelete Dialog */}
      <DocumentDeleteDialog
        isOpen={deleteDialogState.isOpen}
        onClose={closeDeleteDialog}
        mode={deleteDialogState.mode}
        document={deleteDialogState.document}
        onConfirm={confirmDeleteDialogAction}
        isLoading={isMutating}
      />
    </div>
  );
};
