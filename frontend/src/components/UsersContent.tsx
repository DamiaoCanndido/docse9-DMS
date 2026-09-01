'use client';

import React, { useState, useEffect, useTransition, useCallback } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useDebounce } from 'use-debounce';
import { User, Municipality, Role, PaginatedResponse } from '@/types';
import { Button } from '@/components/ui/button';
import {
  createUser,
  updateUser,
  deleteUser,
  restoreUser,
  hardDeleteUser,
} from '@/app/api/users';
import { toast } from 'sonner';
import { isRedirectError } from '@/lib/utils';
import { Users, Plus, Sparkles } from 'lucide-react';
import { UserFilterBar } from './users/UserFilterBar';
import { UserTable } from './users/UserTable';
import { UserFormDialog } from './users/UserFormDialog';
import { UserPermissionsMatrix } from './users/UserPermissionsMatrix';
import { UserCreatedSuccessDialog } from './users/UserCreatedSuccessDialog';
import { UserDeleteDialog, UserDeleteMode } from './users/UserDeleteDialog';

interface UsersContentProps {
  initialData: PaginatedResponse<User>;
  currentUser: User;
  viewTrash: boolean;
  municipalities: Municipality[];
}

export const UsersContent: React.FC<UsersContentProps> = ({
  initialData,
  currentUser,
  viewTrash,
  municipalities,
}) => {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();

  // Search input state with debounce
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
    params.set('page', '1');
    startTransition(() => {
      router.replace(`?${params.toString()}`);
    });
  }, [debouncedSearch, router]);

  // Modals state
  const [isUserModalOpen, setIsUserModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const [isPermModalOpen, setIsPermModalOpen] = useState(false);
  const [selectedPermUser, setSelectedPermUser] = useState<User | null>(null);

  const [isSuccessModalOpen, setIsSuccessModalOpen] = useState(false);
  const [createdUserCredentials, setCreatedUserCredentials] = useState<{
    username: string;
    email: string;
    randomPassword?: string;
  } | null>(null);

  // Delete dialog state
  const [deleteDialogState, setDeleteDialogState] = useState<{
    isOpen: boolean;
    mode: UserDeleteMode;
    user: User | null;
  }>({
    isOpen: false,
    mode: 'delete',
    user: null,
  });
  const [isDeleteLoading, setIsDeleteLoading] = useState(false);

  const roleFilter = searchParams.get('role') || '';
  const municipalityFilter = searchParams.get('municipalityId') || '';

  const isAdmin = currentUser.role === 'ADMIN';
  const isMod = currentUser.role === 'MOD';

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

  // User CRUD Handlers
  const openCreateModal = () => {
    setSelectedUser(null);
    setIsUserModalOpen(true);
  };

  const openEditModal = (user: User) => {
    setSelectedUser(user);
    setIsUserModalOpen(true);
  };

  const openPermissionsModal = (user: User) => {
    setSelectedPermUser(user);
    setIsPermModalOpen(true);
  };

  const openDeleteDialog = (user: User, mode: UserDeleteMode) => {
    setDeleteDialogState({
      isOpen: true,
      mode,
      user,
    });
  };

  const closeDeleteDialog = () => {
    setDeleteDialogState((prev) => ({
      ...prev,
      isOpen: false,
      user: null,
    }));
  };

  const handleSaveUser = async (payload: {
    isEdit: boolean;
    id?: string;
    username: string;
    email: string;
    password?: string;
    role: Role;
    municipalityId: string;
  }) => {
    try {
      if (payload.isEdit && payload.id) {
        await updateUser({
          id: payload.id,
          input: {
            username: payload.username,
            email: payload.email,
            password: payload.password ? payload.password : undefined,
            role: payload.role,
            municipalityId: isAdmin ? payload.municipalityId : undefined,
          },
          path: pathname,
        });
        toast.success('Usuário atualizado com sucesso!');
      } else {
        const res = await createUser({
          input: {
            username: payload.username,
            email: payload.email,
            role: payload.role,
            municipalityId: isAdmin ? payload.municipalityId : currentUser.municipalityId,
          },
          path: pathname,
        });
        toast.success('Usuário criado com sucesso!');

        if (res.randomPassword) {
          setCreatedUserCredentials({
            username: res.user?.username || payload.username,
            email: res.user?.email || payload.email,
            randomPassword: res.randomPassword,
          });
          setIsSuccessModalOpen(true);
        }
      }
    } catch (err: unknown) {
      if (isRedirectError(err)) {
        throw err;
      }
      const errorMsg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Erro ao salvar usuário.';
      toast.error(errorMsg);
      throw err;
    }
  };

  const handleConfirmDelete = useCallback(async () => {
    if (!deleteDialogState.user) return;
    const targetId = deleteDialogState.user.id;
    setIsDeleteLoading(true);

    try {
      if (deleteDialogState.mode === 'delete') {
        await deleteUser({ id: targetId, path: pathname });
        toast.success('Usuário enviado para a lixeira.');
      } else if (deleteDialogState.mode === 'restore') {
        await restoreUser({ id: targetId, path: pathname });
        toast.success('Usuário restaurado com sucesso!');
      } else if (deleteDialogState.mode === 'hardDelete') {
        await hardDeleteUser({ id: targetId, path: pathname });
        toast.success('Usuário excluído definitivamente.');
      }
      closeDeleteDialog();
    } catch (err: unknown) {
      if (isRedirectError(err)) {
        throw err;
      }
      toast.error('Ocorreu um erro ao executar a ação.');
    } finally {
      setIsDeleteLoading(false);
    }
  }, [deleteDialogState, pathname]);

  return (
    <div className="flex flex-col gap-8">
      {/* Header section */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-2 text-teal-600 dark:text-teal-400 text-xs font-semibold uppercase tracking-wider mb-1.5">
            <Sparkles className="w-4 h-4" />
            {isMod ? 'Administração do Município' : 'Administração Global'}
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-foreground flex items-center gap-3">
            <Users className="w-8 h-8 text-teal-600 dark:text-teal-500" />
            {viewTrash ? 'Usuários Excluídos (Lixeira)' : 'Gestão de Usuários'}
          </h1>
          <p className="text-muted-foreground mt-1.5 text-sm max-w-xl">
            {viewTrash
              ? 'Visualize usuários deletados e restaure o acesso deles ou remova-os permanentemente.'
              : 'Gerencie as contas de acesso dos moderadores e funcionários públicos municipais.'}
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

          {!viewTrash && (
            <Button
              variant="default"
              className="flex items-center gap-2 font-semibold px-5 rounded-xl h-11"
              onClick={openCreateModal}
            >
              <Plus className="w-4.5 h-4.5" />
              Novo Usuário
            </Button>
          )}
        </div>
      </div>

      {/* Filter panel */}
      {!viewTrash && (
        <UserFilterBar
          searchValue={searchVal}
          onSearchChange={setSearchVal}
          roleFilter={roleFilter}
          onRoleChange={(val) => handleFilterChange('role', val)}
          municipalityFilter={municipalityFilter}
          onMunicipalityChange={(val) => handleFilterChange('municipalityId', val)}
          municipalities={municipalities}
          isAdmin={isAdmin}
          onClearFilters={handleClearFilters}
        />
      )}

      {/* Users table */}
      <UserTable
        users={initialData.data}
        currentUser={currentUser}
        viewTrash={viewTrash}
        isPending={isPending}
        onEdit={openEditModal}
        onDelete={(u) => openDeleteDialog(u, 'delete')}
        onRestore={(u) => openDeleteDialog(u, 'restore')}
        onHardDelete={(u) => openDeleteDialog(u, 'hardDelete')}
        onPermissions={openPermissionsModal}
        pagination={{
          page: initialData.page,
          pageSize: initialData.pageSize,
          total: initialData.total,
        }}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
      />

      {/* User Create / Edit Dialog */}
      <UserFormDialog
        isOpen={isUserModalOpen}
        onClose={() => setIsUserModalOpen(false)}
        selectedUser={selectedUser}
        currentUser={currentUser}
        municipalities={municipalities}
        onSave={handleSaveUser}
      />

      {/* User Permissions Matrix Dialog */}
      <UserPermissionsMatrix
        isOpen={isPermModalOpen}
        onClose={() => setIsPermModalOpen(false)}
        user={selectedPermUser}
        path={pathname}
      />

      {/* User Created Success (Random Password display) Dialog */}
      <UserCreatedSuccessDialog
        isOpen={isSuccessModalOpen}
        onClose={() => setIsSuccessModalOpen(false)}
        credentials={createdUserCredentials}
      />

      {/* User Delete / Restore / HardDelete Confirmation Dialog */}
      <UserDeleteDialog
        isOpen={deleteDialogState.isOpen}
        onClose={closeDeleteDialog}
        mode={deleteDialogState.mode}
        user={deleteDialogState.user}
        onConfirm={handleConfirmDelete}
        isLoading={isDeleteLoading}
      />
    </div>
  );
};
