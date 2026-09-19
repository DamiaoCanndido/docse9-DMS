/**
 * Retorna a versão da aplicação obtida dinamicamente do GitHub Tag / Git / Ambiente de Build.
 */
export const APP_VERSION =
  process.env.NEXT_PUBLIC_APP_VERSION || 'v1.1.0';
