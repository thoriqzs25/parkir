import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";

export function Login() {
  const { user, login, loading, error: authError } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState(() => {
    return localStorage.getItem("parkir_desktop_last_email") || "";
  });
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (user) {
      navigate("/locations");
    }
  }, [user, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      await login(email, password);
      navigate("/locations");
    } catch (err) {
      setError("Invalid email or password.");
    }
  };

  return (
    <div className="screen login-screen">
      <div className="card login-card">
        <h1 className="login-title">PARKIR</h1>
        <p className="login-subtitle">Operator Desktop</p>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoFocus
            />
          </div>
          <div className="form-group">
            <label>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          {(error || authError) && (
            <p className="error-message">{error || authError}</p>
          )}
          <button type="submit" className="button primary full" disabled={loading}>
            {loading ? "Logging in..." : "Login"}
          </button>
        </form>
      </div>
    </div>
  );
}
