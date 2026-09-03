import type { NextConfig } from "next";
import { execSync } from "child_process";

function getAppVersion(): string {
  // 1. Variável explícita de ambiente (ex: Netlify ou CI/CD)
  if (process.env.NEXT_PUBLIC_APP_VERSION) {
    return process.env.NEXT_PUBLIC_APP_VERSION;
  }

  // 2. Tag do GitHub Actions (se acionado por tag push ou release)
  if (process.env.GITHUB_REF_TYPE === 'tag' && process.env.GITHUB_REF_NAME) {
    return process.env.GITHUB_REF_NAME;
  }

  // 3. Tenta extrair a tag mais recente do Git local
  try {
    const gitTag = execSync("git describe --tags --abbrev=0", {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    if (gitTag) {
      return gitTag;
    }
  } catch {
    // Continua se não houver tag localmente
  }

  // 4. Se não houver tag, tenta usar a tag com hash do commit ou fallback
  try {
    const gitDescribe = execSync("git describe --tags --always", {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    if (gitDescribe) {
      return gitDescribe.startsWith('v') ? gitDescribe : `v0.1.0 (${gitDescribe})`;
    }
  } catch {
    // Fallback padrão
  }

  return "v0.1.0";
}

const appVersion = getAppVersion();

const nextConfig: NextConfig = {
  env: {
    NEXT_PUBLIC_APP_VERSION: appVersion,
  },
  logging: {
    serverFunctions: false,
  },
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on',
          },
          {
            key: 'Strict-Transport-Security',
            value: 'max-age=63072000; includeSubDomains; preload',
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
          {
            key: 'Permissions-Policy',
            value: 'camera=(), microphone=(), geolocation=(), interest-cohort=()',
          },
        ],
      },
    ];
  },
};

export default nextConfig;
