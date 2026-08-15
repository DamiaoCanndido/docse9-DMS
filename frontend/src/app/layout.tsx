import type { Metadata } from 'next';
import { AuthProvider } from '../context/AuthContext';
import { ForceChangePassword } from '@/components/ForceChangePassword';
import { ThemeProvider } from '@/components/ThemeProvider';
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
    <html lang="pt-BR" suppressHydrationWarning className={cn("font-sans", geist.variable)}>
      <body className="antialiased bg-background text-foreground min-h-screen transition-colors duration-200">
        <ThemeProvider attribute="class" defaultTheme="dark" enableSystem disableTransitionOnChange>
          <AuthProvider initialUser={user}>
            {children}
            <ForceChangePassword />
          </AuthProvider>
          <Toaster position="top-right" richColors />
        </ThemeProvider>
      </body>
    </html>
  );
}
