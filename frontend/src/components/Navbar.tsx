'use client';

import React from 'react';
import { useAuth } from '@/context/AuthContext';
import { LogOut } from 'lucide-react';
import { Button } from './ui/Button';

export const Navbar: React.FC = () => {
  const { user, logout } = useAuth();

  return (
    <header className="h-16 border-b border-zinc-800 bg-zinc-950 px-6 flex items-center justify-between md:hidden shrink-0">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center text-white font-bold text-md shadow-md">
          D
        </div>
        <div className="flex flex-col">
          <span className="font-bold text-sm text-white">docSe9 DMS</span>
          {user?.municipality && (
            <span className="text-[9px] text-zinc-500 leading-none">
              {user.municipality.name} ({user.municipality.uf})
            </span>
          )}
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
  );
};
