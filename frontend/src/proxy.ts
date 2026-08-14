import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function proxy(request: NextRequest) {
  const token = request.cookies.get('token')?.value;
  const { pathname } = request.nextUrl;

  // Páginas públicas que não exigem login
  const isPublicPage = pathname === '/login';

  // Ignorar arquivos estáticos específicos e rotas internas
  const isStaticFile = /\.(?:ico|png|jpg|jpeg|svg|css|js|webp|woff|woff2|ttf|eot)$/i.test(pathname);
  if (
    pathname.startsWith('/_next') ||
    pathname.startsWith('/api') ||
    isStaticFile ||
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

// Configura o proxy para rodar em todas as rotas da aplicação exceto assets estáticos
export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:ico|png|jpg|jpeg|svg|css|js|webp|woff|woff2|ttf|eot)$).*)'],
};
