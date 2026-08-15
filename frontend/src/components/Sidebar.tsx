'use client';

import React from 'react';
import Image from 'next/image';
import { useAuth } from '@/context/AuthContext';
import { FileText, LogOut, Building2, User, Users } from 'lucide-react';
import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

// Cria um componente Link animável para evitar aninhar tags <a> e gerar erros de hidratação
const MotionLink = motion.create(Link);

export const Sidebar: React.FC = () => {
  const { user, logout } = useAuth();
  const pathname = usePathname();

  const isAdmin = user?.role === 'ADMIN';
  const isMod = user?.role === 'MOD';

  return (
    <aside className="w-68 bg-card border-r border-border backdrop-blur-xl flex flex-col justify-between hidden md:flex shrink-0 transition-colors duration-200">
      <div className="flex flex-col flex-1 py-8 px-5 gap-10">
        {/* Logo */}
        <div className="flex items-center gap-3 px-2">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center shadow-lg shadow-violet-500/20 shrink-0">
            <Image
              src="/logo.png"
              alt="docseq Logo"
              width={26}
              height={26}
              className="object-contain"
              priority
            />
          </div>
          <div className="flex flex-col">
            <span className="font-bold text-lg text-foreground leading-none">docseq</span>
            <span className="text-[10px] text-muted-foreground font-medium mt-1">SISTEMA MUNICIPAL</span>
          </div>
        </div>

        {/* Navigation Links */}
        <nav className="flex flex-col gap-2">
          <MotionLink
            whileHover={{ x: 4 }}
            href="/"
            className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
              pathname === '/'
                ? 'bg-violet-600/10 border border-violet-500/20 text-violet-600 dark:text-violet-300 shadow-sm'
                : 'border border-transparent text-muted-foreground hover:text-foreground hover:bg-muted/50'
            }`}
          >
            <FileText className={`w-5 h-5 ${pathname === '/' ? 'text-violet-600 dark:text-violet-400' : 'text-muted-foreground'}`} />
            Documentos Oficiais
          </MotionLink>

          {isAdmin && (
            <MotionLink
              whileHover={{ x: 4 }}
              href="/municipalities"
              className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
                pathname.startsWith('/municipalities')
                  ? 'bg-violet-600/10 border border-violet-500/20 text-violet-600 dark:text-violet-300 shadow-sm'
                  : 'border border-transparent text-muted-foreground hover:text-foreground hover:bg-muted/50'
              }`}
            >
              <Building2 className={`w-5 h-5 ${pathname.startsWith('/municipalities') ? 'text-violet-600 dark:text-violet-400' : 'text-muted-foreground'}`} />
              Municípios
            </MotionLink>
          )}

          {(isAdmin || isMod) && (
            <MotionLink
              whileHover={{ x: 4 }}
              href="/users"
              className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
                pathname.startsWith('/users')
                  ? 'bg-violet-600/10 border border-violet-500/20 text-violet-600 dark:text-violet-300 shadow-sm'
                  : 'border border-transparent text-muted-foreground hover:text-foreground hover:bg-muted/50'
              }`}
            >
              <Users className={`w-5 h-5 ${pathname.startsWith('/users') ? 'text-violet-600 dark:text-violet-400' : 'text-muted-foreground'}`} />
              Usuários
            </MotionLink>
          )}

          <MotionLink
            whileHover={{ x: 4 }}
            href="/profile"
            className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
              pathname.startsWith('/profile')
                ? 'bg-violet-600/10 border border-violet-500/20 text-violet-600 dark:text-violet-300 shadow-sm'
                : 'border border-transparent text-muted-foreground hover:text-foreground hover:bg-muted/50'
            }`}
          >
            <User className={`w-5 h-5 ${pathname.startsWith('/profile') ? 'text-violet-600 dark:text-violet-400' : 'text-muted-foreground'}`} />
            Meu Perfil
          </MotionLink>
        </nav>
      </div>

      {/* User Info & Logout */}
      <div className="p-5 border-t border-border bg-card/60 flex flex-col gap-4">
        {user && (
          <Link
            href="/profile"
            className="flex items-center gap-3 px-2 py-2 -mx-2 rounded-xl hover:bg-muted/60 transition-colors group cursor-pointer"
            title="Ir para o meu perfil"
          >
            <div className="w-10 h-10 rounded-full bg-muted border border-border flex items-center justify-center text-foreground font-semibold uppercase group-hover:border-violet-500/50 group-hover:text-violet-600 dark:group-hover:text-violet-300 transition-colors shrink-0">
              {user.username.slice(0, 2)}
            </div>
            <div className="flex flex-col min-w-0">
              <span className="text-sm font-semibold text-foreground truncate group-hover:text-violet-600 dark:group-hover:text-violet-300 transition-colors">{user.username}</span>
              <span className="text-[10px] text-muted-foreground font-medium uppercase tracking-wider mt-0.5 flex items-center gap-1">
                <User className="w-3 h-3 text-muted-foreground" />
                {user.role}
              </span>
              {user.municipality && (
                <span className="text-[10px] text-muted-foreground truncate mt-1 flex items-center gap-1">
                  <Building2 className="w-3 h-3 text-muted-foreground" />
                  {user.municipality.name} ({user.municipality.uf})
                </span>
              )}
            </div>
          </Link>
        )}
        <Button
          variant="outline"
          className="w-full flex items-center justify-center gap-2 border-border text-muted-foreground hover:border-red-500/30 hover:bg-red-500/5 hover:text-red-500 transition-all py-2.5 rounded-xl"
          onClick={logout}
        >
          <LogOut className="w-4 h-4" />
          Sair do Sistema
        </Button>
      </div>
    </aside>
  );
};
