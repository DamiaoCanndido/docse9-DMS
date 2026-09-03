'use client';

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { Document, ContractType } from '@/types';
import { getExpiringContracts } from '@/app/api/documents';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import {
  Bell,
  Sparkles,
  AlertTriangle,
  Clock,
  CheckCheck,
  Check,
  FileSignature,
  ChevronRight,
  Info,
  Calendar,
  Layers,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/date';
import { APP_VERSION } from '@/lib/version';

export type NotificationCategory = 'EXPIRATION' | 'CHANGELOG';
export type NotificationSeverity = 'critical' | 'warning' | 'info' | 'success';

export interface AppNotification {
  id: string;
  category: NotificationCategory;
  title: string;
  description: string;
  timestamp: string;
  severity: NotificationSeverity;
  tag: string;
  link?: string;
  contractData?: {
    order: number;
    contractType?: ContractType;
    daysRemaining: number;
    endDateFormatted: string;
  };
}

const contractTypeLabels: Record<string, string> = {
  service: 'Prestação de Serviço',
  bidding: 'Licitação',
  publicinterest: 'Interesse Público',
};

// Feed de notas de versão e novidades do sistema ("O que há de novo")
const SYSTEM_CHANGELOG_NOTIFICATIONS: AppNotification[] = [
  {
    id: 'changelog_notifications_and_contracts',
    category: 'CHANGELOG',
    title: 'Central de Notificações & Alertas de Vigência',
    description: 'Novo sino no cabeçalho com avisos automáticos de contratos a vencer e feed de novidades da plataforma em tempo real.',
    timestamp: '2026-09-02T12:00:00Z',
    severity: 'success',
    tag: `Versão ${APP_VERSION}`,
  },
  {
    id: 'changelog_contracts_management',
    category: 'CHANGELOG',
    title: 'Gestão Completa de Contratos Públicos',
    description: 'Agora é possível cadastrar e controlar contratos por tipo (Serviço, Licitação e Interesse Público) com valor, duração e data de início.',
    timestamp: '2026-08-20T10:00:00Z',
    severity: 'info',
    tag: 'Recurso',
  },
  {
    id: 'changelog_system_launch',
    category: 'CHANGELOG',
    title: 'Lançamento do Docseq',
    description: 'Sistema oficial para emissão de Ofícios, Decretos, Portarias, Leis e Contratos com numeração automática sequencial.',
    timestamp: '2026-08-01T08:00:00Z',
    severity: 'info',
    tag: 'Plataforma',
  },
];

interface NotificationCenterProps {
  className?: string;
}

export const NotificationCenter: React.FC<NotificationCenterProps> = ({ className }) => {
  const { user } = useAuth();
  const router = useRouter();

  const [isOpen, setIsOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<'all' | 'expiration' | 'changelog'>('all');
  const [readNotificationIds, setReadNotificationIds] = useState<string[]>([]);
  const [contracts, setContracts] = useState<Document[]>([]);
  const [isLoadingContracts, setIsLoadingContracts] = useState(false);

  // Carrega IDs já lidos do localStorage
  useEffect(() => {
    if (!user) return;
    const timer = setTimeout(() => {
      try {
        const storageKey = `docseq_notifications_read_${user.id}`;
        const saved = localStorage.getItem(storageKey);
        if (saved) {
          setReadNotificationIds(JSON.parse(saved));
        }
      } catch (e) {
        console.error('Erro ao ler notificações do localStorage:', e);
      }
    }, 0);

    return () => clearTimeout(timer);
  }, [user]);

  // Salva IDs lidos no localStorage
  const persistReadIds = useCallback((newIds: string[]) => {
    if (!user) return;
    try {
      const storageKey = `docseq_notifications_read_${user.id}`;
      localStorage.setItem(storageKey, JSON.stringify(newIds));
      setReadNotificationIds(newIds);
    } catch (e) {
      console.error('Erro ao salvar notificações no localStorage:', e);
    }
  }, [user]);

  // Busca contratos ativos para calcular alertas de vencimento (se o usuário pertence a um município)
  useEffect(() => {
    if (!user || user.role === 'ADMIN' || !user.municipalityId) {
      return;
    }

    let isMounted = true;

    const timer = setTimeout(() => {
      if (!isMounted) return;
      setIsLoadingContracts(true);

      getExpiringContracts()
        .then((data) => {
          if (isMounted) {
            setContracts(data || []);
          }
        })
        .catch((err) => {
          console.error('Erro ao buscar contratos para alertas:', err);
        })
        .finally(() => {
          if (isMounted) {
            setIsLoadingContracts(false);
          }
        });
    }, 0);

    return () => {
      isMounted = false;
      clearTimeout(timer);
    };
  }, [user]);

  // Calcula notificações de contratos dinamicamente
  const contractNotifications = useMemo<AppNotification[]>(() => {
    if (!contracts || contracts.length === 0) return [];

    const notifications: AppNotification[] = [];
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    contracts.forEach((doc) => {
      if (doc.type !== 'CONTRACT' || !doc.startIn || !doc.duration) return;

      const startDate = new Date(doc.startIn);
      if (isNaN(startDate.getTime())) return;

      const endDate = new Date(startDate);
      endDate.setMonth(endDate.getMonth() + doc.duration);

      const diffTime = endDate.getTime() - today.getTime();
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

      const typeLabel = doc.contractType ? contractTypeLabels[doc.contractType] || doc.contractType : 'Contrato';
      const endDateStr = formatDate(endDate);

      // Alerta para contratos vencidos ou que vencem em até 90 dias
      if (diffDays < 0) {
        const daysPast = Math.abs(diffDays);
        notifications.push({
          id: `contract_expired_${doc.id}`,
          category: 'EXPIRATION',
          title: `Contrato #${doc.order} Expirado`,
          description: `Vigência do contrato de ${typeLabel} expirou há ${daysPast} dia(s) em ${endDateStr}.`,
          timestamp: doc.updatedAt || doc.createdAt,
          severity: 'critical',
          tag: 'Expirado',
          link: '/?type=CONTRACT',
          contractData: {
            order: doc.order,
            contractType: doc.contractType,
            daysRemaining: diffDays,
            endDateFormatted: endDateStr,
          },
        });
      } else if (diffDays <= 30) {
        notifications.push({
          id: `contract_expiring_30_${doc.id}`,
          category: 'EXPIRATION',
          title: `Contrato #${doc.order} Vencendo em ${diffDays === 0 ? 'hoje' : `${diffDays} dias`}`,
          description: `Faltam ${diffDays} dias para o término em ${endDateStr}. Necessário providenciar aditivo ou encerramento.`,
          timestamp: doc.updatedAt || doc.createdAt,
          severity: 'critical',
          tag: 'Urgente',
          link: '/?type=CONTRACT',
          contractData: {
            order: doc.order,
            contractType: doc.contractType,
            daysRemaining: diffDays,
            endDateFormatted: endDateStr,
          },
        });
      } else if (diffDays <= 60) {
        notifications.push({
          id: `contract_expiring_60_${doc.id}`,
          category: 'EXPIRATION',
          title: `Contrato #${doc.order} Vence em ${diffDays} dias`,
          description: `Término de vigência previsto para ${endDateStr} (${typeLabel}).`,
          timestamp: doc.updatedAt || doc.createdAt,
          severity: 'warning',
          tag: 'Atenção',
          link: '/?type=CONTRACT',
          contractData: {
            order: doc.order,
            contractType: doc.contractType,
            daysRemaining: diffDays,
            endDateFormatted: endDateStr,
          },
        });
      } else if (diffDays <= 90) {
        notifications.push({
          id: `contract_expiring_90_${doc.id}`,
          category: 'EXPIRATION',
          title: `Contrato #${doc.order} em Planejamento (${diffDays}d)`,
          description: `Término de vigência em ${endDateStr}. Planeje renovação ou licitação com antecedência.`,
          timestamp: doc.updatedAt || doc.createdAt,
          severity: 'info',
          tag: 'Planejamento',
          link: '/?type=CONTRACT',
          contractData: {
            order: doc.order,
            contractType: doc.contractType,
            daysRemaining: diffDays,
            endDateFormatted: endDateStr,
          },
        });
      }
    });

    // Ordena os alertas: mais urgentes primeiro
    return notifications.sort((a, b) => {
      const daysA = a.contractData?.daysRemaining ?? 999;
      const daysB = b.contractData?.daysRemaining ?? 999;
      return daysA - daysB;
    });
  }, [contracts]);

  // Lista consolidada de todas as notificações
  const allNotifications = useMemo<AppNotification[]>(() => {
    return [...contractNotifications, ...SYSTEM_CHANGELOG_NOTIFICATIONS];
  }, [contractNotifications]);

  // Contagem de notificações não lidas
  const unreadCount = useMemo(() => {
    return allNotifications.filter((n) => !readNotificationIds.includes(n.id)).length;
  }, [allNotifications, readNotificationIds]);

  const unreadExpirationsCount = useMemo(() => {
    return contractNotifications.filter((n) => !readNotificationIds.includes(n.id)).length;
  }, [contractNotifications, readNotificationIds]);

  const unreadChangelogCount = useMemo(() => {
    return SYSTEM_CHANGELOG_NOTIFICATIONS.filter((n) => !readNotificationIds.includes(n.id)).length;
  }, [readNotificationIds]);

  // Filtragem por aba
  const displayedNotifications = useMemo(() => {
    if (activeTab === 'expiration') return contractNotifications;
    if (activeTab === 'changelog') return SYSTEM_CHANGELOG_NOTIFICATIONS;
    return allNotifications;
  }, [activeTab, contractNotifications, allNotifications]);

  const handleMarkAsRead = (id: string, e?: React.MouseEvent) => {
    if (e) e.stopPropagation();
    if (!readNotificationIds.includes(id)) {
      persistReadIds([...readNotificationIds, id]);
    }
  };

  const handleMarkAllAsRead = () => {
    const allIds = allNotifications.map((n) => n.id);
    persistReadIds(Array.from(new Set([...readNotificationIds, ...allIds])));
  };

  const handleNotificationClick = (item: AppNotification) => {
    handleMarkAsRead(item.id);
    if (item.link) {
      setIsOpen(false);
      router.push(item.link);
    }
  };

  const getSeverityBadge = (severity: NotificationSeverity, tag: string) => {
    switch (severity) {
      case 'critical':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-red-500/15 text-red-600 dark:text-red-400 border border-red-500/30">
            <AlertTriangle className="w-3 h-3" />
            {tag}
          </span>
        );
      case 'warning':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30">
            <Clock className="w-3 h-3" />
            {tag}
          </span>
        );
      case 'success':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30">
            <Sparkles className="w-3 h-3" />
            {tag}
          </span>
        );
      case 'info':
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold bg-teal-500/15 text-teal-600 dark:text-teal-400 border border-teal-500/30">
            <Info className="w-3 h-3" />
            {tag}
          </span>
        );
    }
  };

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger render={
        <button
          type="button"
          className={cn(
            "relative p-2 rounded-xl border border-border bg-background hover:bg-muted text-foreground transition-all cursor-pointer flex items-center justify-center h-9 w-9 shadow-xs focus:outline-none focus:ring-2 focus:ring-teal-500/40",
            isOpen && "bg-muted text-teal-600 dark:text-teal-400 border-teal-500/30",
            className
          )}
          aria-label="Notificações e Avisos de Vencimento"
          title={unreadCount > 0 ? `${unreadCount} notificações não lidas` : 'Central de Notificações'}
        >
          <Bell className={cn("w-4.5 h-4.5 transition-transform", unreadCount > 0 && "text-teal-600 dark:text-teal-400")} />

          {/* Badge indicator */}
          {unreadCount > 0 && (
            <span className="absolute -top-1 -right-1 flex h-4.5 min-w-4.5 px-1 items-center justify-center rounded-full bg-teal-600 text-white text-[10px] font-extrabold shadow-sm shadow-teal-600/30 animate-in zoom-in duration-200">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
          )}
        </button>
      } />

      <PopoverContent
        align="end"
        sideOffset={8}
        className="w-88 sm:w-96 p-0 bg-card border border-border rounded-2xl shadow-2xl overflow-hidden z-50 text-foreground"
      >
        {/* Header */}
        <div className="p-4 border-b border-border bg-muted/40 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-teal-600/10 border border-teal-500/20 flex items-center justify-center text-teal-600 dark:text-teal-400">
              <Bell className="w-4 h-4" />
            </div>
            <div>
              <h3 className="text-sm font-bold text-foreground flex items-center gap-1.5 leading-tight">
                Notificações
                {unreadCount > 0 && (
                  <span className="text-[10px] font-extrabold bg-teal-600/15 text-teal-600 dark:text-teal-400 px-2 py-0.5 rounded-full border border-teal-500/30">
                    {unreadCount} nova{unreadCount > 1 ? 's' : ''}
                  </span>
                )}
              </h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">
                Vigência de contratos e novidades do sistema
              </p>
            </div>
          </div>

          {unreadCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleMarkAllAsRead}
              className="h-7 px-2 text-[11px] text-muted-foreground hover:text-foreground hover:bg-muted flex items-center gap-1 cursor-pointer"
              title="Marcar todas como lidas"
            >
              <CheckCheck className="w-3.5 h-3.5" />
              <span>Ler todas</span>
            </Button>
          )}
        </div>

        {/* Tabs Bar */}
        <div className="flex items-center border-b border-border bg-card px-2 pt-1 gap-1 text-xs">
          <button
            type="button"
            onClick={() => setActiveTab('all')}
            className={cn(
              "px-3 py-2 font-semibold transition-all border-b-2 cursor-pointer flex items-center gap-1.5",
              activeTab === 'all'
                ? "border-teal-600 text-teal-600 dark:text-teal-400"
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            <Layers className="w-3.5 h-3.5" />
            Todas
            {unreadCount > 0 && (
              <span className="ml-0.5 px-1.5 py-0.2 rounded-full bg-muted text-[10px] font-bold text-foreground">
                {unreadCount}
              </span>
            )}
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('expiration')}
            className={cn(
              "px-3 py-2 font-semibold transition-all border-b-2 cursor-pointer flex items-center gap-1.5",
              activeTab === 'expiration'
                ? "border-teal-600 text-teal-600 dark:text-teal-400"
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            <Clock className="w-3.5 h-3.5" />
            Vencimentos
            {unreadExpirationsCount > 0 && (
              <span className="ml-0.5 px-1.5 py-0.2 rounded-full bg-red-500/15 text-red-600 dark:text-red-400 text-[10px] font-bold border border-red-500/20">
                {unreadExpirationsCount}
              </span>
            )}
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('changelog')}
            className={cn(
              "px-3 py-2 font-semibold transition-all border-b-2 cursor-pointer flex items-center gap-1.5",
              activeTab === 'changelog'
                ? "border-teal-600 text-teal-600 dark:text-teal-400"
                : "border-transparent text-muted-foreground hover:text-foreground"
            )}
          >
            <Sparkles className="w-3.5 h-3.5" />
            Novidades ✨
            {unreadChangelogCount > 0 && (
              <span className="ml-0.5 px-1.5 py-0.2 rounded-full bg-teal-500/15 text-teal-600 dark:text-teal-400 text-[10px] font-bold border border-teal-500/20">
                {unreadChangelogCount}
              </span>
            )}
          </button>
        </div>

        {/* Notifications List Container */}
        <div className="max-h-84 overflow-y-auto divide-y divide-border scrollbar-thin">
          {isLoadingContracts && activeTab !== 'changelog' && (
            <div className="py-6 text-center text-xs text-muted-foreground flex flex-col items-center justify-center gap-2">
              <div className="w-5 h-5 rounded-full border-2 border-teal-500/20 border-t-teal-500 animate-spin" />
              <span>Verificando prazos e vigências...</span>
            </div>
          )}

          {displayedNotifications.length === 0 ? (
            <div className="py-12 px-4 text-center flex flex-col items-center justify-center text-muted-foreground">
              <div className="w-12 h-12 rounded-2xl bg-muted/60 flex items-center justify-center mb-3 border border-border">
                {activeTab === 'expiration' ? (
                  <Check className="w-6 h-6 text-emerald-500" />
                ) : (
                  <Sparkles className="w-6 h-6 text-teal-600 dark:text-teal-400" />
                )}
              </div>
              <p className="text-xs font-semibold text-foreground">Tudo em dia por aqui!</p>
              <p className="text-[11px] text-muted-foreground mt-1 max-w-xs">
                {activeTab === 'expiration'
                  ? 'Nenhum contrato com vencimento próximo nos próximos 90 dias.'
                  : 'Nenhuma notificação pendente no momento.'}
              </p>
            </div>
          ) : (
            displayedNotifications.map((item) => {
              const isRead = readNotificationIds.includes(item.id);
              const isContract = item.category === 'EXPIRATION';

              return (
                <div
                  key={item.id}
                  onClick={() => handleNotificationClick(item)}
                  className={cn(
                    "p-3.5 transition-colors cursor-pointer group flex items-start gap-3 relative hover:bg-muted/50",
                    !isRead ? "bg-teal-600/[0.03] dark:bg-teal-500/[0.04]" : "bg-card opacity-90"
                  )}
                >
                  {/* Unread indicator dot */}
                  {!isRead && (
                    <span className="w-2 h-2 rounded-full bg-teal-500 shrink-0 mt-1.5 ring-2 ring-teal-500/20" />
                  )}

                  {/* Icon */}
                  <div
                    className={cn(
                      "w-8 h-8 rounded-xl shrink-0 flex items-center justify-center text-xs font-bold border",
                      isContract
                        ? item.severity === 'critical'
                          ? "bg-red-500/10 text-red-500 border-red-500/20"
                          : item.severity === 'warning'
                          ? "bg-amber-500/10 text-amber-500 border-amber-500/20"
                          : "bg-teal-500/10 text-teal-600 dark:text-teal-400 border-teal-500/20"
                        : "bg-teal-500/10 text-teal-600 dark:text-teal-400 border-teal-500/20"
                    )}
                  >
                    {isContract ? (
                      <FileSignature className="w-4 h-4" />
                    ) : (
                      <Sparkles className="w-4 h-4" />
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-1 mb-1">
                      <h4 className={cn(
                        "text-xs font-semibold truncate",
                        !isRead ? "text-foreground font-bold" : "text-foreground/90"
                      )}>
                        {item.title}
                      </h4>
                      {getSeverityBadge(item.severity, item.tag)}
                    </div>

                    <p className="text-[11px] text-muted-foreground leading-relaxed">
                      {item.description}
                    </p>

                    <div className="mt-2 flex items-center justify-between text-[10px] text-muted-foreground/80">
                      <span className="flex items-center gap-1">
                        <Calendar className="w-3 h-3 text-muted-foreground" />
                        {isContract && item.contractData
                          ? `Fim: ${item.contractData.endDateFormatted}`
                          : formatDate(item.timestamp)}
                      </span>

                      {isContract && (
                        <span className="flex items-center gap-0.5 text-teal-600 dark:text-teal-400 font-semibold group-hover:underline">
                          Ver Contrato <ChevronRight className="w-3 h-3" />
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="p-2.5 bg-muted/40 border-t border-border flex items-center justify-between text-[10px] text-muted-foreground px-4">
          <span className="font-medium flex items-center gap-1">
            <Sparkles className="w-3 h-3 text-teal-600 dark:text-teal-400" />
            Docseq <strong className="text-foreground">{APP_VERSION}</strong>
          </span>
          {user?.municipality && (
            <span className="truncate max-w-[160px] text-muted-foreground">
              {user.municipality.name} ({user.municipality.uf})
            </span>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
};
