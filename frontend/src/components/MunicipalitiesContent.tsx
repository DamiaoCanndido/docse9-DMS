'use client';

import React, { useState, useTransition } from 'react';
import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { Municipality, PaginatedResponse, User } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  createMunicipality,
  updateMunicipality,
  deleteMunicipality,
  restoreMunicipality,
  hardDeleteMunicipality,
} from '@/app/api/municipalities';
import { toast } from 'sonner';
import {
  Building2,
  Plus,
  Trash2,
  Edit2,
  RotateCcw,
  Trash,
  ChevronLeft,
  ChevronRight,
  Globe,
  Calendar,
  Sparkles,
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

interface MunicipalitiesContentProps {
  initialData: PaginatedResponse<Municipality>;
  currentUser: User;
  viewTrash: boolean;
}

export const MunicipalitiesContent: React.FC<MunicipalitiesContentProps> = ({
  initialData,
  currentUser,
  viewTrash,
}) => {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingMunicipality, setEditingMunicipality] = useState<Municipality | null>(null);

  // Form states
  const [name, setName] = useState('');
  const [uf, setUf] = useState('');
  const [imageUrl, setImageUrl] = useState('');
  const [formError, setFormError] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  const currentPage = Number(searchParams.get('page')) || 1;

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

  const openCreateModal = () => {
    setEditingMunicipality(null);
    setName('');
    setUf('');
    setImageUrl('');
    setFormError('');
    setIsModalOpen(true);
  };

  const openEditModal = (mun: Municipality) => {
    setEditingMunicipality(mun);
    setName(mun.name);
    setUf(mun.uf);
    setImageUrl(mun.imageUrl || '');
    setFormError('');
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setEditingMunicipality(null);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !uf.trim()) {
      setFormError('Nome e UF são obrigatórios.');
      return;
    }
    if (uf.length !== 2) {
      setFormError('A UF deve conter exatamente 2 caracteres.');
      return;
    }

    setIsSaving(true);
    setFormError('');

    try {
      if (editingMunicipality) {
        await updateMunicipality({
          id: editingMunicipality.id,
          input: { name, uf: uf.toUpperCase(), imageUrl: imageUrl || undefined },
          path: pathname,
        });
        toast.success('Município atualizado com sucesso!');
      } else {
        await createMunicipality({
          input: { name, uf: uf.toUpperCase(), imageUrl: imageUrl || undefined },
          path: pathname,
        });
        toast.success('Município cadastrado com sucesso!');
      }
      closeModal();
    } catch (err: any) {
      setFormError(err.response?.data?.error || 'Erro ao salvar município.');
      toast.error('Ocorreu um erro.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    const confirm = window.confirm('Deseja realmente enviar este município para a lixeira?');
    if (!confirm) return;

    try {
      await deleteMunicipality({ id, path: pathname });
      toast.success('Município enviado para a lixeira.');
    } catch {
      toast.error('Erro ao excluir município.');
    }
  };

  const handleRestore = async (id: string) => {
    try {
      await restoreMunicipality({ id, path: pathname });
      toast.success('Município restaurado com sucesso!');
    } catch {
      toast.error('Erro ao restaurar município.');
    }
  };

  const handleHardDelete = async (id: string) => {
    const confirm = window.confirm(
      'ATENÇÃO: Isso excluirá permanentemente o município e todos os seus dados vinculados. Esta ação NÃO pode ser desfeita. Confirmar?'
    );
    if (!confirm) return;

    try {
      await hardDeleteMunicipality({ id, path: pathname });
      toast.success('Município excluído definitivamente.');
    } catch {
      toast.error('Erro ao excluir permanentemente.');
    }
  };

  return (
    <div className="flex flex-col gap-8">
      {/* Header section */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <div className="flex items-center gap-2 text-violet-400 text-xs font-semibold uppercase tracking-wider mb-1.5">
            <Sparkles className="w-4 h-4" />
            Painel de Administração Global
          </div>
          <h1 className="text-3xl font-extrabold tracking-tight text-white flex items-center gap-3">
            <Building2 className="w-8 h-8 text-violet-500" />
            {viewTrash ? 'Municípios Excluídos (Lixeira)' : 'Gestão de Municípios'}
          </h1>
          <p className="text-zinc-400 mt-1.5 text-sm max-w-xl">
            {viewTrash
              ? 'Gerencie os municípios deletados. Restaure-os ou exclua-os permanentemente do banco de dados.'
              : 'Cadastre, edite e configure as organizações municipais atendidas pelo Docseq DMS.'}
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
              Novo Município
            </Button>
          )}
        </div>
      </div>

      {/* Grid or Table container */}
      <div className="bg-zinc-950/40 border border-zinc-900 rounded-2xl overflow-hidden shadow-2xl relative">
        {isPending && (
          <div className="absolute inset-0 bg-black/60 backdrop-blur-[2px] z-10 flex items-center justify-center transition-all duration-300">
            <div className="flex flex-col items-center gap-3">
              <div className="w-10 h-10 rounded-full border-2 border-violet-500/20 border-t-violet-500 animate-spin" />
              <span className="text-violet-400 text-xs font-bold tracking-widest uppercase">
                Carregando dados...
              </span>
            </div>
          </div>
        )}

        {initialData.data.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24 px-4 text-center">
            <div className="w-14 h-14 rounded-2xl bg-zinc-900/50 flex items-center justify-center text-zinc-500 mb-5 border border-zinc-800">
              <Building2 className="w-7 h-7 text-zinc-600" />
            </div>
            <h3 className="text-xl font-bold text-white">Nenhum município encontrado</h3>
            <p className="text-zinc-500 text-sm mt-2 max-w-sm">
              Não há municípios registrados {viewTrash ? 'na lixeira' : 'no sistema'}.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-zinc-900 bg-zinc-950/70 text-zinc-400 text-[11px] font-bold uppercase tracking-wider">
                  <th className="px-6 py-4.5">Logo</th>
                  <th className="px-6 py-4.5">Nome do Município</th>
                  <th className="px-6 py-4.5">Estado (UF)</th>
                  <th className="px-6 py-4.5">Criado Em</th>
                  <th className="px-6 py-4.5 text-center">Ações</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-900 text-sm text-zinc-300">
                {initialData.data.map((mun) => (
                  <tr key={mun.id} className="hover:bg-zinc-900/10 transition-colors group">
                    <td className="px-6 py-4.5">
                      {mun.imageUrl ? (
                        <img
                          src={mun.imageUrl}
                          alt={mun.name}
                          className="w-9 h-9 rounded-lg bg-zinc-900 border border-zinc-800 object-cover"
                        />
                      ) : (
                        <div className="w-9 h-9 rounded-lg bg-zinc-900 border border-zinc-800 flex items-center justify-center text-zinc-600">
                          <Building2 className="w-5 h-5" />
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4.5 font-bold text-white group-hover:text-violet-400 transition-colors">
                      {mun.name}
                    </td>
                    <td className="px-6 py-4.5">
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold bg-zinc-900 border border-zinc-850 text-zinc-300">
                        <Globe className="w-3.5 h-3.5 text-zinc-500" />
                        {mun.uf}
                      </span>
                    </td>
                    <td className="px-6 py-4.5 text-zinc-500 text-xs font-medium">
                      <div className="flex items-center gap-1.5">
                        <Calendar className="w-3.5 h-3.5 text-zinc-600" />
                        {new Date(mun.createdAt).toLocaleDateString('pt-BR')}
                      </div>
                    </td>
                    <td className="px-6 py-4.5 text-center">
                      <DropdownMenu>
                        <DropdownMenuTrigger render={
                          <button className="p-1.5 text-zinc-400 hover:text-white hover:bg-zinc-900 rounded-lg transition-all cursor-pointer">
                            <MoreHorizontal className="w-4.5 h-4.5" />
                          </button>
                        } />
                        <DropdownMenuContent align="end" className="bg-zinc-950 border-zinc-800 text-zinc-300 min-w-[160px]">
                          {viewTrash ? (
                            <>
                              <DropdownMenuItem
                                onClick={() => handleRestore(mun.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-emerald-400 cursor-pointer focus:bg-zinc-900 focus:text-emerald-400 p-2 text-xs font-medium"
                              >
                                <RotateCcw className="w-4 h-4" />
                                Restaurar
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleHardDelete(mun.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-red-400 cursor-pointer focus:bg-zinc-900 focus:text-red-400 p-2 text-xs font-medium"
                              >
                                <Trash className="w-4 h-4" />
                                Excluir Permanentemente
                              </DropdownMenuItem>
                            </>
                          ) : (
                            <>
                              <DropdownMenuItem
                                onClick={() => openEditModal(mun)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-violet-400 cursor-pointer focus:bg-zinc-900 focus:text-violet-400 p-2 text-xs font-medium"
                              >
                                <Edit2 className="w-4 h-4" />
                                Editar
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleDelete(mun.id)}
                                className="flex items-center gap-2 hover:bg-zinc-900 hover:text-red-400 cursor-pointer focus:bg-zinc-900 focus:text-red-400 p-2 text-xs font-medium"
                              >
                                <Trash2 className="w-4 h-4" />
                                Mover para Lixeira
                              </DropdownMenuItem>
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
              Mostrando {initialData.data.length} de {initialData.total} municípios
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

      {/* Create/Edit Modal via shadcn/ui Dialog */}
      <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
        <DialogContent className="bg-zinc-950 border-zinc-800/80 text-zinc-300 max-w-lg rounded-2xl shadow-2xl p-6">
          <DialogHeader>
            <DialogTitle className="text-xl font-bold text-white">
              {editingMunicipality ? 'Editar Município' : 'Cadastrar Novo Município'}
            </DialogTitle>
            <DialogDescription className="text-zinc-500 text-xs mt-1">
              {editingMunicipality
                ? 'Atualize as informações do município selecionado.'
                : 'Insira os dados da prefeitura para adicioná-la ao sistema.'}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleSave} className="flex flex-col gap-5 mt-4">
            <Input
              label="Nome do Município"
              placeholder="Ex: Passagem"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />

            <Input
              label="Estado (UF)"
              placeholder="Ex: PB"
              maxLength={2}
              value={uf}
              onChange={(e) => setUf(e.target.value)}
              className="uppercase"
              required
            />

            <Input
              label="URL da Imagem do Brasão (opcional)"
              placeholder="Ex: https://..."
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
            />

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
                  onClick={closeModal}
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
