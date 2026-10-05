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

const Music: Record<string, any> = {
  theme: HowlWrapper('/sounds/deal/theme.mp3'),
  heartbeat: HowlWrapper('/sounds/deal/heartbeat.mp3', true),
  bed: HowlWrapper('/sounds/deal/bed-1.mp3', true),
  end: HowlWrapper('/sounds/deal/250-000.mp3'),
  sting: HowlWrapper('/sounds/deal/sting.mp3'),
};

const Sounds: Record<string, any> = {
  box_open: HowlWrapper('/sounds/deal/box-open.mp3'),
  select: HowlWrapper('/sounds/deal/player-select.mp3'),
  phone_ring: HowlWrapper('/sounds/deal/phone-ring.mp3'),
};

const loadSounds = () => {
  Object.values(Music).forEach((m) => m.load());
  Object.values(Sounds).forEach((m) => m.load());
};

const stopMusic = () => Object.values(Music).forEach((m) => m.stop());

const playAmbience = (state: string) => {
  stopMusic();
  switch (state) {
    case 'round_start':
      Music.theme.play();
      break;
    case 'case_select':
    case 'haggle':
    case 'case_swap':
    case 'reveal':
    case 'wheel_start':
      Music.heartbeat.play();
      if (state === 'haggle' || state === 'reveal') Music.sting.play();
      break;
    case 'wheel_spin':
      Music.bed.play();
      break;
    case 'round_end':
    case 'game_end':
      Music.end.play();
      break;
    default:
      break;
  }
};

const stateContent = (game: any, onWheelStop: () => void, onWheelRespin: () => void) => {
  switch (game.state.value) {
    case 'round_start':
      return <ViewTextContent>Round {game.state.round}</ViewTextContent>;
    case 'case_select':
    case 'haggle':
    case 'case_swap':
      return (
        <div style={{ padding: 16, width: '100%', height: '100%' }}>
          <CasesTable game={game} showContent={false} />
        </div>
      );
    case 'reveal':
      return (
        <div style={{ padding: 16, width: '100%', height: '100%' }}>
          <CasesTable game={game} showContent />
        </div>
      );
    case 'wheel_start':
    case 'wheel_spin':
      return <Wheel game={game} onStop={onWheelStop} onRespin={onWheelRespin} />;
    case 'round_end':
      if (game.state.chosenMoney) {
        return <ViewTextContent>Money</ViewTextContent>;
      }
      {
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
    else if (message === 'phone_ring') Sounds.phone_ring.play();
    else if (message === 'sound_stop') stopMusic();
  }, []);

  const game = useGame(DEAL_API, () => {}, onIntercom);

  const onWheelStop = useCallback(() => {
    Music.bed.stop();
    Sounds.select.play();
  }, []);

  const onWheelRespin = useCallback(() => {
    Sounds.select.play();
  }, []);

  useEffect(() => {
    loadSounds();
    return stopMusic;
  }, []);

  useEffect(() => {
    if (game?.state?.value) playAmbience(game.state.value);
  }, [game?.state?.value]);

  const [connected, setConnected] = useAuth(DEAL_API);
  const navigate = useNavigate();
  const onLogout = () => { DEAL_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={DEAL_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  return (
    <GameView>
      <ViewExitButton onClick={onLogout} />
      <div style={{ display: 'flex', flexGrow: 1, height: 0 }}>
        <ViewContent>{stateContent(game, onWheelStop, onWheelRespin)}</ViewContent>
        <div style={{ width: '25%', minWidth: '25%', padding: 16, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
          <PresentsList game={game} showDescription={false} />
        </div>
      </div>
    </GameView>
  );
};

export default DealView;
