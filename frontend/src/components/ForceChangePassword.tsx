'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { motion, AnimatePresence } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ShieldAlert, LogOut, RefreshCw } from 'lucide-react';
import { toast } from 'sonner';

const changePasswordSchema = z
  .object({
    currentPassword: z.string().min(6, 'A senha temporária deve ter pelo menos 6 caracteres'),
    newPassword: z.string().min(6, 'A nova senha deve ter pelo menos 6 caracteres'),
    confirmPassword: z.string().min(6, 'A confirmação de senha deve ter pelo menos 6 caracteres'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'As novas senhas não coincidem',
    path: ['confirmPassword'],
  });

type ChangePasswordFormData = z.infer<typeof changePasswordSchema>;

export const ForceChangePassword: React.FC = () => {
  const router = useRouter();
  const { user, changePassword, logout } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<ChangePasswordFormData>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    },
  });

  const onSubmit = async (data: ChangePasswordFormData) => {
    setError(null);
    setIsLoading(true);

    try {
      await changePassword(data);
      toast.success('Senha alterada com sucesso! Acesso liberado.');
      reset();
      router.refresh();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Ocorreu um erro ao alterar a senha. Tente novamente.');
      }
      toast.error('Falha ao alterar senha.');
    } finally {
      setIsLoading(false);
    }
  };

  // Only render if user is authenticated and must change password
  if (!user || !user.mustChangePassword) {
    return null;
  }

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-md overflow-y-auto px-4 py-10">
        {/* Background gradients for premium glow */}
        <div className="absolute top-[20%] left-[20%] w-[350px] h-[350px] rounded-full bg-violet-600/10 blur-[100px] pointer-events-none" />
        <div className="absolute bottom-[20%] right-[20%] w-[350px] h-[350px] rounded-full bg-indigo-600/10 blur-[100px] pointer-events-none" />

        <motion.div
          initial={{ opacity: 0, scale: 0.95, y: 10 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 10 }}
          transition={{ duration: 0.3, ease: 'easeOut' }}
          className="w-full max-w-lg bg-card border border-border backdrop-blur-2xl rounded-2xl p-6 md:p-8 shadow-2xl relative"
        >
          {/* Top warning info */}
          <div className="flex flex-col items-center text-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-600 dark:text-violet-400">
              <ShieldAlert className="w-6 h-6 animate-pulse" />
            </div>
            <h1 className="text-xl md:text-2xl font-bold tracking-tight text-foreground mt-1">
              Troca de Senha Obrigatória
            </h1>
            <p className="text-sm text-muted-foreground max-w-sm">
              Sua conta foi criada ou redefinida com uma senha temporária. Para sua segurança, você deve alterá-la antes de continuar.
            </p>
          </div>

          {error && (
            <motion.div
              initial={{ opacity: 0, y: -5 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-sm rounded-lg p-3.5 mb-6 text-center font-medium"
            >
              {error}
            </motion.div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5">
            <Input
              label="Senha Temporária"
              type="password"
              placeholder="Digite a senha aleatória recebida"
              disabled={isLoading}
              error={errors.currentPassword?.message}
              {...register('currentPassword')}
            />

            <Input
              label="Nova Senha"
              type="password"
              placeholder="Mínimo 6 caracteres"
              disabled={isLoading}
              error={errors.newPassword?.message}
              {...register('newPassword')}
            />

            <Input
              label="Confirmar Nova Senha"
              type="password"
              placeholder="Repita a nova senha"
              disabled={isLoading}
              error={errors.confirmPassword?.message}
              {...register('confirmPassword')}
            />

            <div className="flex flex-col gap-3 mt-4">
              <Button
                type="submit"
                className="w-full h-11 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 flex items-center justify-center gap-2"
                isLoading={isLoading}
              >
                <RefreshCw className="w-4 h-4" />
                Atualizar Senha
              </Button>
              <Button
                type="button"
                variant="outline"
                className="w-full flex items-center justify-center gap-2 border-border text-muted-foreground hover:border-red-500/30 hover:bg-red-500/5 hover:text-red-500 transition-all py-2.5 h-11"
                onClick={logout}
                disabled={isLoading}
              >
                <LogOut className="w-4 h-4" />
                Sair
              </Button>
            </div>
          </form>
        </motion.div>
      </div>
    </AnimatePresence>
  );
};
