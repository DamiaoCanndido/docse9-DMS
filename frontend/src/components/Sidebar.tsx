'use client';

import React from 'react';
import { useAuth } from '@/context/AuthContext';
import { FileText, LogOut, Building2, User, Users } from 'lucide-react';
import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

// Cria um componente Link animável para evitar aninhar tags <a> e gerar erros de hidratação
const MotionLink = motion(Link);

export const Sidebar: React.FC = () => {
  const { user, logout } = useAuth();
  const pathname = usePathname();

  const isAdmin = user?.role === 'ADMIN';
  const isMod = user?.role === 'MOD';

  return (
    <aside className="w-68 bg-zinc-950/80 border-r border-zinc-800/80 backdrop-blur-xl flex flex-col justify-between hidden md:flex shrink-0">
      <div className="flex flex-col flex-1 py-8 px-5 gap-10">
        {/* Logo */}
        <div className="flex items-center gap-3 px-2">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center text-white font-bold text-xl shadow-lg shadow-violet-500/20">
            D
          </div>
          <div className="flex flex-col">
            <span className="font-bold text-lg text-white leading-none">docSe9 DMS</span>
            <span className="text-[10px] text-zinc-500 font-medium mt-1">SISTEMA MUNICIPAL</span>
          </div>
        </div>

        {/* Navigation Links */}
        <nav className="flex flex-col gap-2">
          <MotionLink
            whileHover={{ x: 4 }}
            href="/"
            className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
              pathname === '/'
                ? 'bg-gradient-to-r from-violet-600/15 to-indigo-600/15 border border-violet-500/20 text-violet-300 shadow-md shadow-violet-500/5'
                : 'border border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <FileText className={`w-5 h-5 ${pathname === '/' ? 'text-violet-400' : 'text-zinc-500'}`} />
            Documentos Oficiais
          </MotionLink>

          {isAdmin && (
            <MotionLink
              whileHover={{ x: 4 }}
              href="/dashboard/municipalities"
              className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
                pathname.startsWith('/dashboard/municipalities')
                  ? 'bg-gradient-to-r from-violet-600/15 to-indigo-600/15 border border-violet-500/20 text-violet-300 shadow-md shadow-violet-500/5'
                  : 'border border-transparent text-zinc-400 hover:text-zinc-200'
              }`}
            >
              <Building2 className={`w-5 h-5 ${pathname.startsWith('/dashboard/municipalities') ? 'text-violet-400' : 'text-zinc-500'}`} />
              Municípios
            </MotionLink>
          )}

          {(isAdmin || isMod) && (
            <MotionLink
              whileHover={{ x: 4 }}
              href="/dashboard/users"
              className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
                pathname.startsWith('/dashboard/users')
                  ? 'bg-gradient-to-r from-violet-600/15 to-indigo-600/15 border border-violet-500/20 text-violet-300 shadow-md shadow-violet-500/5'
                  : 'border border-transparent text-zinc-400 hover:text-zinc-200'
              }`}
            >
              <Users className={`w-5 h-5 ${pathname.startsWith('/dashboard/users') ? 'text-violet-400' : 'text-zinc-500'}`} />
              Usuários
            </MotionLink>
          )}
        </nav>
      </div>



      {/* User Info & Logout */}
      <div className="p-5 border-t border-zinc-800/80 bg-zinc-950/40 flex flex-col gap-5">
        {user && (
          <div className="flex items-center gap-3 px-1">
            <div className="w-10 h-10 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 font-semibold uppercase">
              {user.username.slice(0, 2)}
            </div>
            <div className="flex flex-col min-w-0">
              <span className="text-sm font-semibold text-white truncate">{user.username}</span>
              <span className="text-[10px] text-zinc-400 font-medium uppercase tracking-wider mt-0.5 flex items-center gap-1">
                <User className="w-3 h-3 text-zinc-500" />
                {user.role}
              </span>
              {user.municipality && (
                <span className="text-[10px] text-zinc-500 truncate mt-1 flex items-center gap-1">
                  <Building2 className="w-3 h-3 text-zinc-600" />
                  {user.municipality.name} ({user.municipality.uf})
                </span>
              )}
            </div>
          </div>
        )}
        <Button
          variant="outline"
          className="w-full flex items-center justify-center gap-2 border-zinc-800 hover:border-red-500/30 hover:bg-red-500/5 hover:text-red-400 transition-all py-2.5 rounded-xl text-zinc-400"
          onClick={logout}
        >
          <LogOut className="w-4 h-4" />
          Sair do Sistema
        </Button>
      </div>
    </aside>
  );
};
