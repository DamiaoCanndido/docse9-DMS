'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { useTheme } from 'next-themes';
import { User, UserPermission, DocumentType } from '@/types';
import { updateUserProfile, changePassword } from '@/app/api/auth';
import { getUserPermissions } from '@/app/api/users';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { isRedirectError } from '@/lib/utils';
import { motion, AnimatePresence } from 'framer-motion';
import {
  User as UserIcon,
  Lock,
  Sun,
  Moon,
  Laptop,
  Building2,
  Mail,
  Shield,
  Calendar,
  CheckCircle2,
  Eye,
  EyeOff,
  Sparkles,
  RefreshCw,
  Save,
  KeyRound,
  FileText,
  FileCheck,
  Scale,
  FileSignature,
  Clock,
} from 'lucide-react';

interface ProfileContentProps {
  currentUser: User;
}

export const ProfileContent: React.FC<ProfileContentProps> = ({ currentUser }) => {
  const router = useRouter();
  const { user, updateUser } = useAuth();
  const { theme, setTheme } = useTheme();

  // Active section state
  const [activeTab, setActiveTab] = useState<'profile' | 'security' | 'appearance' | 'permissions'>('profile');

  // Profile edit form state
  const [username, setUsername] = useState(currentUser.username);
  const [email, setEmail] = useState(currentUser.email);
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [profileError, setProfileError] = useState('');

  // Password change form state
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [isChangingPassword, setIsChangingPassword] = useState(false);
  const [passwordError, setPasswordError] = useState('');

  // Permissions state for COMMON users
  const [permissions, setPermissions] = useState<UserPermission[]>([]);
  const [isLoadingPerms, setIsLoadingPerms] = useState(currentUser.role === 'COMMON');

  const mounted = React.useSyncExternalStore(
    () => () => {},
    () => true,
    () => false
  );

  useEffect(() => {
    let isCancelled = false;
    if (currentUser.role === 'COMMON') {
      getUserPermissions(currentUser.id)
        .then((perms) => {
          if (!isCancelled) setPermissions(perms);
        })
        .catch((err) => {
          if (!isCancelled) console.error('Erro ao carregar permissões:', err);
        })
        .finally(() => {
          if (!isCancelled) setIsLoadingPerms(false);
        });
    }
    return () => {
      isCancelled = true;
    };
  }, [currentUser]);

  // Handle profile update
  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !email.trim()) {
      setProfileError('Nome de usuário e e-mail são obrigatórios.');
      return;
    }

    setIsSavingProfile(true);
    setProfileError('');

    try {
      const res = await updateUserProfile({
        username: username.trim(),
        email: email.trim(),
      });

      if (!res.success || !res.user) {
        throw new Error(res.error || 'Erro ao atualizar dados.');
      }

      updateUser(res.user);
      toast.success('Perfil atualizado com sucesso!');
      router.refresh();
    } catch (err: unknown) {
      if (isRedirectError(err)) throw err;
      const errorMsg = (err as Error)?.message || 'Erro ao atualizar o perfil.';
      setProfileError(errorMsg);
      toast.error(errorMsg);
    } finally {
      setIsSavingProfile(false);
    }
  };

  // Handle password change
  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentPassword) {
      setPasswordError('Informe sua senha atual.');
      return;
    }
    if (newPassword.length < 6) {
      setPasswordError('A nova senha deve ter pelo menos 6 caracteres.');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('A nova senha e a confirmação não coincidem.');
      return;
    }

    setIsChangingPassword(true);
    setPasswordError('');

    try {
      const res = await changePassword({
        currentPassword,
        newPassword,
        confirmPassword,
      });

      if (!res.success) {
        throw new Error(res.error || 'Erro ao alterar senha.');
      }

      if (res.user) {
        updateUser(res.user);
      }

      toast.success('Senha alterada com sucesso!');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      router.refresh();
    } catch (err: unknown) {
      if (isRedirectError(err)) throw err;
      const errorMsg = (err as Error)?.message || 'Erro ao alterar senha.';
      setPasswordError(errorMsg);
      toast.error(errorMsg);
    } finally {
      setIsChangingPassword(false);
    }
  };

  const docTypeLabels: Record<DocumentType, { label: string; icon: React.ComponentType<{ className?: string }> }> = {
    NOTICE: { label: 'Ofícios', icon: FileText },
    DECREE: { label: 'Decretos', icon: FileCheck },
    ORDINANCE: { label: 'Portarias', icon: Clock },
    LAW: { label: 'Leis', icon: Scale },
    CONTRACT: { label: 'Contratos', icon: FileSignature },
  };

  const effectiveUser = user || currentUser;

  return (
    <div className="flex flex-col gap-8 max-w-5xl mx-auto transition-colors duration-200">
      {/* Header Profile Hero Card */}
      <div className="p-6 md:p-8 bg-card border border-border backdrop-blur-xl rounded-3xl shadow-xl relative overflow-hidden transition-colors duration-200">
        <div className="absolute top-0 right-0 w-80 h-80 bg-violet-600/10 rounded-full blur-3xl pointer-events-none" />

        <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-6 relative z-10">
          <div className="flex items-center gap-5">
            <div className="w-20 h-20 rounded-2xl bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center text-white font-bold text-2xl uppercase shadow-xl shadow-violet-500/20 shrink-0">
              {effectiveUser.username.slice(0, 2)}
            </div>

            <div className="flex flex-col gap-1.5">
              <div className="flex items-center gap-3 flex-wrap">
                <h1 className="text-2xl md:text-3xl font-extrabold text-foreground tracking-tight">
                  {effectiveUser.username}
                </h1>
                <span
                  className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-bold border ${
                    effectiveUser.role === 'ADMIN'
                      ? 'bg-red-500/10 text-red-500 dark:text-red-400 border-red-500/20'
                      : effectiveUser.role === 'MOD'
                      ? 'bg-violet-500/10 text-violet-600 dark:text-violet-400 border-violet-500/20'
                      : 'bg-muted text-foreground border-border'
                  }`}
                >
                  <Shield className="w-3.5 h-3.5" />
                  {effectiveUser.role === 'ADMIN'
                    ? 'Administrador Global'
                    : effectiveUser.role === 'MOD'
                    ? 'Moderador Municipal'
                    : 'Funcionário'}
                </span>
              </div>

              <div className="flex items-center gap-4 text-xs text-muted-foreground flex-wrap">
                <span className="flex items-center gap-1.5">
                  <Mail className="w-3.5 h-3.5 text-muted-foreground" />
                  {effectiveUser.email}
                </span>

                {effectiveUser.municipality && (
                  <span className="flex items-center gap-1.5">
                    <Building2 className="w-3.5 h-3.5 text-violet-600 dark:text-violet-400" />
                    {effectiveUser.municipality.name} ({effectiveUser.municipality.uf})
                  </span>
                )}
              </div>
            </div>
          </div>

          <div className="flex flex-col items-start md:items-end gap-1 text-xs text-muted-foreground border-t md:border-t-0 pt-4 md:pt-0 border-border w-full md:w-auto">
            <span className="flex items-center gap-1.5">
              <Calendar className="w-3.5 h-3.5 text-muted-foreground" />
              Membro desde: {new Date(effectiveUser.createdAt).toLocaleDateString('pt-BR')}
            </span>
            {effectiveUser.lastLogin && (
              <span className="text-muted-foreground">
                Último acesso: {new Date(effectiveUser.lastLogin).toLocaleDateString('pt-BR')} às{' '}
                {new Date(effectiveUser.lastLogin).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Tabs navigation */}
      <div className="flex items-center gap-2 p-1.5 bg-card/80 border border-border rounded-2xl overflow-x-auto scrollbar-none shadow-md">
        <button
          onClick={() => setActiveTab('profile')}
          className={`flex items-center gap-2.5 px-5 py-3 rounded-xl text-sm font-semibold transition-all cursor-pointer whitespace-nowrap ${
            activeTab === 'profile'
              ? 'bg-violet-600 text-white shadow-md shadow-violet-600/25'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted/60'
          }`}
        >
          <UserIcon className="w-4 h-4" />
          Dados Pessoais
        </button>

        <button
          onClick={() => setActiveTab('security')}
          className={`flex items-center gap-2.5 px-5 py-3 rounded-xl text-sm font-semibold transition-all cursor-pointer whitespace-nowrap ${
            activeTab === 'security'
              ? 'bg-violet-600 text-white shadow-md shadow-violet-600/25'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted/60'
          }`}
        >
          <Lock className="w-4 h-4" />
          Segurança & Senha
        </button>

        <button
          onClick={() => setActiveTab('appearance')}
          className={`flex items-center gap-2.5 px-5 py-3 rounded-xl text-sm font-semibold transition-all cursor-pointer whitespace-nowrap ${
            activeTab === 'appearance'
              ? 'bg-violet-600 text-white shadow-md shadow-violet-600/25'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted/60'
          }`}
        >
          <Sun className="w-4 h-4" />
          Aparência & Temas
        </button>

        {effectiveUser.role === 'COMMON' && (
          <button
            onClick={() => setActiveTab('permissions')}
            className={`flex items-center gap-2.5 px-5 py-3 rounded-xl text-sm font-semibold transition-all cursor-pointer whitespace-nowrap ${
              activeTab === 'permissions'
                ? 'bg-violet-600 text-white shadow-md shadow-violet-600/25'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted/60'
            }`}
          >
            <Shield className="w-4 h-4" />
            Permissões
          </button>
        )}
      </div>

      {/* Tab Contents */}
      <AnimatePresence mode="wait">
        {/* Tab 1: Profile Information */}
        {activeTab === 'profile' && (
          <motion.div
            key="profile"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="p-6 md:p-8 bg-card border border-border backdrop-blur-xl rounded-3xl shadow-xl"
          >
            <div className="flex items-center gap-3 mb-6">
              <div className="w-10 h-10 rounded-xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-600 dark:text-violet-400">
                <UserIcon className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-foreground">Editar Dados do Perfil</h2>
                <p className="text-xs text-muted-foreground">Mantenha suas informações de cadastro sempre atualizadas.</p>
              </div>
            </div>

            {profileError && (
              <div className="bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-xs rounded-xl p-4 mb-6 font-medium">
                {profileError}
              </div>
            )}

            <form onSubmit={handleUpdateProfile} className="flex flex-col gap-5 max-w-xl">
              <Input
                label="Nome de Usuário (Username)"
                placeholder="Ex: seu.usuario"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={isSavingProfile}
                required
              />

              <Input
                label="Endereço de E-mail"
                type="email"
                placeholder="Ex: seu.email@prefeitura.gov.br"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isSavingProfile}
                required
              />

              <div className="flex flex-col gap-2 pt-2">
                <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Município de Atuação</label>
                <div className="p-3.5 bg-muted/60 border border-border rounded-xl flex items-center gap-3 text-sm text-foreground">
                  <Building2 className="w-4 h-4 text-violet-600 dark:text-violet-400 shrink-0" />
                  <span>{effectiveUser.municipality ? `${effectiveUser.municipality.name} (${effectiveUser.municipality.uf})` : 'Acesso Global / Sem vínculo específico'}</span>
                </div>
                <span className="text-[11px] text-muted-foreground">Apenas administradores podem alterar a lotação municipal.</span>
              </div>

              <div className="pt-4 flex justify-start">
                <Button
                  type="submit"
                  variant="default"
                  className="rounded-xl font-bold px-6 h-11 flex items-center gap-2"
                  isLoading={isSavingProfile}
                >
                  <Save className="w-4 h-4" />
                  Salvar Alterações
                </Button>
              </div>
            </form>
          </motion.div>
        )}

        {/* Tab 2: Security & Password Change */}
        {activeTab === 'security' && (
          <motion.div
            key="security"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="p-6 md:p-8 bg-card border border-border backdrop-blur-xl rounded-3xl shadow-xl"
          >
            <div className="flex items-center gap-3 mb-6">
              <div className="w-10 h-10 rounded-xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-600 dark:text-violet-400">
                <KeyRound className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-foreground">Alteração de Senha</h2>
                <p className="text-xs text-muted-foreground">Proteja sua conta utilizando uma senha forte e segura.</p>
              </div>
            </div>

            {passwordError && (
              <div className="bg-red-500/10 border border-red-500/20 text-red-500 dark:text-red-400 text-xs rounded-xl p-4 mb-6 font-medium">
                {passwordError}
              </div>
            )}

            <form onSubmit={handleChangePassword} className="flex flex-col gap-5 max-w-xl">
              <div className="relative">
                <Input
                  label="Senha Atual"
                  type={showCurrentPassword ? 'text' : 'password'}
                  placeholder="Digite sua senha atual"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  disabled={isChangingPassword}
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                  className="absolute right-3.5 top-9 text-muted-foreground hover:text-foreground p-1"
                >
                  {showCurrentPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>

              <div className="relative">
                <Input
                  label="Nova Senha"
                  type={showNewPassword ? 'text' : 'password'}
                  placeholder="Mínimo 6 caracteres"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  disabled={isChangingPassword}
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowNewPassword(!showNewPassword)}
                  className="absolute right-3.5 top-9 text-muted-foreground hover:text-foreground p-1"
                >
                  {showNewPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>

              <Input
                label="Confirmar Nova Senha"
                type="password"
                placeholder="Repita a nova senha"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                disabled={isChangingPassword}
                required
              />

              <div className="pt-4 flex justify-start">
                <Button
                  type="submit"
                  variant="default"
                  className="rounded-xl font-bold px-6 h-11 flex items-center gap-2"
                  isLoading={isChangingPassword}
                >
                  <RefreshCw className="w-4 h-4" />
                  Atualizar Senha
                </Button>
              </div>
            </form>
          </motion.div>
        )}

        {/* Tab 3: Appearance & Themes */}
        {activeTab === 'appearance' && (
          <motion.div
            key="appearance"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="p-6 md:p-8 bg-card border border-border backdrop-blur-xl rounded-3xl shadow-xl"
          >
            <div className="flex items-center gap-3 mb-6">
              <div className="w-10 h-10 rounded-xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-600 dark:text-violet-400">
                <Sparkles className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-foreground">Preferências Visuais & Tema</h2>
                <p className="text-xs text-muted-foreground">Escolha o tema de exibição de sua preferência para a interface.</p>
              </div>
            </div>

            {mounted && (
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 max-w-2xl mt-4">
                {/* Dark Theme Card */}
                <button
                  type="button"
                  onClick={() => {
                    setTheme('dark');
                    toast.success('Tema escuro ativado!');
                  }}
                  className={`p-5 rounded-2xl border text-left flex flex-col justify-between gap-6 transition-all cursor-pointer ${
                    theme === 'dark'
                      ? 'bg-violet-600/10 border-violet-500/50 shadow-md ring-1 ring-violet-500/30'
                      : 'bg-muted/40 border-border hover:border-violet-500/30 hover:bg-muted/70'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="w-10 h-10 rounded-xl bg-muted border border-border flex items-center justify-center text-violet-600 dark:text-violet-400">
                      <Moon className="w-5 h-5" />
                    </div>
                    {theme === 'dark' && <CheckCircle2 className="w-5 h-5 text-violet-600 dark:text-violet-400" />}
                  </div>
                  <div>
                    <span className="font-bold text-foreground block text-sm">Modo Escuro</span>
                    <span className="text-xs text-muted-foreground mt-1 block">Ideal para ambientes com pouca luz e maior conforto visual.</span>
                  </div>
                </button>

                {/* Light Theme Card */}
                <button
                  type="button"
                  onClick={() => {
                    setTheme('light');
                    toast.success('Tema claro ativado!');
                  }}
                  className={`p-5 rounded-2xl border text-left flex flex-col justify-between gap-6 transition-all cursor-pointer ${
                    theme === 'light'
                      ? 'bg-violet-600/10 border-violet-500/50 shadow-md ring-1 ring-violet-500/30'
                      : 'bg-muted/40 border-border hover:border-violet-500/30 hover:bg-muted/70'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="w-10 h-10 rounded-xl bg-muted border border-border flex items-center justify-center text-amber-500">
                      <Sun className="w-5 h-5" />
                    </div>
                    {theme === 'light' && <CheckCircle2 className="w-5 h-5 text-violet-600 dark:text-violet-400" />}
                  </div>
                  <div>
                    <span className="font-bold text-foreground block text-sm">Modo Claro</span>
                    <span className="text-xs text-muted-foreground mt-1 block">Interface limpa com alto contraste e fundo claro.</span>
                  </div>
                </button>

                {/* System Theme Card */}
                <button
                  type="button"
                  onClick={() => {
                    setTheme('system');
                    toast.success('Tema do sistema configurado!');
                  }}
                  className={`p-5 rounded-2xl border text-left flex flex-col justify-between gap-6 transition-all cursor-pointer ${
                    theme === 'system'
                      ? 'bg-violet-600/10 border-violet-500/50 shadow-md ring-1 ring-violet-500/30'
                      : 'bg-muted/40 border-border hover:border-violet-500/30 hover:bg-muted/70'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="w-10 h-10 rounded-xl bg-muted border border-border flex items-center justify-center text-muted-foreground">
                      <Laptop className="w-5 h-5" />
                    </div>
                    {theme === 'system' && <CheckCircle2 className="w-5 h-5 text-violet-600 dark:text-violet-400" />}
                  </div>
                  <div>
                    <span className="font-bold text-foreground block text-sm">Automático</span>
                    <span className="text-xs text-muted-foreground mt-1 block">Sincroniza automaticamente com o tema do sistema operacional.</span>
                  </div>
                </button>
              </div>
            )}
          </motion.div>
        )}

        {/* Tab 4: Permissions (COMMON user only) */}
        {activeTab === 'permissions' && effectiveUser.role === 'COMMON' && (
          <motion.div
            key="permissions"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="p-6 md:p-8 bg-card border border-border backdrop-blur-xl rounded-3xl shadow-xl"
          >
            <div className="flex items-center gap-3 mb-6">
              <div className="w-10 h-10 rounded-xl bg-violet-600/10 border border-violet-500/20 flex items-center justify-center text-violet-600 dark:text-violet-400">
                <Shield className="w-5 h-5" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-foreground">Suas Permissões de Acesso</h2>
                <p className="text-xs text-muted-foreground">
                  Níveis de acesso concedidos para gerenciamento de documentos oficiais do seu município.
                </p>
              </div>
            </div>

            {isLoadingPerms ? (
              <div className="py-12 flex flex-col items-center justify-center gap-3">
                <div className="w-8 h-8 rounded-full border-2 border-violet-500/20 border-t-violet-500 animate-spin" />
                <span className="text-muted-foreground text-xs">Carregando permissões...</span>
              </div>
            ) : (
              <div className="border border-border rounded-2xl overflow-hidden divide-y divide-border bg-card">
                <div className="grid grid-cols-2 p-4 bg-muted/50 text-xs font-bold text-muted-foreground uppercase tracking-wider">
                  <div>Tipo de Documento</div>
                  <div>Nível de Acesso</div>
                </div>

                {(['NOTICE', 'DECREE', 'ORDINANCE', 'LAW', 'CONTRACT'] as DocumentType[]).map((type) => {
                  const perm = permissions.find((p) => p.documentType === type);
                  const level = perm ? perm.level : 'NONE';
                  const meta = docTypeLabels[type];
                  const Icon = meta.icon;

                  return (
                    <div key={type} className="grid grid-cols-2 p-4 items-center gap-4 hover:bg-muted/40 transition-colors">
                      <div className="flex items-center gap-3 font-semibold text-foreground">
                        <Icon className="w-4 h-4 text-violet-600 dark:text-violet-400" />
                        {meta.label}
                      </div>
                      <div>
                        <span
                          className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-semibold border ${
                            level === 'DELETE'
                              ? 'bg-red-500/10 text-red-500 dark:text-red-400 border-red-500/20'
                              : level === 'WRITE'
                              ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20'
                              : level === 'READ'
                              ? 'bg-violet-500/10 text-violet-600 dark:text-violet-400 border-violet-500/20'
                              : 'bg-muted text-muted-foreground border-border'
                          }`}
                        >
                          {level === 'DELETE'
                            ? 'Acesso Total (Criar, Editar, Excluir)'
                            : level === 'WRITE'
                            ? 'Leitura e Escrita (Criar, Editar)'
                            : level === 'READ'
                            ? 'Apenas Leitura (Visualizar)'
                            : 'Sem Acesso'}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
