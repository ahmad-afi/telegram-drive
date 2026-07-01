import { useEffect, useRef, useState } from "react";
import { api } from "./api";

export default function Login({ onAuth }: { onAuth: () => void }) {
  const [qrImage, setQrImage] = useState("");
  const [loginId, setLoginId] = useState("");
  const [err, setErr] = useState("");
  const [status, setStatus] = useState("Menunggu QR...");
  const pollRef = useRef<number | undefined>(undefined);

  async function startQR() {
    setErr("");
    setStatus("Membuat QR...");
    try {
      const { loginId: id, qrImage: img } = await api.qrStart();
      setLoginId(id);
      setQrImage(img);
      setStatus("Scan QR pakai app Telegram di HP (Settings → Devices → Link Device)");
    } catch (e) {
      setErr((e as Error).message);
    }
  }

  useEffect(() => {
    if (!loginId) return;
    const poll = async () => {
      try {
        const res = await api.qrPoll(loginId);
        if (res.status === "ok" && res.session) {
          localStorage.setItem("tg_session", res.session);
          setStatus(`Berhasil! Masuk sebagai ${res.name || "user"}`);
          setTimeout(onAuth, 800);
          return;
        }
      } catch {
        // keep polling
      }
      pollRef.current = window.setTimeout(poll, 2000);
    };
    poll();
    return () => {
      if (pollRef.current) clearTimeout(pollRef.current);
    };
  }, [loginId, onAuth]);

  return (
    <div className="center">
      <div className="card login">
        <h1>Telegram Drive</h1>
        <p className="hint">Login pakai akun Telegram kamu. File disimpan di Saved Messages akunmu sendiri.</p>
        {qrImage && (
          <div className="qr">
            <img src={`data:image/png;base64,${qrImage}`} alt="QR Login" />
          </div>
        )}
        {err && <p className="error">{err}</p>}
        <p className="status">{status}</p>
        <button onClick={startQR} disabled={!qrImage || status.startsWith("Berhasil")}>
          {qrImage ? "Buat Ulang QR" : "Mulai Login"}
        </button>
        <p className="warning">
          Session disimpan di browser kamu. Server tidak menyimpan apa-apa.
        </p>
      </div>
    </div>
  );
}
