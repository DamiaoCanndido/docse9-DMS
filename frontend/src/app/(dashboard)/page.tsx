'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../../context/AuthContext';
import { documentService } from '../../services/documents';
import { Document, DocumentType, ContractType } from '../../types';
import { Button } from '../../components/ui/Button';
import { Input } from '../../components/ui/Input';

export default function DashboardPage() {
  const { user } = useAuth();
  const [documents, setDocuments] = useState<Document[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [isLoading, setIsLoading] = useState(false);

  // Filters state
  const [searchTerm, setSearchTerm] = useState('');
  const [docTypeFilter, setDocTypeFilter] = useState<DocumentType | ''>('');
  const [contractTypeFilter, setContractTypeFilter] = useState<ContractType | ''>('');

  const fetchDocuments = useCallback(async () => {
    setIsLoading(true);
    try {
      const filter = {
        type: docTypeFilter ? docTypeFilter : undefined,
        contractType: contractTypeFilter ? contractTypeFilter : undefined,
        search: searchTerm ? searchTerm : undefined,
      };
      const response = await documentService.getAll(filter, page, pageSize);
      setDocuments(response.data || []);
      setTotal(response.total || 0);
    } catch (error) {
      console.error('Falha ao buscar documentos:', error);
    } finally {
      setIsLoading(false);
    }
  }, [docTypeFilter, contractTypeFilter, searchTerm, page, pageSize]);

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  const handleClearFilters = () => {
    setSearchTerm('');
    setDocTypeFilter('');
    setContractTypeFilter('');
    setPage(1);
  };

  return (
    <div className="flex flex-col gap-8">
      {/* Welcome banner */}
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white">Documentos</h1>
          <p className="text-zinc-400 mt-1">
            Gerencie e pesquise os documentos oficiais de {user?.municipality?.name || 'sua prefeitura'}.
          </p>
        </div>
      </div>

      {/* Filter panel */}
      <div className="p-5 bg-zinc-950 border border-border rounded-xl flex flex-col md:flex-row items-end gap-4 shadow-xl">
        <div className="w-full md:flex-1">
          <Input
            label="Pesquisa rápida"
            placeholder="Buscar por descrição..."
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setPage(1);
            }}
          />
        </div>

        {/* Document Type Dropdown */}
        <div className="w-full md:w-48 flex flex-col gap-1.5">
          <label className="text-sm font-medium text-zinc-400">Tipo de Documento</label>
          <select
            className="w-full bg-zinc-900 border border-border text-foreground px-3 py-2.5 rounded-lg text-sm transition-all focus:outline-none focus:border-primary"
            value={docTypeFilter}
            onChange={(e) => {
              setDocTypeFilter(e.target.value as DocumentType);
              setPage(1);
            }}
          >
            <option value="">Todos</option>
            <option value="NOTICE">Ofício</option>
            <option value="DECREE">Decreto</option>
            <option value="ORDINANCE">Portaria</option>
            <option value="LAW">Lei</option>
            <option value="CONTRACT">Contrato</option>
          </select>
        </div>

        {/* Contract Type Dropdown (visible only when type is CONTRACT) */}
        {docTypeFilter === 'CONTRACT' && (
          <div className="w-full md:w-48 flex flex-col gap-1.5">
            <label className="text-sm font-medium text-zinc-400">Tipo de Contrato</label>
            <select
              className="w-full bg-zinc-900 border border-border text-foreground px-3 py-2.5 rounded-lg text-sm transition-all focus:outline-none focus:border-primary"
              value={contractTypeFilter}
              onChange={(e) => {
                setContractTypeFilter(e.target.value as ContractType);
                setPage(1);
              }}
            >
              <option value="">Todos</option>
              <option value="service">Prestação de Serviço</option>
              <option value="bidding">Licitação</option>
              <option value="publicinterest">Interesse Público</option>
            </select>
          </div>
        )}

        <Button variant="outline" className="w-full md:w-auto h-11" onClick={handleClearFilters}>
          Limpar Filtros
        </Button>
      </div>

      {/* Documents list */}
      <div className="bg-zinc-950 border border-border rounded-xl overflow-hidden shadow-xl">
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 gap-3">
            <div className="w-8 h-8 rounded-full border-2 border-primary/20 border-t-primary animate-spin" />
            <span className="text-zinc-400 text-sm font-medium">Carregando documentos...</span>
          </div>
        ) : documents.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 px-4 text-center">
            <div className="w-12 h-12 rounded-xl bg-zinc-900 flex items-center justify-center text-zinc-500 mb-4 border border-border">
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-white">Nenhum documento encontrado</h3>
            <p className="text-zinc-500 text-sm mt-1 max-w-sm">
              Tente redefinir seus termos de busca ou filtros para listar mais registros.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="border-b border-border bg-zinc-900/50 text-zinc-400 text-xs font-semibold uppercase tracking-wider">
                  <th className="px-6 py-4">Número</th>
                  <th className="px-6 py-4">Tipo</th>
                  <th className="px-6 py-4">Descrição</th>
                  <th className="px-6 py-4">Detalhes Tipo</th>
                  <th className="px-6 py-4 text-right">Data de Registro</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border text-sm text-zinc-300">
                {documents.map((doc) => (
                  <tr key={doc.id} className="hover:bg-zinc-900/30 transition-colors">
                    <td className="px-6 py-4 font-semibold text-white">
                      #{doc.order}
                    </td>
                    <td className="px-6 py-4">
                      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                        doc.type === 'CONTRACT' ? 'bg-violet-500/10 text-violet-400 border border-violet-500/20' : 'bg-zinc-800 text-zinc-400 border border-zinc-700/50'
                      }`}>
                        {doc.type}
                      </span>
                    </td>
                    <td className="px-6 py-4 truncate max-w-xs md:max-w-md">
                      {doc.description}
                    </td>
                    <td className="px-6 py-4">
                      {doc.type === 'CONTRACT' && doc.contractType ? (
                        <span className="text-zinc-400 italic text-xs">
                          Contrato: {doc.contractType}
                        </span>
                      ) : (
                        <span className="text-zinc-600">—</span>
                      )}
                    </td>
                    <td className="px-6 py-4 text-right text-zinc-500 text-xs">
                      {new Date(doc.createdAt).toLocaleDateString('pt-BR')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination bar */}
        {total > pageSize && (
          <div className="px-6 py-4 border-t border-border flex items-center justify-between gap-4">
            <span className="text-xs text-zinc-500 font-medium">
              Mostrando {documents.length} de {total} documentos
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                className="px-3 py-1.5 text-xs h-9"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                Anterior
              </Button>
              <span className="text-xs text-zinc-400 font-semibold px-2">
                Página {page}
              </span>
              <Button
                variant="outline"
                className="px-3 py-1.5 text-xs h-9"
                onClick={() => setPage(p => p + 1)}
                disabled={page * pageSize >= total}
              >
                Próxima
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
