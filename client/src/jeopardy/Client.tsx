import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { HowlWrapper } from 'components/Media';
import { Loading, Toast } from 'components/UI';
import { useAuth, useGame } from 'components/hooks';
import { PlayerAuth } from 'components/Auth';
import { ClientExitButton, ClientHeader, GameClient, ClientContent } from 'components/ClientLayout';
import { JEOPARDY_API } from './api';

const Sounds = {
  do_bet: HowlWrapper('/sounds/jeopardy/do_bet.mp3'),
  schnelle: HowlWrapper('/sounds/jeopardy/schnelle.mp3'),
};

const loadSounds = () => Object.values(Sounds).forEach((m) => m.load());

const formInputStyle: React.CSSProperties = {
  display: 'block', width: '100%', padding: 10,
  fontSize: 18, backgroundColor: 'var(--bg-button)', color: 'white', border: 'none', marginBottom: 10,
};

const submitBtnStyle: React.CSSProperties = {
  display: 'flex', justifyContent: 'center', backgroundColor: 'var(--bg-button)',
  color: 'var(--text)', padding: '8px 12px', cursor: 'pointer',
};

const FinalBet: React.FC = () => {
  const [bet, setBet] = useState('');
  return (
    <div>
      <input style={formInputStyle} type="number" onChange={(e) => setBet(e.target.value)} value={bet} />
      <div style={submitBtnStyle} onClick={() => JEOPARDY_API.finalBet(parseInt(bet))}>Submit</div>
    </div>
  );
};

const FinalAnswer: React.FC = () => {
  const [answer, setAnswer] = useState('');
  return (
    <div>
      <input style={formInputStyle} type="text" onChange={(e) => setAnswer(e.target.value)} value={answer} />
      <div style={submitBtnStyle} onClick={() => JEOPARDY_API.finalAnswer(answer)}>Submit</div>
    </div>
  );
};

const Content: React.FC<{ game: any }> = ({ game }) => {
  const buttonActive = game.state.value === 'answer' && !game.state.answerer;
  const player = game.players.find((p: any) => p.name === JEOPARDY_API.playerName);
  const onBtn = () => { if (buttonActive) JEOPARDY_API.buttonClick(); };

  let content;
  if (['question', 'answer'].includes(game.state.value)) {
    content = (
      <div
        style={{
          width: '90vmin', height: '90vmin', borderRadius: '50%',
          border: '4px solid var(--bg-dark)',
          backgroundColor: buttonActive ? '#c63939' : 'var(--bg-button)',
          boxShadow: '0 8px 0 var(--bg-dark)',
        }}
        onClick={onBtn} onTouchStart={onBtn}
      />
    );
  } else if (game.state.value === 'final_bets' && player.finalBet === 0) {
    content = <FinalBet />;
  } else if (game.state.value === 'final_answer' && !player.finalAnswer) {
    content = <FinalAnswer />;
  } else {
    content = <div style={{ fontSize: 40, fontWeight: 'bold' }}>Jeopardy</div>;
  }

  return <ClientContent>{content}</ClientContent>;
};

const playerStyle = (p: any, game: any): React.CSSProperties => ({
  display: 'flex', flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
  fontSize: 16, minHeight: 40, padding: '3px 10px', margin: 2,
  backgroundColor: p.name === game.state.answerer?.name ? 'var(--bg-select)' : p.name === JEOPARDY_API.playerName ? 'var(--bg-button)' : 'var(--bg-dark)',
  color: p.name === game.state.answerer?.name ? 'var(--text-select)' : undefined,
});

const Players: React.FC<{ game: any }> = ({ game }) => (
  <div style={{ display: 'flex', flexDirection: 'column' }}>
    {game.players.map((p: any) => (
      <div key={p.name} style={playerStyle(p, game)}>
        <div style={{ wordBreak: 'break-all' }}>{p.name}</div>
        <div style={{ fontWeight: 'bold', fontSize: 18, wordBreak: 'break-all' }}>{p.balance}</div>
      </div>
    ))}
  </div>
);

const JeopardyClient: React.FC = () => {
  const game = useGame(JEOPARDY_API, () => {}, (message) => {
    if (message === 'do_bet:' + JEOPARDY_API.playerName) Sounds.do_bet.play();
    else if (message === 'do_answer:' + JEOPARDY_API.playerName) Sounds.schnelle.play();
  });
  const [connected, setConnected] = useAuth(JEOPARDY_API);
  const navigate = useNavigate();

  useEffect(loadSounds, []);

  if (!connected) return <PlayerAuth api={JEOPARDY_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const onLogout = () => { JEOPARDY_API.logout(); navigate('/'); };

  return (
    <GameClient>
      <div style={{ display: 'flex', flexDirection: 'column', flexGrow: 1, width: '100%' }}>
        <ClientHeader><ClientExitButton onClick={onLogout} /></ClientHeader>
        <Content game={game} />
        <Players game={game} />
      </div>
      <Toast />
    </GameClient>
  );
};

export default JeopardyClient;
