'use client';

import React, { useState } from 'react';
import { Document, DocumentType, ContractType, CreateDocumentInput, UpdateDocumentInput } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Calendar } from '@/components/ui/calendar';
import { motion } from 'framer-motion';
import { Calendar as CalendarIcon, Clock, FileCheck } from 'lucide-react';
import { formatDate, formatTime, parseDateSafe, combineDateAndTime, ptBR } from '@/lib/date';
import { cn } from '@/lib/utils';
import { z } from 'zod';

const documentFormSchema = z.object({
  type: z.enum(['NOTICE', 'DECREE', 'ORDINANCE', 'LAW', 'CONTRACT']),
  description: z.string().trim().min(1, 'A descrição do documento é obrigatória.'),
  contractType: z.enum(['service', 'bidding', 'publicinterest']).optional(),
  value: z.string().optional(),
  duration: z.string().optional(),
});

export interface DocumentTypeOption {
  value: DocumentType;
  label: string;
  singleLabel: string;
  icon?: React.ComponentType<{ className?: string }>;
}

interface DocumentFormDialogProps {
  isOpen: boolean;
  onClose: () => void;
  editingDocument: Document | null;
  activeTab: DocumentType;
  canCreate: (type: DocumentType) => boolean;
  docTypesList: DocumentTypeOption[];
  contractTypeLabels: Record<ContractType, string>;
  onSave: (payload: {
    isEdit: boolean;
    id?: string;
    createInput?: CreateDocumentInput;
    updateInput?: UpdateDocumentInput;
  }) => Promise<void>;
  creatorId: string;
  municipalityId: string;
}

interface DocumentFormContentProps {
  onClose: () => void;
  editingDocument: Document | null;
  activeTab: DocumentType;
  canCreate: (type: DocumentType) => boolean;
  docTypesList: DocumentTypeOption[];
  contractTypeLabels: Record<ContractType, string>;
  onSave: (payload: {
    isEdit: boolean;
    id?: string;
    createInput?: CreateDocumentInput;
    updateInput?: UpdateDocumentInput;
  }) => Promise<void>;
  creatorId: string;
  municipalityId: string;
}

