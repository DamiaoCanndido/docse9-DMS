'use client';

import React from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Home, ArrowLeft, FileQuestion, MessageCircle, HelpCircle } from 'lucide-react';

export default function NotFound() {
  const router = useRouter();

  return (
    <main className="relative min-h-screen flex flex-col items-center justify-center bg-background text-foreground overflow-hidden px-4 py-12 transition-colors duration-200">
      {/* Glow / blur background accents matching docseq theme */}
      <div
        className="absolute top-[-15%] left-[-10%] w-[500px] h-[500px] rounded-full bg-teal-600/10 blur-[130px] pointer-events-none"
        aria-hidden="true"
      />
      <div
        className="absolute bottom-[-15%] right-[-10%] w-[500px] h-[500px] rounded-full bg-teal-500/10 blur-[130px] pointer-events-none"
        aria-hidden="true"
      />

      <div className="w-full max-w-lg flex flex-col items-center gap-6 z-10">
        {/* Central 404 Glass Card */}
        <div className="w-full bg-card/85 border border-border backdrop-blur-xl rounded-3xl p-8 sm:p-10 shadow-2xl relative flex flex-col items-center text-center transition-colors duration-200">
          {/* Animated / styled Icon Badge */}
          <div className="relative mb-5">
            <div className="w-20 h-20 rounded-2xl bg-gradient-to-tr from-teal-600 to-teal-500 flex items-center justify-center shadow-lg shadow-teal-500/30 text-white">
              <FileQuestion className="w-10 h-10 animate-in zoom-in duration-300" />
            </div>
            <div className="absolute -bottom-2 -right-2 px-2.5 py-0.5 rounded-full bg-background border border-border text-[11px] font-bold text-teal-600 dark:text-teal-400 shadow-sm tracking-wider uppercase">
              404
            </div>
          </div>

          <span className="text-xs font-bold uppercase tracking-widest text-teal-600 dark:text-teal-400 mb-1.5 flex items-center gap-1.5">
            <HelpCircle className="w-3.5 h-3.5" />
            Erro de Navegação
          </span>

          <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground mb-3">
            Página não encontrada
          </h1>

          <p className="text-sm sm:text-base text-muted-foreground max-w-sm mb-8 leading-relaxed">
            O endereço ou documento que você tentou acessar não foi encontrado, mudou de lugar ou você não possui permissão para visualizá-lo.
          </p>

          {/* Action Buttons */}
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-center gap-3 w-full">
            <Link href="/" className="flex-1">
              <Button className="w-full h-11 px-5 gap-2 font-semibold shadow-md shadow-teal-500/20">
                <Home className="w-4 h-4" />
                Ir para o Início
              </Button>
            </Link>

            <Button
              variant="outline"
              onClick={() => router.back()}
              className="flex-1 h-11 px-5 gap-2 font-medium"
            >
              <ArrowLeft className="w-4 h-4" />
              Voltar Página
            </Button>
          </div>
        </div>

        {/* Support & Quick Help Footer */}
        <footer className="flex flex-wrap items-center justify-center gap-4 text-xs text-muted-foreground">
          <span>Precisa de assistência técnica?</span>
          <a
            href="https://wa.me/558396155351?text=Ol%C3%A1%2C%20encontrei%20um%20erro%20404%20no%20docseq.%20Pode%20me%20ajudar%3F"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 text-teal-600 dark:text-teal-400 hover:text-teal-500 font-medium transition-colors py-1 px-2 rounded-lg hover:bg-muted/60"
            aria-label="Falar com o suporte no WhatsApp"
          >
            <MessageCircle className="w-3.5 h-3.5 text-emerald-500" />
            <span>Suporte WhatsApp</span>
          </a>
        </footer>
      </div>
    </main>
  );
}
