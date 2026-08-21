'use client';

import React from 'react';
import { Document } from '@/types';
import { DocumentDeleteMode } from '@/hooks/useDocumentActions';
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

interface DocumentDeleteDialogProps {
  isOpen: boolean;
  onClose: () => void;
  mode: DocumentDeleteMode;
  document: Document | null;
  onConfirm: () => Promise<void> | void;
  isLoading?: boolean;
}

export const DocumentDeleteDialog: React.FC<DocumentDeleteDialogProps> = ({
  isOpen,
  onClose,
  mode,
  document,
  onConfirm,
  isLoading = false,
}) => {
  if (!document) return null;

  const isHard = mode === 'hardDelete';
  const isRestore = mode === 'restore';

  const title = isRestore
    ? 'Restaurar Documento'
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
                Deseja restaurar o documento oficial <strong>#{document.order}</strong> para a listagem de ativos?
              </span>
            ) : isHard ? (
              <span>
                ATENÇÃO: Esta ação é definitiva e removerá permanentemente o documento <strong>#{document.order}</strong> e todos os registros associados. Esta ação não pode ser desfeita.
              </span>
            ) : (
              <span>
                Deseja mover o documento oficial <strong>#{document.order}</strong> para a lixeira?
              </span>
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="my-2 p-3.5 bg-muted/60 border border-border rounded-xl flex flex-col gap-1 text-xs">
          <span className="font-semibold text-foreground truncate">
            {document.description}
          </span>
        </div>

        {isHard && (
          <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl flex items-start gap-2.5 text-xs text-red-600 dark:text-red-400">
            <Trash className="w-4 h-4 shrink-0 mt-0.5" />
            <span>Os dados serão apagados do banco de dados de forma irreversível.</span>
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
