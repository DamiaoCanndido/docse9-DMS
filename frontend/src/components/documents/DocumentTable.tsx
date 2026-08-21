'use client';

import React from 'react';
import { Document, DocumentType, ContractType, User } from '@/types';
import { PaginationBar } from '@/components/ui/PaginationBar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  FolderOpen,
  Calendar as CalendarIcon,
  FileCheck,
  MoreHorizontal,
  Edit2,
  Trash2,
  RotateCcw,
  Trash,
} from 'lucide-react';
import { formatDateTime } from '@/lib/date';

interface DocumentTableProps {
  documents: Document[];
  activeTab: DocumentType;
  activeTabLabel: string;
  viewTrash: boolean;
  currentUser: User;
  isPending: boolean;
  canEdit: (type: DocumentType) => boolean;
  canDelete: (type: DocumentType) => boolean;
  onEdit: (doc: Document) => void;
  onDelete: (doc: Document) => void;
  onRestore: (doc: Document) => void;
  onHardDelete: (doc: Document) => void;
  pagination: {
    page: number;
    pageSize: number;
    total: number;
  };
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  searchVal?: string;
}

const contractTypeLabels: Record<ContractType, string> = {
  service: 'Prestação de Serviço',
  bidding: 'Licitação',
  publicinterest: 'Interesse Público',
};

export const DocumentTable: React.FC<DocumentTableProps> = ({
  documents,
  activeTab,
  activeTabLabel,
  viewTrash,
  currentUser,
  isPending,
  canEdit,
  canDelete,
  onEdit,
  onDelete,
  onRestore,
  onHardDelete,
  pagination,
  onPageChange,
  onPageSizeChange,
  searchVal,
}) => {
  return (
    <div className="bg-card border border-border rounded-2xl overflow-hidden shadow-xl relative">
      {/* Loading overlay */}
      {isPending && (
        <div className="absolute inset-0 bg-background/60 backdrop-blur-[2px] z-10 flex items-center justify-center transition-all duration-300">
          <div className="flex flex-col items-center gap-3">
            <div className="w-10 h-10 rounded-full border-2 border-teal-500/20 border-t-teal-500 animate-spin" />
            <span className="text-teal-600 dark:text-teal-400 text-xs font-bold tracking-widest uppercase">
              Carregando {activeTabLabel.toLowerCase()}...
            </span>
          </div>
        </div>
      )}

      {documents.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 px-4 text-center">
          <div className="w-14 h-14 rounded-2xl bg-muted/60 flex items-center justify-center text-muted-foreground mb-4 border border-border">
            <FolderOpen className="w-7 h-7 text-muted-foreground" />
          </div>
          <h3 className="text-xl font-bold text-foreground">Nenhum registro encontrado</h3>
          <p className="text-muted-foreground text-sm mt-1.5 max-w-sm">
            Não há {activeTabLabel.toLowerCase()} cadastrados{searchVal ? ' com os termos buscados' : ''}.
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-border bg-muted/50 text-muted-foreground text-[11px] font-bold uppercase tracking-wider">
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
            <tbody className="divide-y divide-border text-sm text-foreground">
              {documents.map((doc) => {
                const showEdit = !viewTrash && canEdit(doc.type);
                const showDelete = !viewTrash && canDelete(doc.type);
                const showRestore = viewTrash && currentUser.role === 'MOD';
                const showHardDelete = viewTrash && currentUser.role === 'MOD';

                return (
                  <tr key={doc.id} className="hover:bg-muted/40 transition-colors group">
                    <td className="px-6 py-4.5 font-bold text-foreground group-hover:text-teal-600 dark:group-hover:text-teal-400 transition-colors">
                      #{doc.order}
                    </td>
                    <td className="px-6 py-4.5 max-w-xs md:max-w-md font-medium text-foreground">
                      {doc.description}
                    </td>
                    {activeTab === 'CONTRACT' ? (
                      <>
                        <td className="px-6 py-4.5 text-xs">
                          <span className="font-semibold text-foreground flex items-center gap-1.5">
                            <FileCheck className="w-3.5 h-3.5 text-teal-600 dark:text-teal-400" />
                            {doc.contractType
                              ? contractTypeLabels[doc.contractType] || doc.contractType
                              : '—'}
                          </span>
                        </td>
                        <td className="px-6 py-4.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                          {doc.value
                            ? doc.value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
                            : '—'}
                        </td>
                        <td className="px-6 py-4.5 text-xs text-muted-foreground font-medium">
                          {doc.duration ? `${doc.duration} meses` : '—'}
                        </td>
                        <td className="px-6 py-4.5 text-right text-muted-foreground text-xs font-medium">
                          <div className="flex items-center justify-end gap-1.5">
                            <CalendarIcon className="w-3.5 h-3.5 text-muted-foreground" />
                            {formatDateTime(doc.startIn)}
                          </div>
                        </td>
                      </>
                    ) : (
                      <>
                        <td className="px-6 py-4.5 text-xs text-muted-foreground">
                          {doc.createdBy?.username || doc.createdBy?.email || 'Sistema'}
                        </td>
                        <td className="px-6 py-4.5 text-right text-muted-foreground text-xs font-medium">
                          <div className="flex items-center justify-end gap-1.5">
                            <CalendarIcon className="w-3.5 h-3.5 text-muted-foreground" />
                            {formatDateTime(doc.createdAt)}
                          </div>
                        </td>
                      </>
                    )}
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
                          {showEdit && (
                            <DropdownMenuItem
                              onClick={() => onEdit(doc)}
                              className="flex items-center gap-2 hover:bg-muted hover:text-teal-600 dark:hover:text-teal-400 cursor-pointer focus:bg-muted focus:text-teal-600 dark:focus:text-teal-400 p-2 text-xs font-medium"
                            >
                              <Edit2 className="w-4 h-4" />
                              Editar
                            </DropdownMenuItem>
                          )}
                          {showDelete && (
                            <DropdownMenuItem
                              onClick={() => onDelete(doc)}
                              className="flex items-center gap-2 hover:bg-muted hover:text-red-500 cursor-pointer focus:bg-muted focus:text-red-500 p-2 text-xs font-medium"
                            >
                              <Trash2 className="w-4 h-4" />
                              Mover para Lixeira
                            </DropdownMenuItem>
                          )}
                          {showRestore && (
                            <DropdownMenuItem
                              onClick={() => onRestore(doc)}
                              className="flex items-center gap-2 hover:bg-muted hover:text-emerald-500 cursor-pointer focus:bg-muted focus:text-emerald-500 p-2 text-xs font-medium"
                            >
                              <RotateCcw className="w-4.5 h-4.5" />
                              Restaurar
                            </DropdownMenuItem>
                          )}
                          {showHardDelete && (
                            <DropdownMenuItem
                              onClick={() => onHardDelete(doc)}
                              className="flex items-center gap-2 hover:bg-muted hover:text-red-500 cursor-pointer focus:bg-muted focus:text-red-500 p-2 text-xs font-medium"
                            >
                              <Trash className="w-4.5 h-4.5" />
                              Excluir Definitivamente
                            </DropdownMenuItem>
                          )}
                          {!showEdit && !showDelete && !showRestore && !showHardDelete && (
                            <span className="p-2 text-xs text-muted-foreground italic block text-center">
                              Sem permissões
                            </span>
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
        currentPage={pagination.page}
        pageSize={pagination.pageSize}
        total={pagination.total}
        itemLabel="documentos"
        isPending={isPending}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
    </div>
  );
};
