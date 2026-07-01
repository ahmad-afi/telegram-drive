export interface FileItem {
  msgId: number;
  name: string;
  size: number;
  mimeType: string;
  folder: string;
  date: number;
}

const BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

function session(): string | null {
  return localStorage.getItem("tg_session");
}

function authHeaders(): Record<string, string> {
  const s = session();
  return s ? { "X-Session": s } : {};
}

async function req<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = { ...authHeaders(), ...(init.headers as Record<string, string>) };
  const res = await fetch(`${BASE}/api${path}`, { ...init, headers });
  if (res.status === 401) {
    localStorage.removeItem("tg_session");
    location.reload();
    throw new Error("unauthorized");
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  async qrStart(): Promise<{ loginId: string; qrImage: string }> {
    const res = await fetch(`${BASE}/api/qr/start`, { method: "POST" });
    if (!res.ok) throw new Error("QR generation failed");
    return res.json();
  },

  async qrPoll(loginId: string): Promise<{ status: string; session?: string; name?: string }> {
    const res = await fetch(`${BASE}/api/qr/poll/${loginId}`);
    if (!res.ok) throw new Error("QR poll failed");
    return res.json();
  },

  listFolders: () => req<string[]>("/folders"),

  listFiles: (folder: string, offset = 0, limit = 0) =>
    req<FileItem[]>(`/files?folder=${encodeURIComponent(folder)}&offset=${offset}&limit=${limit}`),

  searchFiles: (q: string, folder = "") =>
    req<FileItem[]>(`/files?q=${encodeURIComponent(q)}${folder ? `&folder=${encodeURIComponent(folder)}` : ""}`),

  uploadFile(file: File, folder: string, onProgress?: (pct: number) => void): Promise<{ ok: boolean }> {
    const form = new FormData();
    form.append("file", file);
    form.append("folder", folder);
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open("POST", `${BASE}/api/files`);
      const s = session();
      if (s) xhr.setRequestHeader("X-Session", s);
      if (onProgress && xhr.upload) {
        xhr.upload.onprogress = (e) => {
          if (e.lengthComputable) onProgress((e.loaded / e.total) * 100);
        };
      }
      xhr.onload = () => {
        if (xhr.status === 401) {
          localStorage.removeItem("tg_session");
          location.reload();
          reject(new Error("unauthorized"));
          return;
        }
        if (xhr.status < 200 || xhr.status >= 300) {
          try { reject(new Error(JSON.parse(xhr.responseText).message)); }
          catch { reject(new Error(xhr.statusText)); }
          return;
        }
        resolve({ ok: true });
      };
      xhr.onerror = () => reject(new Error("upload failed"));
      xhr.send(form);
    });
  },

  deleteFile: (msgId: number) =>
    req<void>(`/files/${msgId}`, { method: "DELETE" }),

  async downloadFile(msgId: number): Promise<Blob> {
    const res = await fetch(`${BASE}/api/files/${msgId}/download`, {
      headers: authHeaders(),
    });
    if (!res.ok) throw new Error("download failed");
    return res.blob();
  },
};
