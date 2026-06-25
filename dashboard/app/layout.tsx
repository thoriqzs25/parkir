import type { Metadata } from "next";
import { AuthProvider } from "@/hooks/useAuth";
import { Toaster } from "sonner";
import "./globals.css";

export const metadata: Metadata = {
  title: "PARKIR Dashboard",
  description: "Parking Administration System",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased">
        <AuthProvider>
          {children}
          <Toaster position="top-right" richColors />
        </AuthProvider>
      </body>
    </html>
  );
}
