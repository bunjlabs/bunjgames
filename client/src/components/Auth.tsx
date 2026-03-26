import React, { useState, useEffect, useRef } from 'react';
import { useLocation } from 'react-router-dom';
import { toast } from 'react-toastify';
import { Loading, Toast } from './UI';
import type GameApi from 'services/GameApi';

const useQuery = () => new URLSearchParams(useLocation().search);

const authInputStyle: React.CSSProperties = {
  display: 'block', width: '100%', boxSizing: 'border-box',
  padding: 10, fontSize: 18, fontWeight: 'bold', textTransform: 'uppercase', marginBottom: 10,
};

const authButtonStyle = (loading: boolean): React.CSSProperties => ({
  display: 'flex', justifyContent: 'center', alignItems: 'center',
  backgroundColor: 'var(--bg-button)', color: loading ? 'var(--text-gray)' : 'var(--text)',
  padding: '8px 12px', cursor: 'pointer',
});

const GameAuth: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div style={{ height: '100%', backgroundColor: 'var(--bg-dark)', display: 'flex', justifyContent: 'center', alignItems: 'center', flexDirection: 'column' }}>
    {children}
    <Toast />
  </div>
);

const AuthForm: React.FC<React.PropsWithChildren<{ title: string }>> = ({ title, children }) => (
  <div style={{ width: '100%', maxWidth: 500, backgroundColor: 'var(--bg-base)', margin: 16, padding: 32, display: 'flex', flexDirection: 'column', alignSelf: 'center' }}>
    <div style={{ color: 'white', fontSize: 24, marginBottom: 20 }}>{title}</div>
    <div>{children}</div>
  </div>
);

const GameCreateForm: React.FC<{
  api: GameApi & { createGame: (f: HTMLInputElement) => Promise<any> };
  setConnected: (v: boolean) => void;
}> = ({ api, setConnected }) => {
  const [loading, setLoading] = useState(false);
  const inputFile = useRef<HTMLInputElement>(null);

  const onSubmit = () => {
    if (!inputFile.current?.files?.length) {
      toast.dark('Please select game file');
      return;
    }
    setLoading(true);
    api
      .createGame(inputFile.current)
      .then(() => api.connect().then(() => setConnected(true)))
      .catch((e: any) => {
        setLoading(false);
        if (!e.response) toast.dark(e.message);
        else if (e.response.status === 400 && e.response.data) toast.dark(e.response.data.detail);
        else toast.dark('Error while creating game');
      });
  };

  return (
    <AuthForm title="Create game">
      <input ref={inputFile} type="file" disabled={loading} style={{ color: 'var(--text)', marginBottom: 10, fontSize: 18 }} />
      <div style={authButtonStyle(loading)} onClick={onSubmit}>Create</div>
    </AuthForm>
  );
};

const GameOpenForm: React.FC<{
  api: GameApi;
  setConnected: (v: boolean) => void;
}> = ({ api, setConnected }) => {
  const [loading, setLoading] = useState(false);
  const [token, setToken] = useState('');

  const onSubmit = () => {
    if (!token) {
      toast.dark('Please enter token');
      return;
    }
    setLoading(true);
    api.connect(token).then(() => setConnected(true)).catch(() => {
      setLoading(false);
      toast.dark('Game not found');
    });
  };

  return (
    <AuthForm title="Open game">
      <input style={authInputStyle} type="text" placeholder="token" value={token} onChange={(e) => setToken(e.target.value)} disabled={loading} />
      <div style={authButtonStyle(loading)} onClick={onSubmit}>Open</div>
    </AuthForm>
  );
};

const RegisterPlayerForm: React.FC<{
  api: GameApi & { registerPlayer: (t: string, n: string) => Promise<any> };
  setConnected: (v: boolean) => void;
}> = ({ api, setConnected }) => {
  const [loading, setLoading] = useState(false);
  const query = useQuery();
  const tokenFromQuery = query.get('token') || '';
  const [token, setToken] = useState(tokenFromQuery);
  const [name, setName] = useState(api.getSavedUsername() ?? '');

  const onSubmit = () => {
    if (!token || !name) {
      toast.dark('Please enter token and name');
      return;
    }
    setLoading(true);
    api
      .registerPlayer(token, name)
      .then(() =>
        api.connect().then(() => {
          api.setSavedUsername(name.trim());
          setConnected(true);
        }),
      )
      .catch((e: any) => {
        setLoading(false);
        if (!e.response) toast.dark(e.message);
        else if (e.response.status === 400 && e.response.data) toast.dark(e.response.data.detail);
        else toast.dark('Error while registering player');
      });
  };

  return (
    <AuthForm title="Register player">
      <input style={authInputStyle} type="text" maxLength={20} placeholder="name" value={name} onChange={(e) => setName(e.target.value)} disabled={loading} />
      <input style={authInputStyle} type="text" placeholder="token" value={token} onChange={(e) => setToken(e.target.value)} disabled={loading} />
      <div style={authButtonStyle(loading)} onClick={onSubmit}>Register</div>
    </AuthForm>
  );
};

export const AdminAuth: React.FC<{
  api: GameApi & { createGame: (f: HTMLInputElement) => Promise<any> };
  setConnected: (v: boolean) => void;
}> = ({ api, setConnected }) => {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (api.hasToken()) {
      api.connect().then(() => setConnected(true)).catch(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [api, setConnected]);

  if (loading) return <Loading />;

  return (
    <GameAuth>
      <GameCreateForm api={api} setConnected={setConnected} />
      <GameOpenForm api={api} setConnected={setConnected} />
    </GameAuth>
  );
};

type PlayerApi = GameApi & {
  registerPlayer: (t: string, n: string) => Promise<any>;
} & (
  | { playerId: string | null; hasPlayerId: () => boolean }
  | { playerName: string | null; hasPlayerName: () => boolean }
);

const getPlayerIdentifier = (api: PlayerApi): { value: string | null; field: string; has: boolean } => {
  if ('playerName' in api) return { value: api.playerName, field: 'name', has: api.hasPlayerName() };
  return { value: api.playerId, field: 'id', has: api.hasPlayerId() };
};

export const PlayerAuth: React.FC<{
  api: PlayerApi;
  setConnected: (v: boolean) => void;
}> = ({ api, setConnected }) => {
  const [loading, setLoading] = useState(true);
  const query = useQuery();
  const tokenParam = query.get('token');

  useEffect(() => {
    const { value, field, has } = getPlayerIdentifier(api);
    const checkGame = (game: any) =>
      !game.players || !value || game.players.find((p: any) => p[field] === value);

    if (api.hasToken() && has && !tokenParam) {
      api.connect(api.token, checkGame).then(() => setConnected(true)).catch(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [api, tokenParam, setConnected]);

  if (loading) return <Loading />;

  return (
    <GameAuth>
      <RegisterPlayerForm api={api} setConnected={setConnected} />
    </GameAuth>
  );
};
