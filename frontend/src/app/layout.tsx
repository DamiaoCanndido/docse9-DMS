import type { Metadata } from 'next';
import { AuthProvider } from '../context/AuthContext';
import { ForceChangePassword } from '@/components/ForceChangePassword';
import { getUserOrNull } from './api/auth';
import './globals.css';
import { Toaster } from 'sonner';
import { Geist } from "next/font/google";
import { cn } from "@/lib/utils";

const geist = Geist({subsets:['latin'],variable:'--font-sans'});

export const metadata: Metadata = {
  title: 'docseq - Gerenciamento de Documentos',
  description: 'Sistema Avançado de Gerenciamento Eletrônico de Documentos',
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const user = await getUserOrNull();

  return (
    <html lang="pt-BR" className={cn("dark", "font-sans", geist.variable)} style={{ colorScheme: 'dark' }}>
      <body className="antialiased bg-black text-zinc-100 min-h-screen">
        <AuthProvider initialUser={user}>
          {children}
          <ForceChangePassword />
        </AuthProvider>
        <Toaster position="top-right" richColors />
      </body>
    </html>
  );
}
