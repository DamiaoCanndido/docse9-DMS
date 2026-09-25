import { describe, it, expect } from 'vitest';
import { parseDateSafe, formatDate, formatDateTime, formatTime, combineDateAndTime } from '../date';

describe('Date Utilities (lib/date.ts)', () => {
  describe('parseDateSafe', () => {
    it('deve retornar undefined para valores vazios ou inválidos', () => {
      expect(parseDateSafe('')).toBeUndefined();
      expect(parseDateSafe('   ')).toBeUndefined();
      expect(parseDateSafe(null)).toBeUndefined();
      expect(parseDateSafe(undefined)).toBeUndefined();
      expect(parseDateSafe('data-invalida')).toBeUndefined();
    });

    it('deve fazer parse de data histórica antiga no formato brasileiro DD/MM/AAAA (ex: 15/03/1980)', () => {
      const parsed = parseDateSafe('15/03/1980');
      expect(parsed).toBeInstanceOf(Date);
      expect(parsed?.getFullYear()).toBe(1980);
      expect(parsed?.getMonth()).toBe(2); // Março = índice 2
      expect(parsed?.getDate()).toBe(15);
    });

    it('deve validar datas inválidas no formato DD/MM/AAAA (ex: 31/02/1980)', () => {
      expect(parseDateSafe('31/02/1980')).toBeUndefined();
    });

    it('deve lidar com anos bissextos corretamente no formato DD/MM/AAAA', () => {
      // 1980 é bissexto
      const leap = parseDateSafe('29/02/1980');
      expect(leap).toBeInstanceOf(Date);
      expect(leap?.getDate()).toBe(29);
      expect(leap?.getMonth()).toBe(1);

      // 1981 não é bissexto
      expect(parseDateSafe('29/02/1981')).toBeUndefined();
    });

    it('deve fazer parse de formato ISO e YYYY-MM-DD', () => {
      const iso = parseDateSafe('1980-03-15');
      expect(iso?.getFullYear()).toBe(1980);
      expect(iso?.getMonth()).toBe(2);
      expect(iso?.getDate()).toBe(15);
    });
  });

  describe('formatDate', () => {
    it('deve formatar data para DD/MM/AAAA', () => {
      const d = new Date(1980, 2, 15, 12, 0, 0);
      expect(formatDate(d)).toBe('15/03/1980');
    });

    it('deve formatar string de data para DD/MM/AAAA', () => {
      expect(formatDate('1980-03-15')).toBe('15/03/1980');
    });
  });

  describe('combineDateAndTime', () => {
    it('deve combinar data e horário', () => {
      const d = new Date(1980, 2, 15);
      const combined = combineDateAndTime(d, '14:30');
      expect(combined).toContain('1980-03-15');
    });
  });
});
