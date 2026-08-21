'use client';

import React from 'react';
import { DocumentType, ContractType } from '@/types';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, X } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

interface DocumentFilterBarProps {
  activeTab: DocumentType;
  activeTabLabel: string;
  searchValue: string;
  onSearchChange: (value: string) => void;
  yearFilter: string;
  onYearChange: (value: string) => void;
  availableYears: string[];
  contractTypeFilter: ContractType | '';
  onContractTypeChange: (value: ContractType | '') => void;
  onClearFilters: () => void;
}

export const DocumentFilterBar: React.FC<DocumentFilterBarProps> = ({
  activeTab,
  activeTabLabel,
  searchValue,
  onSearchChange,
  yearFilter,
  onYearChange,
  availableYears,
  contractTypeFilter,
  onContractTypeChange,
  onClearFilters,
}) => {
  return (
    <div className="p-5 bg-card border border-border rounded-2xl flex flex-col md:flex-row items-end gap-4 shadow-sm relative overflow-hidden">
      {/* Subtle background glow effect */}
      <div className="absolute -top-24 -left-24 w-48 h-48 bg-teal-600/5 blur-3xl rounded-full pointer-events-none" />

      {/* Search Input */}
      <div className="w-full md:flex-1 relative">
        <Input
          label={`Pesquisa em ${activeTabLabel}`}
          placeholder={`Buscar por descrição em ${activeTabLabel.toLowerCase()}...`}
          value={searchValue}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-10 bg-background text-foreground"
        />
        <Search className="w-4 h-4 text-muted-foreground absolute left-3.5 bottom-3.5 pointer-events-none" />
        {searchValue && (
          <button
            type="button"
            onClick={() => onSearchChange('')}
            className="absolute right-3.5 bottom-3.5 text-muted-foreground hover:text-foreground p-0.5 rounded-full hover:bg-muted transition-all cursor-pointer"
            title="Limpar busca"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Year Selector */}
      <div className="w-full md:w-44 flex flex-col gap-1.5">
        <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">Ano</label>
        <select
          className="w-full bg-background border border-border text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500 cursor-pointer h-10"
          value={yearFilter}
          onChange={(e) => onYearChange(e.target.value)}
        >
          {availableYears.map((y) => (
            <option key={y} value={y} className="bg-card text-foreground">
              {y === 'all' ? 'Todos os Anos' : y}
            </option>
          ))}
        </select>
      </div>

      {/* Contract Type Selector (Visible only when tab is CONTRACT) */}
      <AnimatePresence>
        {activeTab === 'CONTRACT' && (
          <motion.div
            initial={{ opacity: 0, width: 0, scale: 0.95 }}
            animate={{ opacity: 1, width: 'auto', scale: 1 }}
            exit={{ opacity: 0, width: 0, scale: 0.95 }}
            className="w-full md:w-56 flex flex-col gap-1.5"
          >
            <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">
              Tipo de Contrato
            </label>
            <select
              className="w-full bg-background border border-border text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500 cursor-pointer h-10"
              value={contractTypeFilter}
              onChange={(e) => onContractTypeChange(e.target.value as ContractType | '')}
            >
              <option value="" className="bg-card text-foreground">
                Todos os Contratos
              </option>
              <option value="service" className="bg-card text-foreground">
                Prestação de Serviço
              </option>
              <option value="bidding" className="bg-card text-foreground">
                Licitação
              </option>
              <option value="publicinterest" className="bg-card text-foreground">
                Interesse Público
              </option>
            </select>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Clear filters button */}
      <Button
        type="button"
        variant="outline"
        className="w-full md:w-auto h-10 border-border hover:bg-muted text-foreground font-semibold px-5 rounded-xl"
        onClick={onClearFilters}
      >
        Limpar Filtros
      </Button>
    </div>
  );
};
