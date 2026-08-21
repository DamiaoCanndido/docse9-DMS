'use client';

import React from 'react';
import { User } from '@/types';
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
import { Trash2, RotateCcw, AlertTriangle, Trash } from 'lucide-react';

export type UserDeleteMode = 'delete' | 'restore' | 'hardDelete';

interface UserDeleteDialogProps {
  isOpen: boolean;
  onClose: () => void;
  mode: UserDeleteMode;
  user: User | null;
  onConfirm: () => Promise<void> | void;
  isLoading?: boolean;
}

export const UserDeleteDialog: React.FC<UserDeleteDialogProps> = ({
  isOpen,
  onClose,
  mode,
  user,
  onConfirm,
  isLoading = false,
}) => {
  if (!user) return null;

  const isHard = mode === 'hardDelete';
  const isRestore = mode === 'restore';

  const title = isRestore
    ? 'Restaurar Usuário'
    : isHard
    ? 'Excluir Definitivamente'
    : 'Mover para a Lixeira';

  const icon = isRestore ? (
    <RotateCcw className="w-5 h-5 text-emerald-500" />
  ) : isHard ? (
    <AlertTriangle className="w-5 h-5 text-red-500 animate-pulse" />
  ) : (
    <Trash2 className="w-5 h-5 text-red-500" />
  );

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="bg-card border border-border text-foreground max-w-md rounded-2xl shadow-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-lg font-bold text-foreground flex items-center gap-2">
            {icon}
            {title}
          </DialogTitle>
          <DialogDescription className="text-muted-foreground text-xs mt-2 leading-relaxed">
            {isRestore ? (
              <span>
                Deseja restaurar o acesso do usuário <strong>{user.username}</strong> ({user.email})?
              </span>
            ) : isHard ? (
              <span>
                ATENÇÃO: Esta ação é definitiva e removerá permanentemente o usuário <strong>{user.username}</strong> ({user.email}). Esta ação não pode ser desfeita.
              </span>
            ) : (
              <span>
                Deseja mover o usuário <strong>{user.username}</strong> ({user.email}) para a lixeira?
              </span>
            )}
          </DialogDescription>
        </DialogHeader>

        {isHard && (
          <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl flex items-start gap-2.5 text-xs text-red-600 dark:text-red-400">
            <Trash className="w-4 h-4 shrink-0 mt-0.5" />
            <span>Todos os dados e vínculos deste usuário serão excluídos permanentemente.</span>
          </div>
        )}

        <DialogFooter className="mt-4 gap-2 flex flex-row justify-end">
          <DialogClose render={
            <Button
              type="button"
              variant="outline"
              className="border-border text-foreground hover:bg-muted rounded-xl text-xs"
              onClick={onClose}
              disabled={isLoading}
            >
              Cancelar
            </Button>
          } />
          <Button
            type="button"
            variant={isRestore ? 'default' : 'destructive'}
            className={`rounded-xl font-bold px-5 text-xs ${
              isRestore ? 'bg-emerald-600 hover:bg-emerald-700 text-white' : ''
            }`}
            onClick={onConfirm}
            isLoading={isLoading}
          >
            {isRestore ? 'Restaurar' : isHard ? 'Excluir Definitivamente' : 'Mover para Lixeira'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
