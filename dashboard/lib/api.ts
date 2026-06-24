const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function getHealth() {
  const res = await fetch(`${API_BASE_URL}/health`, {
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error("Health check failed");
  }

  return res.json();
}
