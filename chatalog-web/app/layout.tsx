// src/app/layout.tsx
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Chatalog — Media promosi instan dari chat",
  description: "Bikin katalog & brosur otomatis dari WhatsApp.",
  // opsional:
  metadataBase: new URL("https://contoh-domainmu.com"),
  openGraph: {
    title: "Chatalog — Media promosi instan dari chat",
    description: "Bikin katalog & brosur otomatis dari WhatsApp.",
    url: "https://contoh-domainmu.com",
    siteName: "Chatalog",
    images: ["/og-image.png"],
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id">
      <body>{children}</body>
    </html>
  );
}
