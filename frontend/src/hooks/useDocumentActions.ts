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
  getDocumentUploadURL,
  confirmDocumentUpload,
  getDocumentFileURL,
} from '@/app/api/documents';
import axios from 'axios';
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
        if (!result.success) {
          const errorMsg = result.error || 'Erro ao criar o documento.';
          const err = new Error(errorMsg);
          (err as unknown as { response: { status?: number; data: { error: string } } }).response = {
            status: result.status,
            data: { error: errorMsg },
          };
          throw err;
        }
        toast.success('Documento criado com sucesso!');
        return result.data || null;
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const errorMsg =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
          (err as Error)?.message ||
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
        if (!result.success) {
          const errorMsg = result.error || 'Erro ao salvar o documento.';
          const err = new Error(errorMsg);
          (err as unknown as { response: { status?: number; data: { error: string } } }).response = {
            status: result.status,
            data: { error: errorMsg },
          };
          throw err;
        }
        toast.success('Documento atualizado com sucesso!');
        return result.data || null;
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const errorMsg =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
          (err as Error)?.message ||
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
        const result = await deleteDocument({ id, path: pathname });
        if (!result.success) {
          throw new Error(result.error || 'Erro ao excluir documento.');
        }
        toast.success('Documento enviado para a lixeira com sucesso!');
        closeDeleteDialog();
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const msg = (err as Error)?.message || 'Erro ao excluir documento.';
        toast.error(msg);
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
        const result = await restoreDocument({ id, path: pathname });
        if (!result.success) {
          throw new Error(result.error || 'Erro ao restaurar documento.');
        }
        toast.success('Documento restaurado com sucesso!');
        closeDeleteDialog();
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const msg = (err as Error)?.message || 'Erro ao restaurar documento.';
        toast.error(msg);
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
        const result = await hardDeleteDocument({ id, path: pathname });
        if (!result.success) {
          throw new Error(result.error || 'Erro ao excluir definitivamente.');
        }
        toast.success('Documento deletado permanentemente.');
        closeDeleteDialog();
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const msg = (err as Error)?.message || 'Erro ao excluir definitivamente.';
        toast.error(msg);
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

  const handleUploadAttachment = useCallback(
    async (
      docId: string,
      file: File,
      onProgress?: (pct: number) => void
    ): Promise<Document> => {
      setIsMutating(true);
      try {
        // 1. Solicita Presigned URL para upload direto
        const { uploadUrl, fileKey } = await getDocumentUploadURL({
          id: docId,
          fileName: file.name,
          fileSize: file.size,
        });

        // 2. Upload direto via PUT no Cloudflare R2 (sem passar pela API)
        await axios.put(uploadUrl, file, {
          headers: {
            'Content-Type': 'application/pdf',
          },
          onUploadProgress: (progressEvent) => {
            if (progressEvent.total) {
              const pct = Math.round((progressEvent.loaded * 100) / progressEvent.total);
              onProgress?.(pct);
            }
          },
        });

        // 3. Confirma o upload e valida OCR no backend
        const updated = await confirmDocumentUpload({
          id: docId,
          fileKey,
          path: pathname,
        });

        toast.success('Arquivo PDF anexado e validado com sucesso!');
        return updated;
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        let errorMsg =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
          (err as Error)?.message ||
          'Erro ao realizar upload do anexo.';
        if (errorMsg === 'Network Error') {
          errorMsg = 'Falha de rede ou CORS ao enviar o arquivo para o armazenamento (R2).';
        }
        toast.error(errorMsg);
        throw err;
      } finally {
        setIsMutating(false);
      }
    },
    [pathname]
  );

  const handleGetFileURL = useCallback(
    async (docId: string, download = false): Promise<string> => {
      try {
        const res = await getDocumentFileURL({ id: docId, download });
        return res.url;
      } catch (err: unknown) {
        if (isRedirectError(err)) {
          throw err;
        }
        const errorMsg =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
          'Erro ao obter link do documento anexo.';
        toast.error(errorMsg);
        throw err;
      }
    },
    []
  );

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
    handleUploadAttachment,
    handleGetFileURL,
  };
}
