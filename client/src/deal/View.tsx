import React, { useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

import { HowlWrapper } from 'components/Media';
import { Loading } from 'components/UI';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import { GameView, ViewContent, ViewExitButton, ViewTextContent } from 'components/ViewLayout';
import { CasesTable } from './CasesTable';
import { PresentsList } from './PresentsList';
import Wheel from './Wheel';
import { DEAL_API } from './api';

const Music = {
  theme: HowlWrapper('/sounds/deal/theme.mp3', true),
};

const Sounds = {
  box_open: HowlWrapper('/sounds/deal/box-open.mp3'),
  select: HowlWrapper('/sounds/deal/player-select.mp3'),
};

const loadSounds = () => {
  Object.values(Music).forEach((m) => m.load());
  Object.values(Sounds).forEach((m) => m.load());
};

const stopMusic = () => Object.values(Music).forEach((m) => m.stop());

const stateContent = (game: any) => {
  switch (game.state.value) {
    case 'round_start':
      return <ViewTextContent>Round {game.state.round}</ViewTextContent>;
    case 'case_select':
    case 'haggle':
    case 'case_swap':
      return <CasesTable game={game} showContent={false} />;
    case 'reveal':
      return (
        <ViewTextContent>
          {game.state.revealedIndex >= 0 && game.presents[game.state.revealedIndex]?.content}
        </ViewTextContent>
      );
    case 'wheel_start':
    case 'wheel_spin':
      return <Wheel game={game} onStop={() => DEAL_API.nextState(game.state.value)} />;
    case 'round_end': {
      const idx = game.state.wheelResultIndex >= 0 ? game.state.wheelResultIndex : game.state.revealedIndex;
      return <ViewTextContent>{idx >= 0 && game.presents[idx]?.content}</ViewTextContent>;
    }
    case 'game_end':
      return <ViewTextContent>Deal or No Deal</ViewTextContent>;
    default:
      return <ViewTextContent>Deal or No Deal</ViewTextContent>;
  }
};

const DealView: React.FC = () => {
  const onIntercom = useCallback((message: string) => {
    if (message === 'box_open') Sounds.box_open.play();
    else if (message === 'select') Sounds.select.play();
    else if (message === 'sound_stop') stopMusic();
  }, []);

  const game = useGame(DEAL_API, () => {}, onIntercom);

  useEffect(() => {
    loadSounds();
    Music.theme.play();
    return stopMusic;
  }, []);

  const [connected, setConnected] = useAuth(DEAL_API);
  const navigate = useNavigate();
  const onLogout = () => { DEAL_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={DEAL_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  return (
    <GameView>
      <ViewExitButton onClick={onLogout} />
      <div style={{ display: 'flex', flexGrow: 1, height: 0 }}>
        <ViewContent>{stateContent(game)}</ViewContent>
        <div style={{ width: '25%', minWidth: '25%', borderLeft: '8px solid var(--bg-dark)', padding: 16, overflow: 'hidden' }}>
          <PresentsList game={game} showDescription={false} />
        </div>
      </div>
    </GameView>
  );
};

export default DealView;
