import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { HowlWrapper } from 'components/Media';
import { Loading } from 'components/UI';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import { GameView, ViewContent, ViewExitButton, ViewTextContent, QRCodeContent, generateClientUrl } from 'components/ViewLayout';
import { FinalQuestions, Question } from './Question';
import { FEUD_API } from './api';

const Music: Record<string, any> = {
  intro: HowlWrapper('/sounds/feud/intro.mp3'),
  round: HowlWrapper('/sounds/feud/round.mp3'),
  beat: HowlWrapper('/sounds/feud/beat.mp3'),
  timer: HowlWrapper('/sounds/feud/timer.mp3'),
  end: HowlWrapper('/sounds/feud/end.mp3'),
};

const Sounds: Record<string, any> = {
  button: HowlWrapper('/sounds/feud/button.mp3'),
  repeat: HowlWrapper('/sounds/feud/repeat.mp3'),
  right: HowlWrapper('/sounds/feud/right.mp3'),
  wrong: HowlWrapper('/sounds/feud/wrong.mp3'),
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

const stateContent = (game: any) => {
  const answerer = game.answerer && game.players.find((t: any) => t.name === game.answerer?.name);
  switch (game.state) {
    case 'waiting_for_players':
      return <QRCodeContent value={generateClientUrl('/feud/client?token=' + game.token)}>{game.token}</QRCodeContent>;
    case 'round': return <ViewTextContent>Round {game.round}</ViewTextContent>;
    case 'button': case 'answers': case 'answers_reveal': case 'final_questions':
      return <div style={{ padding: 16, width: '100%', height: '100%' }}><Question game={game} showHiddenAnswers={false} /></div>;
    case 'final': return <ViewTextContent>Final</ViewTextContent>;
    case 'final_questions_reveal':
      return <div style={{ padding: 16, width: '100%', height: '100%' }}><FinalQuestions game={game} /></div>;
    case 'end':
      return <ViewTextContent>{answerer.finalScore > 200 ? 'Victory' : 'Defeat'}: {answerer.finalScore}</ViewTextContent>;
    default: return <ViewTextContent>Friends Feud</ViewTextContent>;
  }
};

const FeudView: React.FC = () => {
  const [music, setMusic] = useState<string>();

  const game = useGame(FEUD_API, (game) => {
    if (game.state === 'intro') setMusic(changeMusic(music, 'intro'));
    else if (game.state === 'round') setMusic(changeMusic(music, 'round'));
    else if (game.state === 'answers') setMusic(changeMusic(music, 'beat'));
    else if (game.state === 'final_questions') setMusic(changeMusic(music, 'timer'));
    else if (game.state === 'final_questions_reveal') { stopMusic(); setMusic(undefined); }
    else if (game.state === 'end') setMusic(changeMusic(music, 'end'));
  }, (message) => {
    if (Sounds[message]) Sounds[message].play();
    else if (message === 'sound_stop') setMusic(changeMusic(music, ''));
  });

  useEffect(() => { loadSounds(); return stopMusic; }, []);

  const [connected, setConnected] = useAuth(FEUD_API);
  const navigate = useNavigate();
  const onLogout = () => { FEUD_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={FEUD_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  return (
    <GameView>
      <ViewExitButton onClick={onLogout} />
      <ViewContent>{stateContent(game)}</ViewContent>
    </GameView>
  );
};

export default FeudView;
