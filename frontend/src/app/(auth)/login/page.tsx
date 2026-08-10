'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '@/context/AuthContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

const loginSchema = z.object({
  username: z.string().min(3, 'O usuário deve ter pelo menos 3 caracteres'),
  password: z.string().min(6, 'A senha deve ter pelo menos 6 caracteres'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    setError(null);
    setIsLoading(true);

    try {
      await login(data.username, data.password);
      router.refresh();
      router.push('/');
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Falha ao autenticar. Verifique suas credenciais.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main className="relative min-h-screen flex items-center justify-center bg-black overflow-hidden px-4">
      {/* Background gradients for premium wow effect */}
      <div className="absolute top-[-20%] left-[-10%] w-[600px] h-[600px] rounded-full bg-violet-600/10 blur-[120px] pointer-events-none" />
      <div className="absolute bottom-[-20%] right-[-10%] w-[600px] h-[600px] rounded-full bg-indigo-600/10 blur-[120px] pointer-events-none" />

      {/* Login Card */}
      <div className="w-full max-w-md bg-zinc-950/40 border border-zinc-800/80 backdrop-blur-xl rounded-2xl p-8 shadow-2xl relative">
        <div className="flex flex-col items-center gap-2 mb-8">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center shadow-lg shadow-violet-500/30 text-white font-bold text-xl">
            D
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-white mt-2">
            docSe9 DMS
          </h1>
          <p className="text-sm text-zinc-400">
            Gerenciamento eletrônico de documentos
          </p>
        </div>

        {error && (
          <div className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm rounded-lg p-3.5 mb-6 text-center font-medium">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5">
          <Input
            label="Usuário"
            placeholder="Nome de usuário"
            disabled={isLoading}
            autoComplete="username"
            error={errors.username?.message}
            {...register('username')}
          />

          <Input
            label="Senha"
            type="password"
            placeholder="Sua senha secreta"
            disabled={isLoading}
            autoComplete="current-password"
            error={errors.password?.message}
            {...register('password')}
          />

          <Button
            type="submit"
            className="w-full mt-4 h-11"
            isLoading={isLoading}
          >
            Acessar Sistema
          </Button>
        </form>
      </div>
    </main>
  );
}
