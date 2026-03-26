import React from 'react';
import { useNavigate } from 'react-router-dom';

import { Loading } from 'components/UI';
import { useAuth, useGame } from 'components/hooks';
import { PlayerAuth } from 'components/Auth';
import { GameClient, ClientContent, ClientHeader, ClientExitButton, ClientTextContent, BigButtonContent } from 'components/ClientLayout';
import { FEUD_API } from './api';

const stateContent = (game: any) => {
  const buttonActive = !game.answerer;
  const onBtn = () => buttonActive && FEUD_API.buttonClick(FEUD_API.playerName!);

  switch (game.state) {
    case 'button': return <BigButtonContent active={buttonActive} onClick={onBtn} />;
    case 'end': return <ClientTextContent>Game over</ClientTextContent>;
    default: return <ClientTextContent>Friends Feud</ClientTextContent>;
  }
};

const FeudClient: React.FC = () => {
  const game = useGame(FEUD_API);
  const [connected, setConnected] = useAuth(FEUD_API);
  const navigate = useNavigate();

  if (!connected) return <PlayerAuth api={FEUD_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const onLogout = () => { FEUD_API.logout(); navigate('/'); };

  return (
    <GameClient>
      <ClientHeader><ClientExitButton onClick={onLogout} /></ClientHeader>
      <ClientContent>{stateContent(game)}</ClientContent>
    </GameClient>
  );
};

export default FeudClient;
