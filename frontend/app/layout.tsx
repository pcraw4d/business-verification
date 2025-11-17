import { PerformanceOptimizer } from "@/components/performance/PerformanceOptimizer";
import { Toaster } from "@/components/ui/sonner";
import { getApiBaseUrl } from "@/lib/api-config";
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
  display: "swap", // Optimize font loading
  preload: true,
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
  display: "swap",
  preload: true,
});

export const metadata: Metadata = {
  title: "KYB Platform - Merchant Details",
  description: "Know Your Business verification platform",
  viewport: "width=device-width, initial-scale=1",
  themeColor: "#000000",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        {/* Preconnect to API for faster requests */}
        <link rel="preconnect" href={getApiBaseUrl()} crossOrigin="anonymous" />
        <link rel="dns-prefetch" href={getApiBaseUrl()} />
        {/* Preconnect to Google Fonts */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        {children}
        <Toaster />
        <PerformanceOptimizer />
      </body>
    </html>
  );
}