const DocumentFormContent: React.FC<DocumentFormContentProps> = ({
  onClose,
  editingDocument,
  activeTab,
  canCreate,
  docTypesList,
  contractTypeLabels,
  onSave,
  creatorId,
  municipalityId,
}) => {
  const initialType: DocumentType = editingDocument
    ? editingDocument.type
    : canCreate(activeTab)
    ? activeTab
    : docTypesList.find((t) => canCreate(t.value))?.value || 'NOTICE';

  const [type, setType] = useState<DocumentType>(initialType);
  const [description, setDescription] = useState(editingDocument?.description || '');
  const [contractType, setContractType] = useState<ContractType>(
    editingDocument?.contractType || 'service'
  );
  const [value, setValue] = useState(
    editingDocument?.value ? editingDocument.value.toString() : ''
  );
  const [duration, setDuration] = useState(
    editingDocument?.duration ? editingDocument.duration.toString() : ''
  );

  const initialStartInDate = editingDocument?.startIn
    ? parseDateSafe(editingDocument.startIn)
    : undefined;
  const initialStartInTime = initialStartInDate ? formatTime(initialStartInDate) : '00:00';

  const [startInDate, setStartInDate] = useState<Date | undefined>(initialStartInDate);
  const [startInTime, setStartInTime] = useState(initialStartInTime);

  const initialCreatedAtDate = editingDocument?.createdAt
    ? parseDateSafe(editingDocument.createdAt)
    : undefined;
  const initialCreatedAtTime = initialCreatedAtDate ? formatTime(initialCreatedAtDate) : '00:00';

  const [createdAtDate, setCreatedAtDate] = useState<Date | undefined>(initialCreatedAtDate);
  const [createdAtTime, setCreatedAtTime] = useState(initialCreatedAtTime);

  const [formError, setFormError] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    // Zod validation
    const parsed = documentFormSchema.safeParse({
      type,
      description,
      contractType,
      value,
      duration,
    });

    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message || 'Preencha os campos obrigatórios.');
      return;
    }

    if (type === 'CONTRACT') {
      if (!value.trim() || isNaN(Number(value)) || Number(value) <= 0) {
        setFormError('O valor do contrato é obrigatório e deve ser maior que zero.');
        return;
      }
      if (!duration.trim() || isNaN(Number(duration)) || Number(duration) <= 0) {
        setFormError('A duração em meses do contrato é obrigatória e deve ser maior que zero.');
        return;
      }
      if (!startInDate) {
        setFormError('A data de início do contrato é obrigatória.');
        return;
      }
    }

    setIsSaving(true);
    try {
      if (editingDocument) {
        const updateInput: UpdateDocumentInput = {
          description,
        };
        if (type === 'CONTRACT') {
          updateInput.contractType = contractType;
          updateInput.value = Number(value);
          updateInput.duration = Number(duration);
          updateInput.startIn = combineDateAndTime(startInDate, startInTime);
        } else if (createdAtDate) {
          updateInput.createdAt = combineDateAndTime(createdAtDate, createdAtTime);
        }

        await onSave({ isEdit: true, id: editingDocument.id, updateInput });
      } else {
        const createInput: CreateDocumentInput = {
          type,
          description,
          creatorId,
          municipalityId,
        };
        if (type === 'CONTRACT') {
          createInput.contractType = contractType;
          createInput.value = Number(value);
          createInput.duration = Number(duration);
          createInput.startIn = combineDateAndTime(startInDate, startInTime);
        }

        await onSave({ isEdit: false, createInput });
      }
      onClose();
    } catch (err: unknown) {
      const errorMsg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Erro ao salvar o documento.';
      setFormError(errorMsg);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 mt-4">
      {/* Type selection - read-only on edit */}
      <div className="w-full flex flex-col gap-1.5">
        <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">
          Tipo de Documento
        </label>
        {editingDocument ? (
          <div className="w-full bg-muted border border-border px-3.5 py-2.5 rounded-xl text-sm text-foreground font-semibold">
            {docTypesList.find((t) => t.value === type)?.singleLabel || type}
          </div>
        ) : (
          <Select
            value={type}
            onValueChange={(val) => setType(val as DocumentType)}
            items={{
              NOTICE: 'Ofício',
              DECREE: 'Decreto',
              ORDINANCE: 'Portaria',
              LAW: 'Lei',
              CONTRACT: 'Contrato',
            }}
          >
            <SelectTrigger className="w-full bg-background border-border text-foreground text-sm h-10 rounded-xl">
              <SelectValue placeholder="Selecione o tipo" />
            </SelectTrigger>
            <SelectContent className="bg-popover border-border text-popover-foreground">
              {docTypesList.map((t) => {
                const allowed = canCreate(t.value);
                return (
                  <SelectItem key={t.value} value={t.value} disabled={!allowed}>
                    {t.singleLabel} {!allowed ? '(Sem permissão)' : ''}
                  </SelectItem>
                );
              })}
            </SelectContent>
          </Select>
        )}
      </div>

      {/* Description textarea */}
      <div className="w-full flex flex-col gap-1.5">
        <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">
          Descrição / Ementa
        </label>
        <textarea
          className="w-full h-28 bg-background border border-border text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all duration-200 placeholder:text-muted-foreground focus:outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500"
          placeholder="Descreva o conteúdo do documento ou sua ementa oficial..."
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          required
        />
      </div>

      {/* Specific fields for CONTRACT */}
      {type === 'CONTRACT' && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="border border-border rounded-xl p-4 bg-muted/40 flex flex-col gap-4"
        >
          <h3 className="text-xs font-bold text-teal-600 dark:text-teal-400 uppercase tracking-widest flex items-center gap-1.5 mb-1">
            <FileCheck className="w-4 h-4" />
            Detalhes do Contrato
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-end">
            {/* Contract Type */}
            <div className="w-full flex flex-col gap-1.5">
              <label className="text-xs font-semibold text-foreground/80 uppercase tracking-wider">
                Tipo de Contrato
              </label>
              <Select
                value={contractType}
                onValueChange={(val) => setContractType(val as ContractType)}
                items={contractTypeLabels}
              >
                <SelectTrigger className="w-full bg-background border-border text-foreground text-xs h-10 rounded-xl">
                  <SelectValue placeholder="Selecione o tipo de contrato" />
                </SelectTrigger>
                <SelectContent className="bg-popover border-border text-popover-foreground">
                  <SelectItem value="service">Prestação de Serviço</SelectItem>
                  <SelectItem value="bidding">Licitação</SelectItem>
                  <SelectItem value="publicinterest">Interesse Público</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Value */}
            <Input
              label="Valor do Contrato (R$)"
              type="number"
              step="0.01"
              placeholder="0.00"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              className="text-xs py-2 h-10"
              required
            />

            {/* Duration */}
            <Input
              label="Duração (em meses)"
              type="number"
              placeholder="Ex: 12"
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
              className="text-xs py-2 h-10"
              required
            />

            {/* Start Date via Popover + Calendar */}
            <div className="w-full flex flex-col gap-1.5">
              <label className="text-xs font-semibold text-foreground/80 uppercase tracking-wider">
                Data de Início
              </label>
              <Popover>
                <PopoverTrigger render={
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal bg-background border-border text-foreground text-xs h-10 rounded-xl",
                      !startInDate && "text-muted-foreground"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4 text-muted-foreground" />
                    {startInDate ? formatDate(startInDate) : <span>Selecione a data</span>}
                  </Button>
                } />
                <PopoverContent className="w-auto p-0 bg-popover border-border text-popover-foreground" align="start">
                  <Calendar
                    mode="single"
                    selected={startInDate}
                    onSelect={(date) => setStartInDate(date || undefined)}
                    locale={ptBR}
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* Start Time (HH:mm) */}
            <Input
              label="Hora de Início (HH:mm)"
              type="time"
              value={startInTime}
              onChange={(e) => setStartInTime(e.target.value)}
              className="text-xs py-2 h-10"
              required
            />
          </div>
        </motion.div>
      )}

      {/* CreatedAt Date & Time editing for non-contract documents */}
      {editingDocument && type !== 'CONTRACT' && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="border border-border rounded-xl p-4 bg-muted/40 flex flex-col gap-4"
        >
          <h3 className="text-xs font-bold text-teal-600 dark:text-teal-400 uppercase tracking-widest flex items-center gap-1.5 mb-1">
            <Clock className="w-4 h-4" />
            Data e Hora de Registro Oficial
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 items-end">
            {/* CreatedAt Date via Popover + Calendar */}
            <div className="w-full flex flex-col gap-1.5">
              <label className="text-xs font-semibold text-foreground/80 uppercase tracking-wider">
                Data de Registro
              </label>
              <Popover>
                <PopoverTrigger render={
                  <Button
                    variant="outline"
                    className={cn(
                      "w-full justify-start text-left font-normal bg-background border-border text-foreground text-xs h-10 rounded-xl",
                      !createdAtDate && "text-muted-foreground"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4 text-muted-foreground" />
                    {createdAtDate ? formatDate(createdAtDate) : <span>Selecione a data</span>}
                  </Button>
                } />
                <PopoverContent className="w-auto p-0 bg-popover border-border text-popover-foreground" align="start">
                  <Calendar
                    mode="single"
                    selected={createdAtDate}
                    onSelect={(date) => setCreatedAtDate(date || undefined)}
                    locale={ptBR}
                  />
                </PopoverContent>
              </Popover>
            </div>

            {/* CreatedAt Time Input */}
            <Input
              label="Hora do Registro (HH:mm)"
              type="time"
              value={createdAtTime}
              onChange={(e) => setCreatedAtTime(e.target.value)}
              className="text-xs py-2 h-10"
            />
          </div>
        </motion.div>
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
            className="border-border text-foreground hover:bg-muted rounded-xl"
            onClick={onClose}
            disabled={isSaving}
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
  );
};

export const DocumentFormDialog: React.FC<DocumentFormDialogProps> = ({
  isOpen,
  onClose,
  editingDocument,
  activeTab,
  canCreate,
  docTypesList,
  contractTypeLabels,
  onSave,
  creatorId,
  municipalityId,
}) => {
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="bg-card border border-border text-foreground max-w-xl rounded-2xl shadow-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold text-foreground">
            {editingDocument ? 'Editar Documento' : 'Novo Documento Oficial'}
          </DialogTitle>
          <DialogDescription className="text-muted-foreground text-xs mt-1">
            {editingDocument
              ? 'Modifique os metadados do documento selecionado.'
              : 'Cadastre um novo documento oficial no isolamento do seu município.'}
          </DialogDescription>
        </DialogHeader>

        {isOpen && (
          <DocumentFormContent
            key={editingDocument ? editingDocument.id : 'new'}
            onClose={onClose}
            editingDocument={editingDocument}
            activeTab={activeTab}
            canCreate={canCreate}
            docTypesList={docTypesList}
            contractTypeLabels={contractTypeLabels}
            onSave={onSave}
            creatorId={creatorId}
            municipalityId={municipalityId}
          />
        )}
      </DialogContent>
    </Dialog>
  );
};
