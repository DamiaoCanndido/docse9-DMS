'use client';

import React from 'react';
import { Municipality } from '@/types';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, X } from 'lucide-react';

interface UserFilterBarProps {
  searchValue: string;
  onSearchChange: (value: string) => void;
  roleFilter: string;
  onRoleChange: (value: string) => void;
  municipalityFilter: string;
  onMunicipalityChange: (value: string) => void;
  municipalities: Municipality[];
  isAdmin: boolean;
  onClearFilters: () => void;
}

export const UserFilterBar: React.FC<UserFilterBarProps> = ({
  searchValue,
  onSearchChange,
  roleFilter,
  onRoleChange,
  municipalityFilter,
  onMunicipalityChange,
  municipalities,
  isAdmin,
  onClearFilters,
}) => {
  return (
    <div className="p-6 bg-card border border-border backdrop-blur-md rounded-2xl flex flex-col md:flex-row items-end gap-5 shadow-md relative overflow-hidden">
      <div className="absolute -top-24 -left-24 w-48 h-48 bg-teal-600/5 blur-3xl rounded-full pointer-events-none" />

      {/* Text Search */}
      <div className="w-full md:flex-1 relative">
        <Input
          label="Buscar por nome ou e-mail"
          placeholder="Digite o nome de usuário ou e-mail..."
          value={searchValue}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-10"
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

      {/* Role Filter */}
      <div className="w-full md:w-48 flex flex-col gap-1.5">
        <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">
          Perfil (Role)
        </label>
        <select
          className="w-full bg-background border border-border text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500 cursor-pointer h-10"
          value={roleFilter}
          onChange={(e) => onRoleChange(e.target.value)}
        >
          <option value="" className="bg-card text-foreground">
            Todos
          </option>
          {isAdmin && (
            <option value="ADMIN" className="bg-card text-foreground">
              Administrador
            </option>
          )}
          <option value="MOD" className="bg-card text-foreground">
            Moderador
          </option>
          <option value="COMMON" className="bg-card text-foreground">
            Funcionário Comum
          </option>
        </select>
      </div>

      {/* Municipality Filter (ADMIN only) */}
      {isAdmin && (
        <div className="w-full md:w-56 flex flex-col gap-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-foreground/80">
            Município
          </label>
          <select
            className="w-full bg-background border border-border text-foreground px-3.5 py-2.5 rounded-xl text-sm transition-all focus:outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500 cursor-pointer h-10"
            value={municipalityFilter}
            onChange={(e) => onMunicipalityChange(e.target.value)}
          >
            <option value="" className="bg-card text-foreground">
              Todos
            </option>
            {municipalities.map((m) => (
              <option key={m.id} value={m.id} className="bg-card text-foreground">
                {m.name} ({m.uf})
              </option>
            ))}
          </select>
        </div>
      )}

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
