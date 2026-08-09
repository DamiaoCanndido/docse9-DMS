import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const token = request.cookies.get('token')?.value;
  const { pathname } = request.nextUrl;

  // Páginas públicas que não exigem login
  const isPublicPage = pathname === '/login';

  // Ignorar arquivos estáticos e favicon
  if (
    pathname.startsWith('/_next') ||
    pathname.startsWith('/api') ||
    pathname.includes('.') ||
    pathname === '/favicon.ico'
  ) {
    return NextResponse.next();
  }

  if (!token && !isPublicPage) {
    // Redireciona para /login se tentar acessar página protegida sem token
    return NextResponse.redirect(new URL('/login', request.url));
  }

  if (token && isPublicPage) {
    // Redireciona para o dashboard se tentar acessar login já autenticado
    return NextResponse.redirect(new URL('/', request.url));
  }

  return NextResponse.next();
}

// Configura o middleware para rodar em todas as rotas exceto as listadas acima
export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
};
