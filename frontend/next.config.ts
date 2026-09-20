import type { NextConfig } from "next";
import { execSync } from "child_process";
import packageJson from "./package.json";

function getAppVersion(): string {
  // 1. Variável explícita de ambiente (ex: Netlify ou CI/CD)
  if (process.env.NEXT_PUBLIC_APP_VERSION) {
    const v = process.env.NEXT_PUBLIC_APP_VERSION;
    return v.startsWith('v') ? v : `v${v}`;
  }

  // 2. Tag do Netlify ou GitHub Actions
  if (process.env.TAG) {
    const v = process.env.TAG;
    return v.startsWith('v') ? v : `v${v}`;
  }
  if (process.env.GITHUB_REF_TYPE === 'tag' && process.env.GITHUB_REF_NAME) {
    const v = process.env.GITHUB_REF_NAME;
    return v.startsWith('v') ? v : `v${v}`;
  }

  // 3. Versão oficial do package.json
  if (packageJson && packageJson.version) {
    return `v${packageJson.version}`;
  }

  // 4. Tenta extrair a tag mais recente do Git local
  try {
    const gitTag = execSync("git describe --tags --abbrev=0", {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    if (gitTag) {
      return gitTag.startsWith('v') ? gitTag : `v${gitTag}`;
    }
  } catch {
    // Continua se não houver tag localmente
  }

  return "v1.2.0";
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
