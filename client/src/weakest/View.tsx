import React, {useEffect, useCallback, useRef, useState} from 'react';
import { useNavigate } from 'react-router-dom';

import { HowlWrapper } from 'components/Media';
import { Loading } from 'components/UI';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import { GameView, ViewContent, ViewExitButton, ViewTextContent, ViewBlockContent, QRCodeContent, generateClientUrl } from 'components/ViewLayout';
import FinalQuestions from './FinalQuestions';
import { WEAKEST_API } from './api';

const Music: Record<string, any> = {
  intro: HowlWrapper('/sounds/weakest/intro.mp3'),
  background: HowlWrapper('/sounds/weakest/background.mp3', true),
  questions: HowlWrapper('/sounds/weakest/questions.mp3', true),
};

const Sounds = {
  gong: HowlWrapper('/sounds/weakest/gong.mp3'),
  question_start: HowlWrapper('/sounds/weakest/question_start.mp3'),
  question_end: HowlWrapper('/sounds/weakest/question_end.mp3'),
  weakest_reveal: HowlWrapper('/sounds/weakest/weakest_reveal.mp3'),
};

const loadSounds = () => {
  Object.values(Music).forEach((m) => m.load());
  Object.values(Sounds).forEach((m) => m.load());
};

const stopMusic = () => Object.values(Music).forEach((m) => m.stop());

const changeMusic = (old: string | undefined, next: string) => {
  if (old !== next) {
    stopMusic();
    if (Music[next]) Music[next].play();
    return next;
  }
  return old;
};

const Timer: React.FC<{ game: any }> = ({ game }) => {
  const [time, setTime] = useState(Math.max(Math.floor(game.roundState.time / 1000), 0));

  useEffect(() => {
    const fromTime = Date.now();
    const timer = setInterval(() => setTime(WEAKEST_API.getTime(fromTime)), 100);
    return () => clearInterval(timer);
  }, [game]);

  return <ViewTextContent>{Math.floor((time % 3600) / 60000)}:{Math.floor(time % 60) / 1000}</ViewTextContent>;
};

const stateContent = (game: any) => {
  switch (game.state.value) {
    case 'waiting_for_players':
      return <QRCodeContent value={generateClientUrl('/weakest/client?token=' + game.token)}>{game.token}</QRCodeContent>;
    case 'round': return <ViewTextContent>Round {game.roundState.number}</ViewTextContent>;
    case 'questions': return <Timer game={game} />;
    case 'weakest_choose': return <ViewTextContent>Choose the Weakest</ViewTextContent>;
    case 'weakest_reveal': return <ViewTextContent>{game.roundState.kicked?.name}</ViewTextContent>;
    case 'final': return <ViewTextContent>Final</ViewTextContent>;
    case 'final_questions': return <ViewBlockContent><FinalQuestions game={game} /></ViewBlockContent>;
    case 'end': return <ViewTextContent>Game over</ViewTextContent>;
    default: return <ViewTextContent>The Weakest</ViewTextContent>;
  }
};

const WeakestView: React.FC = () => {
  const musicRef = useRef<string | undefined>(undefined);

  const onStateChange = useCallback((game: any) => {
    const state = game.state.value;
    if (['intro'].includes(state)) {
      musicRef.current = changeMusic(musicRef.current, 'intro');
    } else if (['questions', 'final_questions'].includes(state)) {
      musicRef.current = changeMusic(musicRef.current, 'questions');
    } else {
      musicRef.current = changeMusic(musicRef.current, 'background');
    }

    if (['questions', 'final_questions'].includes(state)) Sounds.question_start.play();
    else if (state === 'weakest_reveal') Sounds.weakest_reveal.play();
    else if (['weakest_choose', 'end'].includes(state)) Sounds.question_end.play();
  }, []);

  const onIntercom = useCallback((message: string) => {
    if (message === 'gong') Sounds.gong.play();
    else if (message === 'sound_stop') musicRef.current = changeMusic(musicRef.current, '');
  }, []);

  const game = useGame(WEAKEST_API, onStateChange, onIntercom);

  useEffect(() => { loadSounds(); return stopMusic; }, []);

  const [connected, setConnected] = useAuth(WEAKEST_API);
  const navigate = useNavigate();
  const onLogout = () => { WEAKEST_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={WEAKEST_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  return (
    <GameView>
      <ViewExitButton onClick={onLogout} />
      <ViewContent>{stateContent(game)}</ViewContent>
    </GameView>
  );
};

export default WeakestView;
