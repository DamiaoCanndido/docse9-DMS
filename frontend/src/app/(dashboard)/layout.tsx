import React from 'react';
import { getMe } from '@/app/api/auth';
import { Sidebar } from '@/components/Sidebar';
import { Navbar } from '@/components/Navbar';

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  // Executa no servidor para proteger as rotas internas e obter os dados do usuário logado
  await getMe();

  return (
    <div className="flex h-screen bg-background text-foreground overflow-hidden transition-colors duration-200">
      {/* Sidebar - Desktop */}
      <Sidebar />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden min-w-0">
        {/* Header - Mobile */}
        <Navbar />

        {/* Page children */}
        <main className="flex-1 overflow-y-auto bg-background p-6 md:p-8 relative transition-colors duration-200">
          <div className="absolute top-0 right-0 w-[400px] h-[400px] rounded-full bg-violet-600/5 blur-[100px] pointer-events-none" />
          <div className="max-w-6xl mx-auto w-full pb-10">
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
