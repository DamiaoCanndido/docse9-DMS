'use client';

import React from 'react';
import { Button } from '@/components/ui/button';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { cn } from '@/lib/utils';

interface PaginationBarProps {
  currentPage: number;
  pageSize: number;
  total: number;
  itemLabel?: string;
  isPending?: boolean;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  className?: string;
}

export const PaginationBar: React.FC<PaginationBarProps> = ({
  currentPage,
  pageSize,
  total,
  itemLabel = 'registros',
  isPending = false,
  onPageChange,
  onPageSizeChange,
  className,
}) => {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const startItem = total === 0 ? 0 : (currentPage - 1) * pageSize + 1;
  const endItem = Math.min(total, currentPage * pageSize);

  const canGoPrevious = currentPage > 1 && !isPending;
  const canGoNext = currentPage < totalPages && !isPending;

  return (
    <nav
      aria-label="Navegação de páginas"
      className={cn(
        'px-4 sm:px-6 py-4 border-t border-border bg-card/60 flex flex-col sm:flex-row items-center justify-between gap-4 w-full transition-colors duration-200',
        className
      )}
    >
      {/* Informações e Seletor de Tamanho de Página */}
      <div className="flex flex-wrap items-center justify-center sm:justify-start gap-3 w-full sm:w-auto">
        <span className="text-xs text-muted-foreground font-medium">
          Mostrando <strong className="text-foreground font-bold">{startItem}</strong>-
          <strong className="text-foreground font-bold">{endItem}</strong> de{' '}
          <strong className="text-foreground font-bold">{total}</strong> {itemLabel}
        </span>

        {onPageSizeChange && (
          <div className="flex items-center gap-1.5 ml-0 sm:ml-2">
            <label htmlFor="pageSizeSelect" className="text-xs text-muted-foreground uppercase font-semibold">
              Exibir:
            </label>
            <select
              id="pageSizeSelect"
              value={pageSize}
              disabled={isPending}
              onChange={(e) => onPageSizeChange(Number(e.target.value))}
              aria-label="Selecione a quantidade de itens por página"
              className="bg-muted/70 border border-border text-foreground text-xs rounded-xl px-2.5 py-1.5 h-9 focus:outline-none focus:border-violet-500 cursor-pointer disabled:opacity-50"
            >
              <option value={10}>10 por pág</option>
              <option value={25}>25 por pág</option>
              <option value={50}>50 por pág</option>
            </select>
          </div>
        )}
      </div>

      {/* Controles de Navegação com Touch Target Aumentado (Touch-friendly >= 44px) */}
      <div className="flex items-center gap-1.5 sm:gap-2">
        {/* Primeira página */}
        <Button
          variant="outline"
          size="icon"
          aria-label="Ir para a primeira página"
          title="Primeira página"
          className="h-11 w-11 sm:h-9 sm:w-9 border-border text-muted-foreground hover:text-foreground hover:bg-muted rounded-xl min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0 flex items-center justify-center"
          onClick={() => onPageChange(1)}
          disabled={!canGoPrevious}
        >
          <ChevronsLeft className="w-4 h-4" />
        </Button>

        {/* Página anterior */}
        <Button
          variant="outline"
          aria-label="Ir para a página anterior"
          className="h-11 px-3.5 sm:h-9 sm:px-3 border-border text-muted-foreground hover:text-foreground hover:bg-muted rounded-xl min-h-[44px] text-xs font-semibold flex items-center justify-center"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={!canGoPrevious}
        >
          <ChevronLeft className="w-4 h-4 mr-1 sm:mr-0.5" />
          <span className="inline sm:inline">Anterior</span>
        </Button>

        {/* Indicador de Página */}
        <span className="text-xs text-foreground font-bold px-3 py-2 rounded-xl bg-muted/80 border border-border min-h-[44px] sm:min-h-0 flex items-center justify-center">
          {currentPage} / {totalPages}
        </span>

        {/* Próxima página */}
        <Button
          variant="outline"
          aria-label="Ir para a próxima página"
          className="h-11 px-3.5 sm:h-9 sm:px-3 border-border text-muted-foreground hover:text-foreground hover:bg-muted rounded-xl min-h-[44px] text-xs font-semibold flex items-center justify-center"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={!canGoNext}
        >
          <span className="inline sm:inline">Próxima</span>
          <ChevronRight className="w-4 h-4 ml-1 sm:ml-0.5" />
        </Button>

        {/* Última página */}
        <Button
          variant="outline"
          size="icon"
          aria-label="Ir para a última página"
          title="Última página"
          className="h-11 w-11 sm:h-9 sm:w-9 border-border text-muted-foreground hover:text-foreground hover:bg-muted rounded-xl min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0 flex items-center justify-center"
          onClick={() => onPageChange(totalPages)}
          disabled={!canGoNext}
        >
          <ChevronsRight className="w-4 h-4" />
        </Button>
      </div>
    </nav>
  );
};
