'use client';

import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Sparkles, Key, Eye, EyeOff, Check, Copy, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';

interface UserCreatedSuccessDialogProps {
  isOpen: boolean;
  onClose: () => void;
  credentials: {
    username: string;
    email: string;
    randomPassword?: string;
  } | null;
}

export const UserCreatedSuccessDialog: React.FC<UserCreatedSuccessDialogProps> = ({
  isOpen,
  onClose,
  credentials,
}) => {
  const [showPassword, setShowPassword] = useState(false);
  const [copiedPassword, setCopiedPassword] = useState(false);

  const handleCopyPassword = () => {
    if (!credentials?.randomPassword) return;
    navigator.clipboard.writeText(credentials.randomPassword);
    setCopiedPassword(true);
    toast.success('Senha copiada para a área de transferência!');
    setTimeout(() => setCopiedPassword(false), 3000);
  };

  if (!credentials) return null;

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="bg-card border border-border text-foreground max-w-lg rounded-2xl shadow-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold text-foreground flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-emerald-500" />
            Usuário Criado com Sucesso!
          </DialogTitle>
          <DialogDescription className="text-muted-foreground text-xs mt-1">
            Copie a senha temporária abaixo para enviar ao usuário.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4 mt-3">
          <div className="p-3.5 bg-muted/60 border border-border rounded-xl flex flex-col gap-2">
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground font-medium">Nome de Usuário:</span>
              <span className="font-bold text-foreground">{credentials.username}</span>
            </div>
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground font-medium">E-mail:</span>
              <span className="font-medium text-foreground">{credentials.email}</span>
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-semibold uppercase tracking-wider text-amber-600 dark:text-amber-400 flex items-center gap-1.5">
              <Key className="w-3.5 h-3.5" />
              Senha Temporária Gerada
            </label>

            <div className="relative flex items-center">
              <input
                type={showPassword ? 'text' : 'password'}
                readOnly
                value={credentials.randomPassword || ''}
                className="w-full bg-muted/80 border border-amber-500/40 text-amber-600 dark:text-amber-300 font-mono text-base px-4 py-3 rounded-xl pr-24 focus:outline-none focus:border-amber-500 select-all"
              />
              <div className="absolute right-2 flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="p-1.5 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted transition-colors cursor-pointer"
                  title={showPassword ? 'Ocultar senha' : 'Exibir senha'}
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
                <Button
                  type="button"
                  variant="default"
                  size="sm"
                  className="h-8 px-3 text-xs bg-amber-500 hover:bg-amber-600 text-black font-bold rounded-lg flex items-center gap-1.5 cursor-pointer"
                  onClick={handleCopyPassword}
                >
                  {copiedPassword ? (
                    <>
                      <Check className="w-3.5 h-3.5" />
                      Copiado!
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5" />
                      Copiar
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>

          <div className="p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl flex items-start gap-2.5 text-xs text-amber-600 dark:text-amber-300">
            <AlertCircle className="w-4 h-4 text-amber-500 shrink-0 mt-0.5" />
            <span>
              <strong>Atenção:</strong> Por motivos de segurança, esta senha temporária não será exibida novamente. Certifique-se de copiá-la e enviá-la para o usuário agora.
            </span>
          </div>
        </div>

        <DialogFooter className="mt-4 flex justify-end">
          <Button
            type="button"
            variant="default"
            className="rounded-xl font-bold px-6"
            onClick={onClose}
          >
            Concluído
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
