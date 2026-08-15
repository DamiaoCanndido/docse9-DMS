'use client';

import React, { useState, useSyncExternalStore } from 'react';
import Image from 'next/image';
import { useAuth } from '@/context/AuthContext';
import { useTheme } from 'next-themes';
import { LogOut, Menu, X, FileText, Building2, User, Users, Sun, Moon } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { motion, AnimatePresence } from 'framer-motion';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

const MotionLink = motion.create(Link);
const emptySubscribe = () => () => {};

export const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const { setTheme, resolvedTheme } = useTheme();
  const mounted = useSyncExternalStore(emptySubscribe, () => true, () => false);
  const [isOpen, setIsOpen] = useState(false);

  const isAdmin = user?.role === 'ADMIN';
  const isMod = user?.role === 'MOD';

  const closeMenu = () => setIsOpen(false);

  const toggleTheme = () => {
    if (resolvedTheme === 'dark') {
      setTheme('light');
    } else {
      setTheme('dark');
    }
  };

  const getPageTitle = () => {
    if (pathname.startsWith('/municipalities')) return 'Gestão de Municípios';
    if (pathname.startsWith('/users')) return 'Gestão de Usuários';
    if (pathname.startsWith('/profile')) return 'Meu Perfil';
    return 'Documentos Oficiais';
  };

  return (
    <>
      {/* Mobile Header (< md) */}
      <header className="h-16 border-b border-border bg-card/90 backdrop-blur-md px-4 flex items-center justify-between md:hidden shrink-0 transition-colors duration-200 z-20">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="icon"
            className="h-9 w-9 text-muted-foreground hover:text-foreground hover:bg-muted border border-border"
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
                priority
              />
            </div>
            <div className="flex flex-col">
              <span className="font-bold text-sm text-foreground leading-tight">docseq</span>
              {user?.municipality && (
                <span className="text-[9px] text-muted-foreground leading-none">
                  {user.municipality.name} ({user.municipality.uf})
                </span>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Quick Theme Toggle Mobile */}
          {mounted && (
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-muted-foreground hover:text-foreground hover:bg-muted border border-border rounded-lg"
              onClick={toggleTheme}
              aria-label="Alternar tema claro/escuro"
              title={resolvedTheme === 'dark' ? 'Mudar para Tema Claro' : 'Mudar para Tema Escuro'}
            >
              {resolvedTheme === 'dark' ? (
                <Sun className="w-4 h-4 text-amber-400 animate-in fade-in zoom-in duration-200" />
              ) : (
                <Moon className="w-4 h-4 text-violet-600 animate-in fade-in zoom-in duration-200" />
              )}
            </Button>
          )}

          <Button
            variant="outline"
            className="px-2.5 py-1 text-xs h-8 border-border text-muted-foreground hover:border-red-500/30 hover:bg-red-500/5 hover:text-red-500 gap-1 flex items-center"
            onClick={logout}
          >
            <LogOut className="w-3.5 h-3.5" />
            Sair
          </Button>
        </div>
      </header>

      {/* Desktop Top Navbar (>= md) */}
      <header className="h-16 border-b border-border bg-card/80 backdrop-blur-md px-8 hidden md:flex items-center justify-between shrink-0 transition-colors duration-200 z-20">
        <div className="flex items-center gap-4">
          <h2 className="text-base font-bold text-foreground flex items-center gap-2">
            {getPageTitle()}
          </h2>

          {user?.municipality && (
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-muted border border-border text-foreground">
              <Building2 className="w-3.5 h-3.5 text-violet-600 dark:text-violet-400" />
              <span>{user.municipality.name}</span>
              <span className="text-[10px] text-muted-foreground">({user.municipality.uf})</span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-3">
          {/* Quick Theme Toggle Desktop */}
          {mounted && (
            <Button
              variant="outline"
              size="sm"
              className="h-9 px-3 rounded-xl border-border bg-background hover:bg-muted text-foreground flex items-center gap-2 text-xs font-medium cursor-pointer transition-colors shadow-xs"
              onClick={toggleTheme}
              title={resolvedTheme === 'dark' ? 'Mudar para Tema Claro' : 'Mudar para Tema Escuro'}
            >
              {resolvedTheme === 'dark' ? (
                <>
                  <Sun className="w-4 h-4 text-amber-400" />
                  <span>Modo Claro</span>
                </>
              ) : (
                <>
                  <Moon className="w-4 h-4 text-violet-600" />
                  <span>Modo Escuro</span>
                </>
              )}
            </Button>
          )}

          {/* User Profile Pill */}
          {user && (
            <Link
              href="/profile"
              className="flex items-center gap-2.5 px-3 py-1.5 rounded-xl border border-border bg-background hover:bg-muted text-foreground transition-all cursor-pointer group shadow-xs"
              title="Meu perfil e configurações"
            >
              <div className="w-6 h-6 rounded-full bg-violet-600/10 border border-violet-500/20 text-violet-600 dark:text-violet-300 flex items-center justify-center font-bold text-[11px] uppercase group-hover:bg-violet-600 group-hover:text-white transition-colors">
                {user.username.slice(0, 2)}
              </div>
              <span className="text-xs font-semibold text-foreground group-hover:text-violet-600 dark:group-hover:text-violet-400 transition-colors">
                {user.username}
              </span>
            </Link>
          )}
        </div>
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
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 md:hidden"
            />

            {/* Drawer Container */}
            <motion.div
              initial={{ x: '-100%' }}
              animate={{ x: 0 }}
              exit={{ x: '-100%' }}
              transition={{ type: 'spring', damping: 25, stiffness: 200 }}
              className="fixed top-0 left-0 bottom-0 w-72 bg-card border-r border-border z-50 md:hidden flex flex-col justify-between py-6 px-5 shadow-2xl transition-colors duration-200"
            >
              <div className="flex flex-col gap-8">
                {/* Header of Drawer */}
                <div className="flex items-center justify-between pb-4 border-b border-border">
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
                      <span className="font-bold text-md text-foreground leading-none">docseq</span>
                      <span className="text-[9px] text-muted-foreground font-medium mt-1">SISTEMA MUNICIPAL</span>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-muted-foreground hover:text-foreground hover:bg-muted"
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
                      onClick={closeMenu}
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
                      onClick={closeMenu}
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
                    onClick={closeMenu}
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

              {/* User Info & Logout inside Drawer */}
              <div className="pt-4 border-t border-border flex flex-col gap-4">
                {user && (
                  <Link
                    href="/profile"
                    onClick={closeMenu}
                    className="flex items-center gap-3 px-2 py-1.5 -mx-2 rounded-xl hover:bg-muted/60 transition-colors group cursor-pointer"
                    title="Ir para o meu perfil"
                  >
                    <div className="w-9 h-9 rounded-full bg-muted border border-border flex items-center justify-center text-foreground font-semibold uppercase text-sm group-hover:border-violet-500/50 group-hover:text-violet-600 dark:group-hover:text-violet-300 transition-colors shrink-0">
                      {user.username.slice(0, 2)}
                    </div>
                    <div className="flex flex-col min-w-0">
                      <span className="text-xs font-semibold text-foreground truncate group-hover:text-violet-600 dark:group-hover:text-violet-300 transition-colors">{user.username}</span>
                      <span className="text-[9px] text-muted-foreground font-medium uppercase tracking-wider mt-0.5 flex items-center gap-1">
                        <User className="w-3 h-3 text-muted-foreground" />
                        {user.role}
                      </span>
                      {user.municipality && (
                        <span className="text-[9px] text-muted-foreground truncate mt-1 flex items-center gap-1">
                          <Building2 className="w-3 h-3 text-muted-foreground" />
                          {user.municipality.name} ({user.municipality.uf})
                        </span>
                      )}
                    </div>
                  </Link>
                )}
                <Button
                  variant="outline"
                  className="w-full flex items-center justify-center gap-2 border-border text-muted-foreground hover:border-red-500/30 hover:bg-red-500/5 hover:text-red-500 transition-all py-2 rounded-xl text-xs h-9"
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
