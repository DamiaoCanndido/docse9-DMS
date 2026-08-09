'use client';

import React from 'react';
import { useAuth } from '../../context/AuthContext';
import { Button } from '../../components/ui/Button';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();

  return (
    <div className="flex h-screen bg-black overflow-hidden">
      {/* Sidebar */}
      <aside className="w-64 bg-zinc-950 border-r border-border flex flex-col justify-between hidden md:flex">
        <div className="flex flex-col flex-1 py-6 px-4 gap-8">
          {/* Logo */}
          <div className="flex items-center gap-3 px-2">
            <div className="w-9 h-9 rounded-lg bg-primary flex items-center justify-center text-white font-bold text-lg shadow-lg shadow-violet-500/20">
              D
            </div>
            <span className="font-bold text-lg text-white">docSe9 DMS</span>
          </div>

          {/* Navigation Links */}
          <nav className="flex flex-col gap-1.5">
            <a
              href="#"
              className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium bg-zinc-900 text-white transition-all"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
              </svg>
              Documentos
            </a>
          </nav>
        </div>

        {/* User Info & Logout */}
        <div className="p-4 border-t border-border bg-zinc-950/50 flex flex-col gap-4">
          {user && (
            <div className="flex flex-col gap-0.5 px-2">
              <span className="text-sm font-semibold text-white truncate">{user.username}</span>
              <span className="text-xs text-zinc-400 uppercase tracking-wider">{user.role}</span>
              {user.municipality && (
                <span className="text-xs text-zinc-500 truncate mt-1">
                  {user.municipality.name} ({user.municipality.uf})
                </span>
              )}
            </div>
          )}
          <Button variant="outline" className="w-full flex items-center gap-2" onClick={logout}>
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Sair do Sistema
          </Button>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header (Mobile Logo & Layout) */}
        <header className="h-16 border-b border-border bg-zinc-950 flex items-center justify-between px-6 md:hidden">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center text-white font-bold text-md">
              D
            </div>
            <span className="font-bold text-md text-white">docSe9 DMS</span>
          </div>
          <Button variant="outline" className="px-3 py-1.5 text-xs" onClick={logout}>
            Sair
          </Button>
        </header>

        {/* Page children */}
        <main className="flex-1 overflow-y-auto bg-black p-6 relative">
          <div className="absolute top-0 right-0 w-[400px] h-[400px] rounded-full bg-violet-600/5 blur-[100px] pointer-events-none" />
          <div className="max-w-6xl mx-auto">{children}</div>
        </main>
      </div>
    </div>
  );
}
