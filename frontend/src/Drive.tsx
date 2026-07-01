import { useEffect, useRef, useState } from "react";
import { api, type FileItem } from "./api";
import { fmtSize } from "./fmt";

const ROOT = "root";
const PAGE_SIZE = 50;

export default function Drive({ onLogout }: { onLogout: () => void }) {
  const [folders, setFolders] = useState<string[]>([]);
  const [files, setFiles] = useState<FileItem[]>([]);
  const [current, setCurrent] = useState(ROOT);
  const [q, setQ] = useState("");
  const [busy, setBusy] = useState(false);
  const [progress, setProgress] = useState(0);
  const [newFolder, setNewFolder] = useState("");
  const [showNewFolder, setShowNewFolder] = useState(false);
  const [page, setPage] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [dragover, setDragover] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const pendingFolderRef = useRef<string | null>(null);

  async function refresh() {
    try {
      if (q) {
        setFolders([]);
        setFiles(await api.searchFiles(q, current === ROOT ? "" : current));
        setHasMore(false);
      } else {
        const [folders, fi] = await Promise.all([
          api.listFolders(),
          api.listFiles(current, 0, PAGE_SIZE),
        ]);
        setFolders(folders);
        setFiles(fi);
        setHasMore(fi.length === PAGE_SIZE);
      }
      setPage(0);
    } catch (e) {
      console.error(e);
    }
  }

  async function loadMore() {
    const next = page + 1;
    try {
      const fi = await api.listFiles(current, next * PAGE_SIZE, PAGE_SIZE);
      setFiles((prev) => [...prev, ...fi]);
      setHasMore(fi.length === PAGE_SIZE);
      setPage(next);
    } catch (e) {
      console.error(e);
    }
  }

  useEffect(() => {
    refresh();
  }, [current, q]);

  function uploadFile(file: File) {
    setBusy(true);
    setProgress(0);
    const targetFolder = pendingFolderRef.current ?? (current === ROOT ? "root" : current);
    pendingFolderRef.current = null;
    api
      .uploadFile(file, targetFolder, setProgress)
      .then(() => {
        setCurrent(targetFolder);
        refresh();
      })
      .catch((err) => alert(err.message))
      .finally(() => {
        setBusy(false);
        setProgress(0);
        if (inputRef.current) inputRef.current.value = "";
      });
  }

  function onUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0];
    if (!f) return;
    uploadFile(f);
  }

  function onDrop(e: React.DragEvent) {
    e.preventDefault();
    setDragover(false);
    const f = e.dataTransfer.files?.[0];
    if (f) uploadFile(f);
  }

  function uploadToNewFolder(e: React.FormEvent) {
    e.preventDefault();
    if (!newFolder.trim()) return;
    pendingFolderRef.current = newFolder.trim();
    inputRef.current?.click();
  }

  async function downloadFile(msgId: number, name: string) {
    try {
      const blob = await api.downloadFile(msgId);
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = name;
      a.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      alert((err as Error).message);
    }
  }

  function deleteFile(msgId: number) {
    if (!confirm("Hapus file ini?")) return;
    api
      .deleteFile(msgId)
      .then(() => setFiles((prev) => prev.filter((f) => f.msgId !== msgId)))
      .catch((err) => alert(err.message));
  }

  return (
    <div className="app">
      <header>
        <div className="search">
          <input
            placeholder="Cari file..."
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </div>
        <button onClick={onLogout}>Keluar</button>
      </header>

      <div className="body">
        <aside>
          <button onClick={() => inputRef.current?.click()} disabled={busy}>
            {busy ? `Uploading... ${Math.round(progress)}%` : "Upload File"}
          </button>
          {busy && (
            <div className="progress-bar">
              <div className="progress-fill" style={{ width: `${progress}%` }} />
            </div>
          )}
          <input ref={inputRef} type="file" hidden onChange={onUpload} />

          <button onClick={() => setShowNewFolder(!showNewFolder)}>
            Upload ke Folder Baru
          </button>
          {showNewFolder && (
            <form onSubmit={uploadToNewFolder}>
              <input
                placeholder="Nama folder"
                value={newFolder}
                onChange={(e) => setNewFolder(e.target.value)}
                autoFocus
              />
              <button type="submit">Upload ke sini</button>
            </form>
          )}

          <nav>
            <div
              className={`nav-item ${current === ROOT ? "active" : ""}`}
              onClick={() => { setCurrent(ROOT); setQ(""); }}
            >
              Root
            </div>
            {folders.map((f) => (
              <div
                key={f}
                className={`nav-item ${current === f ? "active" : ""}`}
                onClick={() => { setCurrent(f); setQ(""); }}
              >
                {f}
              </div>
            ))}
          </nav>
        </aside>

        <main
          onDragOver={(e) => { e.preventDefault(); setDragover(true); }}
          onDragLeave={() => setDragover(false)}
          onDrop={onDrop}
        >
          <div className={`drop-zone ${dragover ? "dragover" : ""}`} onClick={() => inputRef.current?.click()}>
            {busy ? `Uploading... ${Math.round(progress)}%` : "Drop file di sini atau klik untuk upload"}
          </div>
          {!files.length && q && <p className="empty">Tidak ada hasil.</p>}
          {files.map((f) => (
            <div key={f.msgId} className="file-row">
              <span className="file-icon">📄</span>
              <div className="file-info">
                <span className="file-name" onClick={() => downloadFile(f.msgId, f.name)}>
                  {f.name}
                </span>
                <small>{fmtSize(f.size)}</small>
              </div>
              <div className="file-actions">
                <button className="link" onClick={() => downloadFile(f.msgId, f.name)}>
                  Download
                </button>
                <button className="mini" onClick={() => deleteFile(f.msgId)}>
                  Hapus
                </button>
              </div>
            </div>
          ))}
          {hasMore && (
            <button className="load-more" onClick={loadMore}>Muat lebih banyak</button>
          )}
        </main>
      </div>
    </div>
  );
}
