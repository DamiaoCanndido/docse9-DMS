'use client';

import React, { useState } from 'react';
import Image from 'next/image';
import { useAuth } from '@/context/AuthContext';
import { LogOut, Menu, X, FileText, Building2, User, Users } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { motion, AnimatePresence } from 'framer-motion';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

const MotionLink = motion.create(Link);

export const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);

  const isAdmin = user?.role === 'ADMIN';
  const isMod = user?.role === 'MOD';

  const closeMenu = () => setIsOpen(false);

  return (
    <>
      <header className="h-16 border-b border-zinc-800 bg-zinc-950 px-4 flex items-center justify-between md:hidden shrink-0">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="icon"
            className="h-9 w-9 text-zinc-400 hover:text-white hover:bg-zinc-900 border border-zinc-800/50"
            onClick={() => setIsOpen(true)}
            aria-label="Abrir menu de navegação"
          >
            <Menu className="w-5 h-5" />
          </Button>

          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center shadow-md shrink-0">
              <Image
                src="/logo.png"
                alt="docseq Logo"
                width={20}
                height={20}
                className="object-contain"
              />
            </div>
            <div className="flex flex-col">
              <span className="font-bold text-sm text-white leading-tight">docseq</span>
              {user?.municipality && (
                <span className="text-[9px] text-zinc-500 leading-none">
                  {user.municipality.name} ({user.municipality.uf})
                </span>
              )}
            </div>
          </div>
        </div>

        <Button
          variant="outline"
          className="px-3 py-1.5 text-xs h-8 border-zinc-800 hover:border-red-500/30 hover:bg-red-500/5 hover:text-red-400 gap-1 flex items-center"
          onClick={logout}
        >
          <LogOut className="w-3.5 h-3.5" />
          Sair
        </Button>
      </header>

      {/* Drawer Mobile Navigation */}
      <AnimatePresence>
        {isOpen && (
          <>
            {/* Backdrop Overlay */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={closeMenu}
              className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 md:hidden"
            />

            {/* Drawer Container */}
            <motion.div
              initial={{ x: '-100%' }}
              animate={{ x: 0 }}
              exit={{ x: '-100%' }}
              transition={{ type: 'spring', damping: 25, stiffness: 200 }}
              className="fixed top-0 left-0 bottom-0 w-72 bg-zinc-950 border-r border-zinc-800 z-50 md:hidden flex flex-col justify-between py-6 px-5 shadow-2xl"
            >
              <div className="flex flex-col gap-8">
                {/* Header of Drawer */}
                <div className="flex items-center justify-between pb-4 border-b border-zinc-900">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center shadow-lg shrink-0">
                      <Image
                        src="/logo.png"
                        alt="docseq Logo"
                        width={24}
                        height={24}
                        className="object-contain"
                      />
                    </div>
                    <div className="flex flex-col">
                      <span className="font-bold text-md text-white leading-none">docseq</span>
                      <span className="text-[9px] text-zinc-500 font-medium mt-1">SISTEMA MUNICIPAL</span>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-zinc-400 hover:text-white hover:bg-zinc-900"
                    onClick={closeMenu}
                    aria-label="Fechar menu"
                  >
                    <X className="w-4 h-4" />
                  </Button>
                </div>

                {/* Navigation Links inside Drawer */}
                <nav className="flex flex-col gap-2">
                  <MotionLink
                    whileHover={{ x: 4 }}
                    href="/"
                    onClick={closeMenu}
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
                      href="/municipalities"
                      onClick={closeMenu}
                      className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
                        pathname.startsWith('/municipalities')
                          ? 'bg-gradient-to-r from-violet-600/15 to-indigo-600/15 border border-violet-500/20 text-violet-300 shadow-md shadow-violet-500/5'
                          : 'border border-transparent text-zinc-400 hover:text-zinc-200'
                      }`}
                    >
                      <Building2 className={`w-5 h-5 ${pathname.startsWith('/municipalities') ? 'text-violet-400' : 'text-zinc-500'}`} />
                      Municípios
                    </MotionLink>
                  )}

                  {(isAdmin || isMod) && (
                    <MotionLink
                      whileHover={{ x: 4 }}
                      href="/users"
                      onClick={closeMenu}
                      className={`flex items-center gap-3.5 px-4 py-3 rounded-xl text-sm font-medium transition-all cursor-pointer ${
                        pathname.startsWith('/users')
                          ? 'bg-gradient-to-r from-violet-600/15 to-indigo-600/15 border border-violet-500/20 text-violet-300 shadow-md shadow-violet-500/5'
                          : 'border border-transparent text-zinc-400 hover:text-zinc-200'
                      }`}
                    >
                      <Users className={`w-5 h-5 ${pathname.startsWith('/users') ? 'text-violet-400' : 'text-zinc-500'}`} />
                      Usuários
                    </MotionLink>
                  )}
                </nav>
              </div>

              {/* User Info & Logout inside Drawer */}
              <div className="pt-4 border-t border-zinc-900 flex flex-col gap-4">
                {user && (
                  <div className="flex items-center gap-3 px-1">
                    <div className="w-9 h-9 rounded-full bg-zinc-800 border border-zinc-700 flex items-center justify-center text-zinc-300 font-semibold uppercase text-sm">
                      {user.username.slice(0, 2)}
                    </div>
                    <div className="flex flex-col min-w-0">
                      <span className="text-xs font-semibold text-white truncate">{user.username}</span>
                      <span className="text-[9px] text-zinc-400 font-medium uppercase tracking-wider mt-0.5 flex items-center gap-1">
                        <User className="w-3 h-3 text-zinc-500" />
                        {user.role}
                      </span>
                      {user.municipality && (
                        <span className="text-[9px] text-zinc-500 truncate mt-1 flex items-center gap-1">
                          <Building2 className="w-3 h-3 text-zinc-600" />
                          {user.municipality.name} ({user.municipality.uf})
                        </span>
                      )}
                    </div>
                  </div>
                )}
                <Button
                  variant="outline"
                  className="w-full flex items-center justify-center gap-2 border-zinc-800 hover:border-red-500/30 hover:bg-red-500/5 hover:text-red-400 transition-all py-2 rounded-xl text-zinc-400 text-xs h-9"
                  onClick={() => {
                    closeMenu();
                    logout();
                  }}
                >
                  <LogOut className="w-3.5 h-3.5" />
                  Sair do Sistema
                </Button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
};
