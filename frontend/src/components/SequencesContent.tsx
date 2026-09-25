'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { User, SequenceItemResponse, DocumentType, ContractType } from '@/types';
import { getMunicipalitySequences, setMunicipalitySequenceOffset } from '@/app/api/documents';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { 
  ListOrdered, 
  Building2, 
  Calendar as CalendarIcon, 
  CheckCircle2, 
  AlertCircle, 
  Loader2, 
  RefreshCw, 
  Save, 
  FileText, 
  Scale, 
  ScrollText, 
  Briefcase 
} from 'lucide-react';
import { toast } from 'sonner';

interface SequencesContentProps {
  currentUser: User;
  initialData?: SequenceItemResponse[];
}

const currentYear = new Date().getFullYear();
const availableYears = [
  (currentYear - 2).toString(),
  (currentYear - 1).toString(),
  currentYear.toString(),
  (currentYear + 1).toString(),
];

export const SequencesContent: React.FC<SequencesContentProps> = ({
  currentUser,
  initialData = [],
}) => {
  const [selectedYear, setSelectedYear] = useState<number>(currentYear);
  const [sequences, setSequences] = useState<SequenceItemResponse[]>(initialData);
  const [initialOrders, setInitialOrders] = useState<Record<string, string>>({});
  const [savingKey, setSavingKey] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const getSequenceKey = (type: DocumentType, contractType?: ContractType | null) => {
    return `${type}:${contractType || ''}`;
  };

  const syncInitialOrders = useCallback((items: SequenceItemResponse[]) => {
    const orders: Record<string, string> = {};
    items.forEach((item) => {
      const key = getSequenceKey(item.type, item.contractType);
      orders[key] = item.initialOrder.toString();
    });
    setInitialOrders(orders);
  }, []);

  const fetchSequences = useCallback(async (year: number) => {
    setIsLoading(true);
    try {
      const data = await getMunicipalitySequences(currentUser.municipalityId, year);
      setSequences(data);
      syncInitialOrders(data);
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Erro ao carregar sequências.';
      toast.error(msg);
    } finally {
      setIsLoading(false);
    }
  }, [currentUser.municipalityId, syncInitialOrders]);

  useEffect(() => {
    if (initialData.length > 0) {
      syncInitialOrders(initialData);
    } else {
      fetchSequences(selectedYear);
    }
  }, [fetchSequences, initialData, selectedYear, syncInitialOrders]);

  const handleYearChange = (newYearStr: string) => {
    const yr = Number(newYearStr);
    setSelectedYear(yr);
    fetchSequences(yr);
  };

  const handleOrderChange = (key: string, val: string) => {
    setInitialOrders((prev) => ({ ...prev, [key]: val }));
  };

  const handleSaveOffset = async (item: SequenceItemResponse) => {
    const key = getSequenceKey(item.type, item.contractType);
    const orderStr = initialOrders[key];
    const orderNum = Number(orderStr);

    if (!orderStr || isNaN(orderNum) || orderNum < 1 || !Number.isInteger(orderNum)) {
      toast.error('O marco inicial deve ser um número inteiro maior ou igual a 1.');
      return;
    }

    setSavingKey(key);
    try {
      await setMunicipalitySequenceOffset(currentUser.municipalityId, {
        type: item.type,
        contractType: item.contractType || undefined,
        year: item.type === 'LAW' ? undefined : selectedYear,
        initialOrder: orderNum,
      });

      toast.success(`Marco inicial de ${getTypeLabel(item.type, item.contractType)} atualizado para nº ${orderNum}!`);
      // Atualiza a listagem para recalcular nextOrder
      await fetchSequences(selectedYear);
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Erro ao atualizar marco inicial.';
      toast.error(msg);
    } finally {
      setSavingKey(null);
    }
  };

  const getTypeLabel = (type: DocumentType, contractType?: ContractType | null) => {
    switch (type) {
      case 'NOTICE':
        return 'Ofício';
      case 'DECREE':
        return 'Decreto';
      case 'ORDINANCE':
        return 'Portaria';
      case 'LAW':
        return 'Lei Municipal';
      case 'CONTRACT':
        if (contractType === 'service') return 'Contrato (Prestação de Serviço)';
        if (contractType === 'bidding') return 'Contrato (Licitação)';
        if (contractType === 'publicinterest') return 'Contrato (Interesse Público)';
        return 'Contrato';
      default:
        return type;
    }
  };

  const getTypeIcon = (type: DocumentType) => {
    switch (type) {
      case 'NOTICE':
        return <FileText className="w-4 h-4 text-teal-600 dark:text-teal-400" />;
      case 'DECREE':
        return <ScrollText className="w-4 h-4 text-sky-600 dark:text-sky-400" />;
      case 'ORDINANCE':
        return <Briefcase className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />;
      case 'LAW':
        return <Scale className="w-4 h-4 text-amber-600 dark:text-amber-400" />;
      case 'CONTRACT':
        return <ListOrdered className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />;
      default:
        return <FileText className="w-4 h-4" />;
    }
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Header com Informações do Município e Seletor de Ano */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 p-6 bg-card border border-border rounded-2xl shadow-xs">
        <div className="flex items-start gap-4">
          <div className="w-12 h-12 rounded-xl bg-teal-500/10 border border-teal-500/20 text-teal-600 dark:text-teal-400 flex items-center justify-center shrink-0 mt-0.5">
            <ListOrdered className="w-6 h-6" />
          </div>
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">Sequências & Marcos Iniciais</h1>
              {currentUser.municipality && (
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-muted border border-border text-foreground">
                  <Building2 className="w-3.5 h-3.5 text-teal-600 dark:text-teal-400" />
                  {currentUser.municipality.name} ({currentUser.municipality.uf})
                </span>
              )}
            </div>
            <p className="text-xs text-muted-foreground leading-relaxed max-w-2xl">
              Configure o ponto de partida numérico para municípios que aderiram no meio do exercício ou com acervos físicos prévios.
              A geração automática continuará a partir do maior número entre o marco configurado e os atos já emitidos.
            </p>
          </div>
        </div>

        {/* Controles de Ano e Recarregar */}
        <div className="flex items-center gap-3 self-end md:self-center">
          <div className="flex flex-col gap-1">
            <label className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground">
              Exercício (Ano)
            </label>
            <div className="flex items-center gap-2">
              <select
                className="bg-background border border-border text-foreground px-3 py-1.5 rounded-xl text-xs font-semibold focus:outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500 h-9 cursor-pointer"
                value={selectedYear.toString()}
                onChange={(e) => handleYearChange(e.target.value)}
              >
                {availableYears.map((yr) => (
                  <option key={yr} value={yr} className="bg-card text-foreground">
                    Ano {yr}
                  </option>
                ))}
              </select>

              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => fetchSequences(selectedYear)}
                disabled={isLoading}
                className="h-9 px-3 border-border hover:bg-muted text-foreground rounded-xl"
                title="Atualizar lista"
              >
                <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin text-teal-600' : ''}`} />
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Tabela de Séries e Marcos Iniciais */}
      <div className="bg-card border border-border rounded-2xl shadow-xs overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-border bg-muted/30 text-[11px] font-bold text-muted-foreground uppercase tracking-wider">
                <th className="py-3.5 px-4 sm:px-6">Tipo de Documento</th>
                <th className="py-3.5 px-4 text-center">Exercício / Regime</th>
                <th className="py-3.5 px-4 text-center">Último Nº Emitido</th>
                <th className="py-3.5 px-4 text-center">Marco Inicial (Offset)</th>
                <th className="py-3.5 px-4 text-center">Próximo Nº a Ser Gerado</th>
                <th className="py-3.5 px-4 sm:px-6 text-right">Ação</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border text-xs">
              {sequences.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-muted-foreground">
                    {isLoading ? (
                      <div className="flex items-center justify-center gap-2">
                        <Loader2 className="w-4 h-4 animate-spin text-teal-600" />
                        <span>Carregando sequências numéricas...</span>
                      </div>
                    ) : (
                      'Nenhuma série documental encontrada para este município.'
                    )}
                  </td>
                </tr>
              ) : (
                sequences.map((item) => {
                  const key = getSequenceKey(item.type, item.contractType);
                  const isSavingThis = savingKey === key;
                  const currentVal = initialOrders[key] !== undefined ? initialOrders[key] : item.initialOrder.toString();
                  const isLaw = item.type === 'LAW';

                  return (
                    <tr
                      key={key}
                      className="hover:bg-muted/20 transition-colors duration-150"
                    >
                      {/* Tipo de Documento */}
                      <td className="py-3.5 px-4 sm:px-6">
                        <div className="flex items-center gap-3">
                          <div className="p-2 rounded-lg bg-muted border border-border shrink-0">
                            {getTypeIcon(item.type)}
                          </div>
                          <div className="flex flex-col">
                            <span className="font-bold text-foreground">
                              {getTypeLabel(item.type, item.contractType)}
                            </span>
                            <span className="text-[10px] text-muted-foreground uppercase tracking-wider">
                              Código: {item.type}
                            </span>
                          </div>
                        </div>
                      </td>

                      {/* Exercício / Regime */}
                      <td className="py-3.5 px-4 text-center">
                        {isLaw ? (
                          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20">
                            Perpétuo (Contínuo)
                          </span>
                        ) : (
                          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-teal-500/10 text-teal-600 dark:text-teal-400 border border-teal-500/20">
                            Anual ({selectedYear})
                          </span>
                        )}
                      </td>

                      {/* Último Nº Emitido */}
                      <td className="py-3.5 px-4 text-center font-mono">
                        {item.currentOrder > 0 ? (
                          <span className="font-semibold text-foreground text-sm">
                            {item.currentOrder}
                          </span>
                        ) : (
                          <span className="text-muted-foreground italic text-[11px]">
                            Nenhum
                          </span>
                        )}
                      </td>

                      {/* Marco Inicial Configurado */}
                      <td className="py-3.5 px-4 text-center">
                        <div className="inline-flex items-center justify-center">
                          <Input
                            type="number"
                            min={1}
                            step={1}
                            value={currentVal}
                            onChange={(e) => handleOrderChange(key, e.target.value)}
                            className="w-24 text-center h-8 text-xs font-mono font-bold bg-background text-foreground"
                            aria-label={`Marco inicial para ${getTypeLabel(item.type, item.contractType)}`}
                          />
                        </div>
                      </td>

                      {/* Próximo Número a Ser Gerado */}
                      <td className="py-3.5 px-4 text-center">
                        <span className="inline-flex items-center justify-center min-w-10 px-2.5 py-1 rounded-lg text-xs font-mono font-bold bg-teal-500/15 text-teal-700 dark:text-teal-300 border border-teal-500/30">
                          {item.nextOrder}
                        </span>
                      </td>

                      {/* Ações */}
                      <td className="py-3.5 px-4 sm:px-6 text-right">
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => handleSaveOffset(item)}
                          disabled={isSavingThis || isLoading}
                          className="h-8 px-3 rounded-xl border-border text-foreground hover:bg-teal-500/10 hover:text-teal-600 dark:hover:text-teal-400 hover:border-teal-500/30 font-semibold text-xs gap-1.5 transition-all"
                        >
                          {isSavingThis ? (
                            <>
                              <Loader2 className="w-3.5 h-3.5 animate-spin" />
                              <span>Salvando...</span>
                            </>
                          ) : (
                            <>
                              <Save className="w-3.5 h-3.5" />
                              <span>Salvar</span>
                            </>
                          )}
                        </Button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
