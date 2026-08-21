'use client';

import { useState, useCallback } from 'react';
import { usePathname } from 'next/navigation';
import { Document, CreateDocumentInput, UpdateDocumentInput } from '@/types';
import {
  createDocument,
  updateDocument,
  deleteDocument,
  restoreDocument,
  hardDeleteDocument,
} from '@/app/api/documents';
import { toast } from 'sonner';
import { isRedirectError } from '@/lib/utils';

export type DocumentDeleteMode = 'delete' | 'restore' | 'hardDelete';

export interface DeleteDialogState {
  isOpen: boolean;
  mode: DocumentDeleteMode;
  document: Document | null;
}

export function useDocumentActions() {
  const pathname = usePathname();
  const [isMutating, setIsMutating] = useState(false);
  const [deleteDialogState, setDeleteDialogState] = useState<DeleteDialogState>({
    isOpen: false,
    mode: 'delete',
    document: null,
  });

  const openDeleteDialog = useCallback((doc: Document, mode: DocumentDeleteMode) => {
    setDeleteDialogState({
      isOpen: true,
      mode,
      document: doc,
    });
  }, []);

  const closeDeleteDialog = useCallback(() => {
    setDeleteDialogState((prev) => ({
      ...prev,
      isOpen: false,
      document: null,
    }));
  }, []);

  const handleCreate = useCallback(
    async (input: CreateDocumentInput): Promise<Document | null> => {
      setIsMutating(true);
      try {
        const result = await createDocument({ input, path: pathname });
        toast.success('Documento criado com sucesso!');
        return result;
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const errorMsg =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
          'Erro ao criar o documento.';
        toast.error(errorMsg);
        throw err;
      } finally {
        setIsMutating(false);
      }
    },
    [pathname]
  );

  const handleUpdate = useCallback(
    async (id: string, input: UpdateDocumentInput): Promise<Document | null> => {
      setIsMutating(true);
      try {
        const result = await updateDocument({ id, input, path: pathname });
        toast.success('Documento atualizado com sucesso!');
        return result;
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const errorMsg =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
          'Erro ao salvar o documento.';
        toast.error(errorMsg);
        throw err;
      } finally {
        setIsMutating(false);
      }
    },
    [pathname]
  );

  const handleDelete = useCallback(
    async (id: string): Promise<void> => {
      setIsMutating(true);
      try {
        await deleteDocument({ id, path: pathname });
        toast.success('Documento enviado para a lixeira com sucesso!');
        closeDeleteDialog();
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        toast.error('Erro ao excluir documento.');
      } finally {
        setIsMutating(false);
      }
    },
    [pathname, closeDeleteDialog]
  );

  const handleRestore = useCallback(
    async (id: string): Promise<void> => {
      setIsMutating(true);
      try {
        await restoreDocument({ id, path: pathname });
        toast.success('Documento restaurado com sucesso!');
        closeDeleteDialog();
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        toast.error('Erro ao restaurar documento.');
      } finally {
        setIsMutating(false);
      }
    },
    [pathname, closeDeleteDialog]
  );

  const handleHardDelete = useCallback(
    async (id: string): Promise<void> => {
      setIsMutating(true);
      try {
        await hardDeleteDocument({ id, path: pathname });
        toast.success('Documento deletado permanentemente.');
        closeDeleteDialog();
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        toast.error('Erro ao excluir definitivamente.');
      } finally {
        setIsMutating(false);
      }
    },
    [pathname, closeDeleteDialog]
  );

  const confirmDeleteDialogAction = useCallback(async () => {
    if (!deleteDialogState.document) return;
    const docId = deleteDialogState.document.id;

    switch (deleteDialogState.mode) {
      case 'delete':
        await handleDelete(docId);
        break;
      case 'restore':
        await handleRestore(docId);
        break;
      case 'hardDelete':
        await handleHardDelete(docId);
        break;
    }
  }, [deleteDialogState, handleDelete, handleRestore, handleHardDelete]);

  return {
    isMutating,
    deleteDialogState,
    openDeleteDialog,
    closeDeleteDialog,
    confirmDeleteDialogAction,
    handleCreate,
    handleUpdate,
    handleDelete,
    handleRestore,
    handleHardDelete,
  };
}
