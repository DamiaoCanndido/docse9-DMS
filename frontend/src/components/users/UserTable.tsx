'use client';

import React from 'react';
import { User } from '@/types';
import { PaginationBar } from '@/components/ui/PaginationBar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Users,
  Building2,
  Mail,
  UserCheck,
  MoreHorizontal,
  Shield,
  Edit2,
  Trash2,
  RotateCcw,
  Trash,
} from 'lucide-react';
import { formatDateTime } from '@/lib/date';

interface UserTableProps {
  users: User[];
  currentUser: User;
  viewTrash: boolean;
  isPending: boolean;
  onEdit: (user: User) => void;
  onDelete: (user: User) => void;
  onRestore: (user: User) => void;
  onHardDelete: (user: User) => void;
  onPermissions: (user: User) => void;
  pagination: {
    page: number;
    pageSize: number;
    total: number;
  };
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

export const UserTable: React.FC<UserTableProps> = ({
  users,
  currentUser,
  viewTrash,
  isPending,
  onEdit,
  onDelete,
  onRestore,
  onHardDelete,
  onPermissions,
  pagination,
  onPageChange,
  onPageSizeChange,
}) => {
  return (
    <div className="bg-card border border-border rounded-2xl overflow-hidden shadow-xl relative">
      {isPending && (
        <div className="absolute inset-0 bg-background/60 backdrop-blur-[2px] z-10 flex items-center justify-center transition-all duration-300">
          <div className="flex flex-col items-center gap-3">
            <div className="w-10 h-10 rounded-full border-2 border-teal-500/20 border-t-teal-500 animate-spin" />
            <span className="text-teal-600 dark:text-teal-400 text-xs font-bold tracking-widest uppercase">
              Atualizando dados...
            </span>
          </div>
        </div>
      )}

      {users.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 px-4 text-center">
          <div className="w-14 h-14 rounded-2xl bg-muted/50 flex items-center justify-center text-muted-foreground mb-5 border border-border">
            <Users className="w-7 h-7 text-muted-foreground" />
          </div>
          <h3 className="text-xl font-bold text-foreground">Nenhum usuário encontrado</h3>
          <p className="text-muted-foreground text-sm mt-2 max-w-sm">
            Tente redefinir os filtros ou buscar por outros termos.
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-border bg-muted/50 text-muted-foreground text-[11px] font-bold uppercase tracking-wider">
                <th className="px-6 py-4.5">Usuário</th>
                <th className="px-6 py-4.5">Perfil</th>
                <th className="px-6 py-4.5">Município</th>
                <th className="px-6 py-4.5">Último Acesso</th>
                <th className="px-6 py-4.5 text-center">Ações</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border text-sm text-foreground">
              {users.map((usr) => (
                <tr key={usr.id} className="hover:bg-muted/40 transition-colors group">
                  <td className="px-6 py-4.5">
                    <div className="flex items-center gap-3">
                      <div className="w-9 h-9 rounded-full bg-muted border border-border flex items-center justify-center text-muted-foreground font-semibold uppercase text-xs">
                        {usr.username.slice(0, 2)}
                      </div>
                      <div className="flex flex-col">
                        <span className="font-bold text-foreground group-hover:text-teal-600 dark:group-hover:text-teal-400 transition-colors">
                          {usr.username}
                        </span>
                        <span className="text-[10px] text-muted-foreground font-medium flex items-center gap-1 mt-0.5">
                          <Mail className="w-3 h-3 text-muted-foreground" />
                          {usr.email}
                        </span>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4.5">
                    <span
                      className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold border ${
                        usr.role === 'ADMIN'
                          ? 'bg-red-500/10 text-red-500 dark:text-red-400 border-red-500/20'
                          : usr.role === 'MOD'
                          ? 'bg-teal-500/10 text-teal-600 dark:text-teal-400 border-teal-500/20'
                          : 'bg-muted text-muted-foreground border-border'
                      }`}
                    >
                      {usr.role}
                    </span>
                  </td>
                  <td className="px-6 py-4.5">
                    {usr.municipality ? (
                      <div className="flex flex-col">
                        <span className="font-semibold text-foreground flex items-center gap-1.5">
                          <Building2 className="w-3.5 h-3.5 text-muted-foreground" />
                          {usr.municipality.name}
                        </span>
                        <span className="text-[10px] text-muted-foreground ml-5">
                          Estado: {usr.municipality.uf}
                        </span>
                      </div>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </td>
                  <td className="px-6 py-4.5 text-muted-foreground text-xs font-medium">
                    {usr.lastLogin ? (
                      <div className="flex items-center gap-1.5">
                        <UserCheck className="w-3.5 h-3.5 text-muted-foreground" />
                        {formatDateTime(usr.lastLogin)}
                      </div>
                    ) : (
                      <span className="text-muted-foreground italic">Nunca acessou</span>
                    )}
                  </td>
                  <td className="px-6 py-4.5 text-center">
                    <DropdownMenu>
                      <DropdownMenuTrigger render={
                        <button className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg transition-all cursor-pointer">
                          <MoreHorizontal className="w-4.5 h-4.5" />
                        </button>
                      } />
                      <DropdownMenuContent
                        align="end"
                        className="bg-popover border-border text-popover-foreground min-w-[170px]"
                      >
                        {viewTrash ? (
                          <>
                            <DropdownMenuItem
                              onClick={() => onRestore(usr)}
                              className="flex items-center gap-2 hover:bg-muted hover:text-emerald-500 cursor-pointer focus:bg-muted focus:text-emerald-500 p-2 text-xs font-medium"
                            >
                              <RotateCcw className="w-4.5 h-4.5" />
                              Restaurar
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => onHardDelete(usr)}
                              className="flex items-center gap-2 hover:bg-muted hover:text-red-500 cursor-pointer focus:bg-muted focus:text-red-500 p-2 text-xs font-medium"
                            >
                              <Trash className="w-4.5 h-4.5" />
                              Excluir Definitivamente
                            </DropdownMenuItem>
                          </>
                        ) : (
                          <>
                            {usr.role === 'COMMON' && (
                              <DropdownMenuItem
                                onClick={() => onPermissions(usr)}
                                className="flex items-center gap-2 hover:bg-muted hover:text-amber-500 cursor-pointer focus:bg-muted focus:text-amber-500 p-2 text-xs font-medium"
                              >
                                <Shield className="w-4.5 h-4.5" />
                                Permissões
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuItem
                              onClick={() => onEdit(usr)}
                              className="flex items-center gap-2 hover:bg-muted hover:text-teal-600 dark:hover:text-teal-400 cursor-pointer focus:bg-muted focus:text-teal-600 dark:focus:text-teal-400 p-2 text-xs font-medium"
                            >
                              <Edit2 className="w-4.5 h-4.5" />
                              Editar
                            </DropdownMenuItem>
                            {usr.id !== currentUser.id && (
                              <DropdownMenuItem
                                onClick={() => onDelete(usr)}
                                className="flex items-center gap-2 hover:bg-muted hover:text-red-500 cursor-pointer focus:bg-muted focus:text-red-500 p-2 text-xs font-medium"
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

      {/* Pagination bar */}
      <PaginationBar
        currentPage={pagination.page}
        pageSize={pagination.pageSize}
        total={pagination.total}
        itemLabel="usuários"
        isPending={isPending}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
    </div>
  );
};
