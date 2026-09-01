'use client';

import React, { useState } from 'react';
import { User, Municipality, Role } from '@/types';
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
import { Building2, Key } from 'lucide-react';
import { z } from 'zod';

const userFormSchema = z.object({
  username: z
    .string()
    .trim()
    .min(3, 'O nome de usuário deve ter pelo menos 3 caracteres.')
    .max(255, 'O nome de usuário não pode ultrapassar 255 caracteres.'),
  email: z
    .string()
    .trim()
    .email('Formato de e-mail inválido.')
    .max(255, 'O e-mail não pode ultrapassar 255 caracteres.'),
  password: z.string().optional(),
  role: z.enum(['ADMIN', 'MOD', 'COMMON']),
  municipalityId: z.string().optional(),
});

interface UserFormDialogProps {
  isOpen: boolean;
  onClose: () => void;
  selectedUser: User | null;
  currentUser: User;
  municipalities: Municipality[];
  onSave: (payload: {
    isEdit: boolean;
    id?: string;
    username: string;
    email: string;
    password?: string;
    role: Role;
    municipalityId: string;
  }) => Promise<{ randomPassword?: string } | void>;
}

interface UserFormContentProps {
  onClose: () => void;
  selectedUser: User | null;
  currentUser: User;
  municipalities: Municipality[];
  onSave: (payload: {
    isEdit: boolean;
    id?: string;
    username: string;
    email: string;
    password?: string;
    role: Role;
    municipalityId: string;
  }) => Promise<{ randomPassword?: string } | void>;
}

const UserFormContent: React.FC<UserFormContentProps> = ({
  onClose,
  selectedUser,
  currentUser,
  municipalities,
  onSave,
}) => {
  const isAdmin = currentUser.role === 'ADMIN';
  const isMod = currentUser.role === 'MOD';

  const defaultMunicipalityId = selectedUser
    ? selectedUser.municipalityId
    : isMod
    ? currentUser.municipalityId
    : municipalities[0]?.id || '';

  const [username, setUsername] = useState(selectedUser?.username || '');
  const [email, setEmail] = useState(selectedUser?.email || '');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<Role>(selectedUser?.role || 'COMMON');
  const [municipalityId, setMunicipalityId] = useState(defaultMunicipalityId);
  const [formError, setFormError] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');

    const targetMunicipalityId = isAdmin ? municipalityId : currentUser.municipalityId;

    const parsed = userFormSchema.safeParse({
      username,
      email,
      password: password || undefined,
      role,
      municipalityId: targetMunicipalityId,
    });

    if (!parsed.success) {
      setFormError(parsed.error.issues[0]?.message || 'Verifique os campos obrigatórios.');
      return;
    }

    if (!targetMunicipalityId && !isAdmin) {
      setFormError('Município vinculado é obrigatório.');
      return;
    }

    setIsSaving(true);
    try {
      await onSave({
        isEdit: !!selectedUser,
        id: selectedUser?.id,
        username,
        email,
        password: password ? password : undefined,
        role,
        municipalityId: targetMunicipalityId,
      });
      onClose();
    } catch (err: unknown) {
      const errorMsg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Erro ao salvar usuário.';
      setFormError(errorMsg);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 mt-4">
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

      {selectedUser ? (
        <Input
          label="Senha (deixe em branco para não alterar)"
          type="password"
          placeholder="Mínimo 6 caracteres"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      ) : (
        <div className="p-3.5 bg-teal-500/10 border border-teal-500/20 rounded-xl flex items-start gap-3 text-xs text-teal-600 dark:text-teal-300">
          <Key className="w-4 h-4 text-teal-600 dark:text-teal-400 shrink-0 mt-0.5" />
          <span>
            Uma <strong>senha temporária aleatória</strong> será gerada automaticamente pelo sistema após a criação. Você poderá visualizá-la e copiá-la na tela seguinte.
          </span>
        </div>
      )}

      {/* Perfil Selection */}
      <div className="w-full flex flex-col gap-1.5">
        <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">
          Perfil de Acesso
        </label>
        <Select value={role} onValueChange={(val) => setRole((val as Role) || 'COMMON')}>
          <SelectTrigger className="w-full bg-background border-border text-foreground text-sm h-10 rounded-xl">
            <SelectValue placeholder="Selecione o perfil" />
          </SelectTrigger>
          <SelectContent className="bg-popover border-border text-popover-foreground">
            {isAdmin && <SelectItem value="ADMIN" label="Administrador Global">Administrador Global</SelectItem>}
            <SelectItem value="MOD" label="Moderador Municipal">Moderador Municipal</SelectItem>
            <SelectItem value="COMMON" label="Funcionário Comum">Funcionário Comum</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Municipality Selection (ADMIN only) */}
      {isAdmin && (
        <div className="w-full flex flex-col gap-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">
            Município Vinculado
          </label>
          <Select value={municipalityId} onValueChange={(val) => setMunicipalityId(val || '')}>
            <SelectTrigger className="w-full bg-background border-border text-foreground text-sm h-10 rounded-xl">
              <SelectValue placeholder="Selecione o município">
                {municipalityId
                  ? (() => {
                      const m = municipalities.find((m) => m.id === municipalityId);
                      return m ? `${m.name} (${m.uf})` : municipalityId;
                    })()
                  : null}
              </SelectValue>
            </SelectTrigger>
            <SelectContent className="bg-popover border-border text-popover-foreground">
              {municipalities.map((m) => (
                <SelectItem key={m.id} value={m.id} label={`${m.name} (${m.uf})`}>
                  {m.name} ({m.uf})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      {isMod && (
        <div className="p-3 bg-muted border border-border rounded-xl flex items-center gap-2">
          <Building2 className="w-4.5 h-4.5 text-muted-foreground" />
          <span className="text-xs text-muted-foreground">
            Vinculado a: <strong className="text-foreground">{currentUser.municipality?.name} ({currentUser.municipality?.uf})</strong>
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

export const UserFormDialog: React.FC<UserFormDialogProps> = ({
  isOpen,
  onClose,
  selectedUser,
  currentUser,
  municipalities,
  onSave,
}) => {
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="bg-card border border-border text-foreground max-w-lg rounded-2xl shadow-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold text-foreground">
            {selectedUser ? 'Editar Usuário' : 'Cadastrar Novo Usuário'}
          </DialogTitle>
          <DialogDescription className="text-muted-foreground text-xs mt-1">
            Configure as credenciais e o escopo de acesso do usuário.
          </DialogDescription>
        </DialogHeader>

        {isOpen && (
          <UserFormContent
            key={selectedUser ? selectedUser.id : 'new'}
            onClose={onClose}
            selectedUser={selectedUser}
            currentUser={currentUser}
            municipalities={municipalities}
            onSave={onSave}
          />
        )}
      </DialogContent>
    </Dialog>
  );
};
