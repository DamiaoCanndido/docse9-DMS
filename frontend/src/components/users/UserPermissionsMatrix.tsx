'use client';

import React, { useState, useEffect } from 'react';
import { User, DocumentType, PermissionLevel } from '@/types';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Shield } from 'lucide-react';
import { getUserPermissions, updateUserPermissions } from '@/app/api/users';
import { toast } from 'sonner';

interface UserPermissionsMatrixProps {
  isOpen: boolean;
  onClose: () => void;
  user: User | null;
  path: string;
}

interface UserPermissionsContentProps {
  onClose: () => void;
  user: User;
  path: string;
}

const docTypesList: { value: DocumentType; label: string }[] = [
  { value: 'NOTICE', label: 'Ofício' },
  { value: 'DECREE', label: 'Decreto' },
  { value: 'ORDINANCE', label: 'Portaria' },
  { value: 'LAW', label: 'Lei' },
  { value: 'CONTRACT', label: 'Contrato' },
];

const permLevels: { value: PermissionLevel; label: string }[] = [
  { value: 'NONE', label: 'Nenhum' },
  { value: 'READ', label: 'Visualizar' },
  { value: 'WRITE', label: 'Criar / Editar' },
  { value: 'DELETE', label: 'Excluir' },
];

const defaultPermissions: Record<DocumentType, PermissionLevel> = {
  NOTICE: 'NONE',
  DECREE: 'NONE',
  ORDINANCE: 'NONE',
  LAW: 'NONE',
  CONTRACT: 'NONE',
};

const UserPermissionsContent: React.FC<UserPermissionsContentProps> = ({
  onClose,
  user,
  path,
}) => {
  const [permissions, setPermissions] = useState<Record<DocumentType, PermissionLevel>>(defaultPermissions);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    let isMounted = true;

    getUserPermissions(user.id)
      .then((userPerms) => {
        if (!isMounted) return;
        const permsMap = { ...defaultPermissions };
        userPerms.forEach((p) => {
          permsMap[p.documentType] = p.level;
        });
        setPermissions(permsMap);
      })
      .catch((err) => {
        if (!isMounted) return;
        console.error('Erro ao buscar permissões:', err);
        toast.error('Erro ao buscar permissões do usuário.');
        onClose();
      })
      .finally(() => {
        if (isMounted) setIsLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [user.id, onClose]);

  const handlePermissionChange = (type: DocumentType, level: PermissionLevel) => {
    setPermissions((prev) => ({ ...prev, [type]: level }));
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const permsArray = (Object.keys(permissions) as DocumentType[]).map((type) => ({
        documentType: type,
        level: permissions[type],
      }));

      await updateUserPermissions({
        id: user.id,
        permissions: permsArray,
        path,
      });

      toast.success('Permissões atualizadas com sucesso!');
      onClose();
    } catch {
      toast.error('Erro ao atualizar permissões.');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <>
      <DialogHeader>
        <DialogTitle className="text-xl font-bold text-foreground flex items-center gap-2">
          <Shield className="w-5 h-5 text-amber-500" />
          Permissões do Usuário: {user.username}
        </DialogTitle>
        <DialogDescription className="text-muted-foreground text-xs mt-1">
          Configure o nível de permissão que o funcionário comum terá para cada tipo de documento oficial.
        </DialogDescription>
      </DialogHeader>

      {isLoading ? (
        <div className="py-12 flex flex-col items-center justify-center gap-3">
          <div className="w-8 h-8 rounded-full border-2 border-amber-500/20 border-t-amber-500 animate-spin" />
          <span className="text-muted-foreground text-xs">Carregando permissões...</span>
        </div>
      ) : (
        <div className="flex flex-col gap-6 mt-4">
          <div className="border border-border rounded-xl overflow-hidden divide-y divide-border bg-card">
            <div className="grid grid-cols-2 md:grid-cols-5 p-3.5 bg-muted/50 text-xs font-bold text-muted-foreground uppercase tracking-wider">
              <div className="col-span-2">Tipo de Documento</div>
              <div className="col-span-3">Nível de Acesso</div>
            </div>

            {docTypesList.map((type) => (
              <div
                key={type.value}
                className="grid grid-cols-2 md:grid-cols-5 p-3.5 items-center gap-4 hover:bg-muted/30 transition-colors"
              >
                <div className="col-span-2 font-semibold text-foreground">
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
                          ? 'bg-amber-500/10 border border-amber-500/35 text-amber-600 dark:text-amber-400 font-bold'
                          : 'bg-muted border border-border text-muted-foreground hover:text-foreground'
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
                className="border-border text-foreground hover:bg-muted rounded-xl"
                onClick={onClose}
                disabled={isSaving}
              >
                Cancelar
              </Button>
            } />
            <Button
              type="button"
              variant="default"
              className="rounded-xl font-bold bg-amber-500/20 border border-amber-500/35 text-amber-600 dark:text-amber-400 hover:bg-amber-500/30 hover:text-amber-700 dark:hover:text-amber-300 shadow-sm px-6"
              onClick={handleSave}
              isLoading={isSaving}
            >
              Salvar Permissões
            </Button>
          </DialogFooter>
        </div>
      )}
    </>
  );
};

export const UserPermissionsMatrix: React.FC<UserPermissionsMatrixProps> = ({
  isOpen,
  onClose,
  user,
  path,
}) => {
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="bg-card border border-border text-foreground max-w-2xl rounded-2xl shadow-2xl p-6">
        {isOpen && user && (
          <UserPermissionsContent
            key={user.id}
            user={user}
            onClose={onClose}
            path={path}
          />
        )}
      </DialogContent>
    </Dialog>
  );
};
