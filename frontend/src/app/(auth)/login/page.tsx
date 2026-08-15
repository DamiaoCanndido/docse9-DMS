'use client';

import React, { useState } from 'react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '@/context/AuthContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { MessageCircle } from 'lucide-react';

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
    <main className="relative min-h-screen flex flex-col items-center justify-center bg-background text-foreground overflow-hidden px-4 py-8 transition-colors duration-200">
      {/* Background gradients for premium wow effect */}
      <div className="absolute top-[-20%] left-[-10%] w-[600px] h-[600px] rounded-full bg-violet-600/10 blur-[120px] pointer-events-none" />
      <div className="absolute bottom-[-20%] right-[-10%] w-[600px] h-[600px] rounded-full bg-indigo-600/10 blur-[120px] pointer-events-none" />

      <div className="w-full max-w-md flex flex-col items-center gap-6 z-10">
        {/* Login Card */}
        <div className="w-full bg-card border border-border backdrop-blur-xl rounded-2xl p-6 sm:p-8 shadow-2xl relative transition-colors duration-200">
          <div className="flex flex-col items-center gap-2 mb-8">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center shadow-lg shadow-violet-500/30">
              <Image
                src="/logo.png"
                alt="docseq Logo"
                width={32}
                height={32}
                className="object-contain"
                priority
              />
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-foreground mt-2">
              docseq
            </h1>
            <p className="text-sm text-muted-foreground">
              Gerenciamento eletrônico de documentos
            </p>
          </div>

          {error && (
            <div className="bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-sm rounded-lg p-3.5 mb-6 text-center font-medium">
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

        {/* Mobile-first Footer Support & GitHub links */}
        <footer className="flex flex-wrap items-center justify-center gap-3 sm:gap-6 text-xs text-muted-foreground">
          <a
            href="https://wa.me/558396155351?text=Ol%C3%A1%2C%20preciso%20de%20suporte%20no%20docseq.%20Pode%20me%20ajudar%3F"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 hover:text-emerald-500 transition-colors py-1.5 px-3 rounded-lg hover:bg-muted/60 active:scale-95"
            aria-label="Falar com o desenvolvedor no WhatsApp"
          >
            <MessageCircle className="w-4 h-4 text-emerald-500" />
            <span>Suporte WhatsApp</span>
          </a>

          <span className="hidden sm:inline text-border">•</span>

          <a
            href="https://github.com/DamiaoCanndido/docse9-DMS"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 hover:text-foreground transition-colors py-1.5 px-3 rounded-lg hover:bg-muted/60 active:scale-95"
            aria-label="Ver código no GitHub"
          >
            <svg className="w-4 h-4 text-muted-foreground" viewBox="0 0 24 24" fill="currentColor">
              <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
            <span>GitHub do Projeto</span>
          </a>
        </footer>
      </div>
    </main>
  );
}
