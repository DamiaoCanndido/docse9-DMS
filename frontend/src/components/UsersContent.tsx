'use client';

import React, { useState, useEffect, useTransition } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useDebounce } from 'use-debounce';
import { User, Municipality, Role, PaginatedResponse, UserPermission, DocumentType, PermissionLevel } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  createUser,
  updateUser,
  deleteUser,
  restoreUser,
  hardDeleteUser,
  getUserPermissions,
  updateUserPermissions,
} from '@/app/api/users';
import { toast } from 'sonner';
import {
  Users,
  Plus,
  Trash2,
  Edit2,
  RotateCcw,
  Trash,
  ChevronLeft,
  ChevronRight,
  Shield,
  Search,
  X,
  Sparkles,
  Building2,
  Mail,
  UserCheck,
  MoreHorizontal,
} from 'lucide-react';
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

  // Search input state
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
  const [isPermModalOpen, setIsPermModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  // User form states
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<Role>('COMMON');
  const [municipalityId, setMunicipalityId] = useState('');
  const [formError, setFormError] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  // Permissions state
  const [permissions, setPermissions] = useState<Record<DocumentType, PermissionLevel>>({
    NOTICE: 'NONE',
    DECREE: 'NONE',
    ORDINANCE: 'NONE',
    LAW: 'NONE',
    CONTRACT: 'NONE',
  });
  const [isLoadingPerms, setIsLoadingPerms] = useState(false);
  const [isSavingPerms, setIsSavingPerms] = useState(false);

  const currentPage = Number(searchParams.get('page')) || 1;
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

  // User CRUD
  const openCreateModal = () => {
    setSelectedUser(null);
    setUsername('');
    setEmail('');
    setPassword('');
    setRole('COMMON');
    setMunicipalityId(isMod ? currentUser.municipalityId : municipalities[0]?.id || '');
    setFormError('');
    setIsUserModalOpen(true);
  };

  const openEditModal = (user: User) => {
    setSelectedUser(user);
    setUsername(user.username);
    setEmail(user.email);
    setPassword('');
    setRole(user.role);
    setMunicipalityId(user.municipalityId);
    setFormError('');
    setIsUserModalOpen(true);
  };

  const handleSaveUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !email.trim()) {
      setFormError('Nome de usuário e e-mail são obrigatórios.');
      return;
    }
    if (!selectedUser && !password.trim()) {
      setFormError('A senha é obrigatória para novos usuários.');
      return;
    }

    setIsSaving(true);
    setFormError('');

    try {
      if (selectedUser) {
        await updateUser({
          id: selectedUser.id,
          input: {
            username,
            email,
            password: password ? password : undefined,
            role,
            municipalityId: isAdmin ? municipalityId : undefined,
          },
          path: pathname,
        });
        toast.success('Usuário atualizado com sucesso!');
      } else {
        await createUser({
          input: {
            username,
            email,
            password,
            role,
            municipalityId: isAdmin ? municipalityId : currentUser.municipalityId,
          },
          path: pathname,
        });
        toast.success('Usuário criado com sucesso!');
      }
      setIsUserModalOpen(false);
    } catch (err: any) {
      setFormError(err.response?.data?.error || 'Erro ao salvar usuário.');
      toast.error('Ocorreu um erro.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeleteUser = async (id: string) => {
    const confirm = window.confirm('Deseja realmente enviar este usuário para a lixeira?');
    if (!confirm) return;

    try {
      await deleteUser({ id, path: pathname });
      toast.success('Usuário enviado para a lixeira.');
    } catch {
      toast.error('Erro ao excluir usuário.');
    }
  };

  const handleRestoreUser = async (id: string) => {
    try {
      await restoreUser({ id, path: pathname });
      toast.success('Usuário restaurado com sucesso!');
    } catch {
      toast.error('Erro ao restaurar usuário.');
    }
  };

  const handleHardDeleteUser = async (id: string) => {
    const confirm = window.confirm(
      'ATENÇÃO: Esta ação é definitiva e removerá permanentemente o usuário. Confirmar?'
    );
    if (!confirm) return;

    try {
      await hardDeleteUser({ id, path: pathname });
      toast.success('Usuário excluído definitivamente.');
    } catch {
      toast.error('Erro ao excluir permanentemente.');
    }
  };

  // Permissions management
  const openPermissionsModal = async (user: User) => {
    setSelectedUser(user);
    setIsPermModalOpen(true);
    setIsLoadingPerms(true);

    try {
      const userPerms = await getUserPermissions(user.id);
      const permsMap = {
        NOTICE: 'NONE' as PermissionLevel,
        DECREE: 'NONE' as PermissionLevel,
        ORDINANCE: 'NONE' as PermissionLevel,
        LAW: 'NONE' as PermissionLevel,
        CONTRACT: 'NONE' as PermissionLevel,
      };
      
      userPerms.forEach((p) => {
        permsMap[p.documentType] = p.level;
      });

      setPermissions(permsMap);
    } catch {
      toast.error('Erro ao buscar permissões do usuário.');
      setIsPermModalOpen(false);
    } finally {
      setIsLoadingPerms(false);
    }
  };

  const handleSavePermissions = async () => {
    if (!selectedUser) return;
    setIsSavingPerms(true);

    try {
      const permsArray = (Object.keys(permissions) as DocumentType[]).map((type) => ({
        documentType: type,
        level: permissions[type],
      }));

      await updateUserPermissions({
        id: selectedUser.id,
        permissions: permsArray,
        path: pathname,
      });

      toast.success('Permissões atualizadas com sucesso!');
      setIsPermModalOpen(false);
    } catch {
      toast.error('Erro ao atualizar permissões.');
    } finally {
      setIsSavingPerms(false);
    }
  };

  const handlePermissionChange = (type: DocumentType, level: PermissionLevel) => {
    setPermissions((prev) => ({ ...prev, [type]: level }));
  };

  const docTypesList: { value: DocumentType; label: string }[] = [
    { value: 'NOTICE', label: 'Ofício (Notices)' },
    { value: 'DECREE', label: 'Decreto (Decrees)' },
    { value: 'ORDINANCE', label: 'Portaria (Ordinances)' },
    { value: 'LAW', label: 'Lei (Laws)' },
    { value: 'CONTRACT', label: 'Contrato (Contracts)' },
  ];

  const permLevels: { value: PermissionLevel; label: string }[] = [
    { value: 'NONE', label: 'Nenhum' },
    { value: 'READ', label: 'Visualizar' },
    { value: 'WRITE', label: 'Criar / Editar' },
    { value: 'DELETE', label: 'Excluir' },
  ];

  return (
    <div className="flex flex-col gap-8">
      {/* Header section */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-2 text-violet-400 text-xs font-semibold uppercase tracking-wider mb-1.5">
            <Sparkles className="w-4 h-4" />
            {isMod ? 'Administração do Município' : 'Administração Global'}
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white flex items-center gap-3">
            <Users className="w-8 h-8 text-violet-500" />
            {viewTrash ? 'Usuários Excluídos (Lixeira)' : 'Gestão de Usuários'}
          </h1>
          <p className="text-zinc-400 mt-1.5 text-sm max-w-xl">
            {viewTrash
              ? 'Visualize usuários deletados e restaure o acesso deles ou remova-os permanentemente.'
              : 'Gerencie as contas de acesso dos moderadores e funcionários públicos municipais.'}
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
        <div className="p-6 bg-zinc-950/60 border border-zinc-800/80 backdrop-blur-md rounded-2xl flex flex-col md:flex-row items-end gap-5 shadow-2xl relative overflow-hidden">
          <div className="absolute -top-24 -left-24 w-48 h-48 bg-violet-600/5 blur-3xl rounded-full pointer-events-none" />

          {/* Text Search */}
          <div className="w-full md:flex-1 relative">
            <Input
              label="Buscar por nome ou e-mail"
              placeholder="Digite o nome de usuário ou e-mail..."
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

          {/* Role Filter */}
          <div className="w-full md:w-48 flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-wider text-zinc-400">Perfil (Role)</label>
            <select
              className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 cursor-pointer"
              value={roleFilter}
              onChange={(e) => handleFilterChange('role', e.target.value)}
            >
              <option value="">Todos</option>
              {isAdmin && <option value="ADMIN">Administrador</option>}
              <option value="MOD">Moderador</option>
              <option value="COMMON">Funcionário Comum</option>
            </select>
          </div>

          {/* Municipality Filter (ADMIN only) */}
          {isAdmin && (
            <div className="w-full md:w-56 flex flex-col gap-1.5">
              <label className="text-xs font-semibold uppercase tracking-wider text-zinc-400">Município</label>
              <select
                className="w-full bg-zinc-900 border border-zinc-800 text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 cursor-pointer"
                value={municipalityFilter}
                onChange={(e) => handleFilterChange('municipalityId', e.target.value)}
              >
                <option value="">Todos</option>
                {municipalities.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name} ({m.uf})
                  </option>
                ))}
              </select>
            </div>
          )}

          <Button
            variant="outline"
            className="w-full md:w-auto h-11 border-zinc-800 hover:border-zinc-700 hover:bg-zinc-900 text-zinc-300 font-semibold px-5 rounded-xl"
            onClick={handleClearFilters}
          >
            Limpar Filtros
          </Button>
        </div>
      )}

      {/* Users table */}
      <div className="bg-zinc-950/40 border border-zinc-900 rounded-2xl overflow-hidden shadow-2xl relative">
        {isPending && (
          <div className="absolute inset-0 bg-black/60 backdrop-blur-[2px] z-10 flex items-center justify-center transition-all duration-300">
            <div className="flex flex-col items-center gap-3">
              <div className="w-10 h-10 rounded-full border-2 border-violet-500/20 border-t-violet-500 animate-spin" />
              <span className="text-violet-400 text-xs font-bold tracking-widest uppercase">
                Atualizando dados...
              </span>
            </div>
          </div>
        )}

        {initialData.data.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24 px-4 text-center">
            <div className="w-14 h-14 rounded-2xl bg-zinc-900/50 flex items-center justify-center text-zinc-500 mb-5 border border-zinc-800">
              <Users className="w-7 h-7 text-zinc-600" />
            </div>
            <h3 className="text-xl font-bold text-white">Nenhum usuário encontrado</h3>
            <p className="text-zinc-500 text-sm mt-2 max-w-sm">
              Tente redefinir os filtros ou buscar por outros termos.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-zinc-900 bg-zinc-950/70 text-zinc-400 text-[11px] font-bold uppercase tracking-wider">
                  <th className="px-6 py-4.5">Usuário</th>
                  <th className="px-6 py-4.5">Perfil</th>
                  <th className="px-6 py-4.5">Município</th>
                  <th className="px-6 py-4.5">Último Acesso</th>
                  <th className="px-6 py-4.5 text-center">Ações</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-900 text-sm text-zinc-300">
                {initialData.data.map((usr) => (
                  <tr key={usr.id} className="hover:bg-zinc-900/10 transition-colors group">
                    <td className="px-6 py-4.5">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-400 font-semibold uppercase text-xs">
                          {usr.username.slice(0, 2)}
                        </div>
                        <div className="flex flex-col">
                          <span className="font-bold text-white group-hover:text-violet-400 transition-colors">
                            {usr.username}
                          </span>
                          <span className="text-[10px] text-zinc-500 font-medium flex items-center gap-1 mt-0.5">
                            <Mail className="w-3 h-3 text-zinc-600" />
                            {usr.email}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4.5">
                      <span
                        className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold border ${
                          usr.role === 'ADMIN'
                            ? 'bg-red-500/5 text-red-400 border-red-500/20'
                            : usr.role === 'MOD'
                            ? 'bg-violet-500/5 text-violet-400 border-violet-500/20'
                            : 'bg-zinc-900/80 text-zinc-400 border-zinc-800'
                        }`}
                      >
                        {usr.role}
                      </span>
                    </td>
                    <td className="px-6 py-4.5">
                      {usr.municipality ? (
                        <div className="flex flex-col">
                          <span className="font-semibold text-zinc-300 flex items-center gap-1.5">
                            <Building2 className="w-3.5 h-3.5 text-zinc-500" />
                            {usr.municipality.name}
                          </span>
                          <span className="text-[10px] text-zinc-500 ml-5">
                            Estado: {usr.municipality.uf}
                          </span>
                        </div>
                      ) : (
                        <span className="text-zinc-650">—</span>
                      )}
                    </td>
                    <td className="px-6 py-4.5 text-zinc-500 text-xs font-medium">
                      {usr.lastLogin ? (
                        <div className="flex items-center gap-1.5">
                          <UserCheck className="w-3.5 h-3.5 text-zinc-600" />
                          {new Date(usr.lastLogin).toLocaleDateString('pt-BR')} às{' '}
                          {new Date(usr.lastLogin).toLocaleTimeString('pt-BR', {
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </div>
                      ) : (
                        <span className="text-zinc-700 italic">Nunca acessou</span>
                      )}
                    </td>
                    <td className="px-6 py-4.5 text-center">
                      <DropdownMenu>
                        <DropdownMenuTrigger render={
                          <button className="p-1.5 text-zinc-400 hover:text-white hover:bg-zinc-900 rounded-lg transition-all cursor-pointer">
                            <MoreHorizontal className="w-4.5 h-4.5" />
                          </button>
                        } />
                        <DropdownMenuContent align="end" className="bg-zinc-950 border-zinc-800 text-zinc-300 min-w-[170px]">
                          {viewTrash ? (
                            <>
                              <DropdownMenuItem
                                onClick={() => handleRestoreUser(usr.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-emerald-400 cursor-pointer focus:bg-zinc-900 focus:text-emerald-400 p-2 text-xs font-medium"
                              >
                                <RotateCcw className="w-4.5 h-4.5" />
                                Restaurar
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleHardDeleteUser(usr.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-red-400 cursor-pointer focus:bg-zinc-900 focus:text-red-400 p-2 text-xs font-medium"
                              >
                                <Trash className="w-4.5 h-4.5" />
                                Excluir Definitivamente
                              </DropdownMenuItem>
                            </>
                          ) : (
                            <>
                              {usr.role === 'COMMON' && (
                                <DropdownMenuItem
                                  onClick={() => openPermissionsModal(usr)}
                                  className="flex items-center gap-2 hover:bg-zinc-900 hover:text-amber-400 cursor-pointer focus:bg-zinc-900 focus:text-amber-400 p-2 text-xs font-medium"
                                >
                                  <Shield className="w-4.5 h-4.5" />
                                  Permissões
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuItem
                                onClick={() => openEditModal(usr)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-violet-400 cursor-pointer focus:bg-zinc-900 focus:text-violet-400 p-2 text-xs font-medium"
                              >
                                <Edit2 className="w-4.5 h-4.5" />
                                Editar
                              </DropdownMenuItem>
                              {usr.id !== currentUser.id && (
                                <DropdownMenuItem
                                  onClick={() => handleDeleteUser(usr.id)}
                                  className="flex items-center gap-2 hover:bg-zinc-900 hover:text-red-400 cursor-pointer focus:bg-zinc-900 focus:text-red-400 p-2 text-xs font-medium"
                                >
                                  <Trash2 className="w-4.5 h-4.5" />
                                  Mover para Lixeira
                                </DropdownMenuItem>
                              )}
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {initialData.total > initialData.pageSize && (
          <div className="px-6 py-4.5 border-t border-zinc-900 bg-zinc-950/20 flex flex-col sm:flex-row items-center justify-between gap-4">
            <span className="text-xs text-zinc-500 font-semibold uppercase tracking-wider">
              Mostrando {initialData.data.length} de {initialData.total} usuários
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

      {/* User Create/Edit Modal via shadcn/ui Dialog */}
      <Dialog open={isUserModalOpen} onOpenChange={setIsUserModalOpen}>
        <DialogContent className="bg-zinc-950 border-zinc-800/80 text-zinc-300 max-w-lg rounded-2xl shadow-2xl p-6">
          <DialogHeader>
            <DialogTitle className="text-xl font-bold text-white">
              {selectedUser ? 'Editar Usuário' : 'Cadastrar Novo Usuário'}
            </DialogTitle>
            <DialogDescription className="text-zinc-500 text-xs mt-1">
              Configure as credenciais e o escopo de acesso do usuário.
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSaveUser} className="flex flex-col gap-5 mt-4">
            <Input
              label="Nome de Usuário (Username)"
              placeholder="Ex: joao.silva"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />

            <Input
              label="E-mail"
              type="email"
              placeholder="Ex: joao@prefeitura.gov.br"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />

            <Input
              label={selectedUser ? 'Senha (deixe em branco para não alterar)' : 'Senha de Acesso'}
              type="password"
              placeholder="Mínimo 6 caracteres"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required={!selectedUser}
            />

            {/* Perfil Selection via shadcn/ui Select */}
            <div className="w-full flex flex-col gap-1.5">
              <label className="text-sm font-medium text-zinc-400">Perfil de Acesso</label>
              <Select value={role} onValueChange={(val) => setRole((val as Role) || 'COMMON')}>
                <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-300 text-sm h-10">
                  <SelectValue placeholder="Selecione o perfil" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-950 border-zinc-800 text-zinc-300">
                  {isAdmin && <SelectItem value="ADMIN">Administrador Global</SelectItem>}
                  <SelectItem value="MOD">Moderador Municipal</SelectItem>
                  <SelectItem value="COMMON">Funcionário Comum</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Municipality Selection (ADMIN only) via shadcn/ui Select */}
            {isAdmin && (
              <div className="w-full flex flex-col gap-1.5">
                <label className="text-sm font-medium text-zinc-400">Município Vinculado</label>
                <Select value={municipalityId} onValueChange={(val) => setMunicipalityId(val || '')}>
                  <SelectTrigger className="w-full bg-zinc-900 border-zinc-800 text-zinc-300 text-sm h-10">
                    <SelectValue placeholder="Selecione o município" />
                  </SelectTrigger>
                  <SelectContent className="bg-zinc-950 border-zinc-800 text-zinc-300">
                    {municipalities.map((m) => (
                      <SelectItem key={m.id} value={m.id}>
                        {m.name} ({m.uf})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {isMod && (
              <div className="p-3 bg-zinc-900/50 border border-zinc-850 rounded-xl flex items-center gap-2">
                <Building2 className="w-4.5 h-4.5 text-zinc-500" />
                <span className="text-xs text-zinc-400">
                  Vinculado a: <strong>{currentUser.municipality?.name} ({currentUser.municipality?.uf})</strong>
                </span>
              </div>
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
                  onClick={() => setIsUserModalOpen(false)}
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

      {/* User Permissions Modal via shadcn/ui Dialog */}
      <Dialog open={isPermModalOpen} onOpenChange={setIsPermModalOpen}>
        <DialogContent className="bg-zinc-950 border-zinc-800/80 text-zinc-300 max-w-2xl rounded-2xl shadow-2xl p-6">
          <DialogHeader>
            <DialogTitle className="text-xl font-bold text-white flex items-center gap-2">
              <Shield className="w-5 h-5 text-amber-500" />
              Permissões do Usuário: {selectedUser?.username}
            </DialogTitle>
            <DialogDescription className="text-zinc-500 text-xs mt-1">
              Configure o nível de permissão que o funcionário comum terá para cada tipo de documento oficial.
            </DialogDescription>
          </DialogHeader>

          {isLoadingPerms ? (
            <div className="py-12 flex flex-col items-center justify-center gap-3">
              <div className="w-8 h-8 rounded-full border-2 border-amber-500/20 border-t-amber-500 animate-spin" />
              <span className="text-zinc-500 text-xs">Carregando permissões...</span>
            </div>
          ) : (
            <div className="flex flex-col gap-6 mt-4">
              <div className="border border-zinc-900 rounded-xl overflow-hidden divide-y divide-zinc-900 bg-zinc-950/20">
                <div className="grid grid-cols-2 md:grid-cols-5 p-3.5 bg-zinc-950 text-xs font-bold text-zinc-500 uppercase tracking-wider">
                  <div className="col-span-2">Tipo de Documento</div>
                  <div className="col-span-3">Nível de Acesso</div>
                </div>

                {docTypesList.map((type) => (
                  <div
                    key={type.value}
                    className="grid grid-cols-2 md:grid-cols-5 p-3.5 items-center gap-4 hover:bg-zinc-900/5 transition-colors"
                  >
                    <div className="col-span-2 font-semibold text-zinc-200">
                      {type.label}
                    </div>
                    <div className="col-span-3 flex items-center gap-1.5 md:gap-3 flex-wrap">
                      {permLevels.map((lvl) => (
                        <button
                          key={lvl.value}
                          type="button"
                          onClick={() => handlePermissionChange(type.value, lvl.value)}
                          className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
                            permissions[type.value] === lvl.value
                              ? 'bg-amber-500/10 border border-amber-500/35 text-amber-400'
                              : 'bg-zinc-900 border border-zinc-800 text-zinc-500 hover:text-zinc-300'
                          }`}
                        >
                          {lvl.label}
                        </button>
                      ))}
                    </div>
                  </div>
                ))}
              </div>

              <DialogFooter className="gap-2 flex flex-row justify-end">
                <DialogClose render={
                  <Button
                    type="button"
                    variant="outline"
                    className="border-zinc-800 text-zinc-300 rounded-xl"
                    onClick={() => setIsPermModalOpen(false)}
                  >
                    Cancelar
                  </Button>
                } />
                <Button
                  type="button"
                  variant="default"
                  className="rounded-xl font-bold bg-amber-500/20 border border-amber-500/35 text-amber-400 hover:bg-amber-500/30 hover:text-amber-300 shadow-md shadow-amber-500/5 px-6"
                  onClick={handleSavePermissions}
                  isLoading={isSavingPerms}
                >
                  Salvar Permissões
                </Button>
              </DialogFooter>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};
