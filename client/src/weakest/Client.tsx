import React from 'react';
import { useNavigate } from 'react-router-dom';

import { Loading, VerticalList, ListItem } from 'components/UI';
import { useAuth, useGame } from 'components/hooks';
import { PlayerAuth } from 'components/Auth';
import { GameClient, ClientContent, ClientHeader, ClientExitButton, ClientTextContent, BigButtonContent } from 'components/ClientLayout';
import { WEAKEST_API } from './api';

const playerStyle = (p: any, player: any): React.CSSProperties => ({
  backgroundColor:
    p.name === player.vote ? 'var(--bg-select)' :
    p.name === player.name ? 'var(--bg-button)' : 'var(--bg-dark)',
  color: p.name === player.vote ? 'var(--text-select)' : undefined,
  cursor: p.name !== player.name ? 'pointer' : undefined,
});

const Players: React.FC<{ game: any; player: any; onClick: (name: string) => void }> = ({ game, player, onClick }) => (
  <VerticalList style={{ width: '100%' } as any}>
    {game.players.filter((p: any) => p.active).map((p: any) => (
      <ListItem
        key={p.name}
        className={p.name !== player.name ? 'clickable' : ''}
        style={playerStyle(p, player) as any}
        onClick={() => p.name !== player.name && onClick(p.name)}
      >
        {p.name}
      </ListItem>
    ))}
  </VerticalList>
);

const stateContent = (game: any) => {
  const player = game.players.find((p: any) => p.name === WEAKEST_API.playerName);
  const buttonActive = game.roundState.answerer?.name === WEAKEST_API.playerName;

  switch (game.state.value) {
    case 'questions':
      return <BigButtonContent active={buttonActive} onClick={() => buttonActive && WEAKEST_API.saveBank()}>Bank ({game.roundState.score})</BigButtonContent>;
    case 'weakest_choose':
      return <Players game={game} player={player} onClick={(name) => WEAKEST_API.vote(name)} />;
    case 'end':
      return <ClientTextContent>Game over</ClientTextContent>;
    default:
      return <ClientTextContent>The Weakest</ClientTextContent>;
  }
};

const WeakestClient: React.FC = () => {
  const game = useGame(WEAKEST_API);
  const [connected, setConnected] = useAuth(WEAKEST_API);
  const navigate = useNavigate();

  if (!connected) return <PlayerAuth api={WEAKEST_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const onLogout = () => { WEAKEST_API.logout(); navigate('/'); };

  return (
    <GameClient>
      <ClientHeader><ClientExitButton onClick={onLogout} /></ClientHeader>
      <ClientContent>{stateContent(game)}</ClientContent>
    </GameClient>
  );
};

export default WeakestClient;
