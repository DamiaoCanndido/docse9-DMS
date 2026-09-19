'use client';

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/context/AuthContext';
import { APP_VERSION } from '@/lib/version';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogAction,
} from '@/components/ui/alert-dialog';
import {
  Sparkles,
  Bell,
  Clock,
  ShieldCheck,
  ArrowRight,
  FileSignature,
} from 'lucide-react';

export const WhatsNewAlertDialog: React.FC = () => {
  const { user } = useAuth();
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    if (!user || user.mustChangePassword) return;

    // Chave única de visualização vinculada ao ID do usuário e à versão atual
    const storageKey = `docseq_whats_new_seen_${user.id}_${APP_VERSION}`;

    const timer = setTimeout(() => {
      try {
        const hasSeen = localStorage.getItem(storageKey);
        if (!hasSeen) {
          setIsOpen(true);
        }
      } catch (err) {
        console.error('Erro ao verificar status do WhatsNewAlertDialog no localStorage:', err);
      }
    }, 400);

    return () => clearTimeout(timer);
  }, [user]);

  const handleDismiss = () => {
    if (user) {
      try {
        const storageKey = `docseq_whats_new_seen_${user.id}_${APP_VERSION}`;
        localStorage.setItem(storageKey, 'true');
      } catch (err) {
        console.error('Erro ao salvar status do WhatsNewAlertDialog no localStorage:', err);
      }
    }
    setIsOpen(false);
  };

  if (!isOpen) return null;

  return (
    <AlertDialog open={isOpen} onOpenChange={(open) => !open && handleDismiss()}>
      <AlertDialogContent className="max-w-lg p-0 overflow-hidden border-border bg-card shadow-2xl rounded-2xl">
        {/* Top Header Banner with Gradient Accent */}
        <div className="bg-gradient-to-tr from-teal-600 to-teal-500 p-6 text-white relative overflow-hidden">
          <div className="absolute top-0 right-0 w-32 h-32 bg-white/10 rounded-full blur-2xl pointer-events-none" />
          
          <div className="flex items-center gap-2 mb-2">
            <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-extrabold bg-white/20 text-white backdrop-blur-md border border-white/20">
              <Sparkles className="w-3.5 h-3.5" />
              Novidades do Sistema
            </span>
            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-white/10 text-white/90 border border-white/15">
              {APP_VERSION}
            </span>
          </div>

          <AlertDialogTitle className="text-xl font-extrabold text-white tracking-tight">
            Bem-vindo às novidades do Docseq!
          </AlertDialogTitle>
          <AlertDialogDescription className="text-white/80 text-xs mt-1 leading-relaxed">
            Confira os novos recursos que preparamos para modernizar ainda mais a gestão de documentos de {user?.municipality?.name || 'seu município'}.
          </AlertDialogDescription>
        </div>

        {/* Feature Highlights Body */}
        <div className="p-6 flex flex-col gap-4">
          <div className="flex items-start gap-3.5 p-3 rounded-xl bg-muted/40 border border-border/70 hover:border-teal-500/30 transition-colors">
            <div className="w-9 h-9 rounded-xl bg-teal-600/10 border border-teal-500/20 text-teal-600 dark:text-teal-400 flex items-center justify-center shrink-0">
              <Bell className="w-4.5 h-4.5" />
            </div>
            <div>
              <h4 className="text-xs font-bold text-foreground">
                Central de Notificações no Topo
              </h4>
              <p className="text-[11px] text-muted-foreground mt-0.5 leading-relaxed">
                Acesse o novo sino no cabeçalho para ver notas de atualização da plataforma e alertas de prazos importantes em tempo real.
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3.5 p-3 rounded-xl bg-muted/40 border border-border/70 hover:border-teal-500/30 transition-colors">
            <div className="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-600 dark:text-amber-400 flex items-center justify-center shrink-0">
              <Clock className="w-4.5 h-4.5" />
            </div>
            <div>
              <h4 className="text-xs font-bold text-foreground flex items-center gap-1.5">
                Alertas de Vigência de Contratos
                <span className="text-[9px] font-extrabold text-amber-600 dark:text-amber-400 bg-amber-500/10 px-1.5 py-0.2 rounded-md">
                  Automático
                </span>
              </h4>
              <p className="text-[11px] text-muted-foreground mt-0.5 leading-relaxed">
                O sistema monitora automaticamente o término da vigência de contratos, alertando com 30, 60 e 90 dias de antecedência para renovações e aditivos.
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3.5 p-3 rounded-xl bg-muted/40 border border-border/70 hover:border-teal-500/30 transition-colors">
            <div className="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0">
              <ShieldCheck className="w-4.5 h-4.5" />
            </div>
            <div>
              <h4 className="text-xs font-bold text-foreground">
                Identificação de Versão Dinâmica
              </h4>
              <p className="text-[11px] text-muted-foreground mt-0.5 leading-relaxed">
                Acompanhe a versão oficial do sistema integrada na tela de login, barra lateral e painel principal.
              </p>
            </div>
          </div>
        </div>

        {/* Action Footer */}
        <AlertDialogFooter className="p-4 bg-muted/30 border-t border-border flex items-center justify-between">
          <span className="text-[10px] text-muted-foreground flex items-center gap-1">
            <FileSignature className="w-3.5 h-3.5 text-teal-600 dark:text-teal-400" />
            Você pode rever essas novidades a qualquer momento no sino de notificações.
          </span>
          <AlertDialogAction
            onClick={handleDismiss}
            className="w-full sm:w-auto h-10 px-5 rounded-xl font-bold bg-teal-600 text-white hover:bg-teal-700 shadow-md shadow-teal-600/20 flex items-center gap-2 cursor-pointer"
          >
            <span>Entendi, vamos começar</span>
            <ArrowRight className="w-4 h-4" />
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};
