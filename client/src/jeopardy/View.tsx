import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { AudioPlayer, ImagePlayer, VideoPlayer, HowlWrapper } from 'components/Media';
import { Loading } from 'components/UI';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import { GameView, ViewContent, ViewExitButton, ViewTextContent, QRCodeContent, generateClientUrl } from 'components/ViewLayout';
import { ThemesList, ThemesGrid, QuestionsGrid } from './Themes';
import { getRoundName, EventType } from './Common';
import { JEOPARDY_API } from './api';

const Music = {
  intro: HowlWrapper('/sounds/jeopardy/intro.mp3'),
  themes: HowlWrapper('/sounds/jeopardy/themes.mp3'),
  round: HowlWrapper('/sounds/jeopardy/round.mp3'),
  minute: HowlWrapper('/sounds/jeopardy/minute.mp3'),
  auction: HowlWrapper('/sounds/jeopardy/auction.mp3'),
  bagcat: HowlWrapper('/sounds/jeopardy/bagcat.mp3'),
  game_end: HowlWrapper('/sounds/jeopardy/game_end.mp3'),
};

const Sounds = { skip: HowlWrapper('/sounds/jeopardy/skip.mp3') };

const loadSounds = () => {
  Object.values(Music).forEach((m) => m.load());
  Object.values(Sounds).forEach((m) => m.load());
};
const resetSounds = () => Object.values(Music).forEach((m) => m.stop());

const delocalize = (url: string) => {
  if (url.slice(7).includes('https:')) {
    return url.replace(window.location.protocol + '//' + window.location.hostname, '');
  }
  return url;
};

const mediaStyle: React.CSSProperties = {
  display: 'flex', flexDirection: 'column', alignItems: 'center',
  justifyContent: 'center', width: '100%', height: '100%',
  fontSize: 38, color: 'var(--text)'
};
const fullImg: React.CSSProperties = {
  width: '95vw', height: '95vh',
  display: 'flex', justifyContent: 'center', alignItems: 'center',
};

const QuestionMessage: React.FC<{
  game: any; text?: string; image?: string; audio?: string; video?: string; isContentPlaying?: boolean;
}> = ({ game, text, image, audio, video, isContentPlaying }) => {
  if (game.state.value === "final_answer") {
    audio = undefined; video = undefined;
  }
  return (
      <div style={mediaStyle}>
        {text && !image && !video && <p style={{ textAlign: 'center', padding: 16 }}>{text}</p>}
        {image && <div style={fullImg}><ImagePlayer game={game} url={delocalize(image)} style={{ width: '100%', height: '100%' }} /></div>}
        {audio && <AudioPlayer controls playing={isContentPlaying} game={game} url={delocalize(audio)} />}
        {video && <div style={fullImg}><VideoPlayer controls playing={isContentPlaying} game={game} url={delocalize(video)} style={{ width: '100%', height: '100%' }} /></div>}
      </div>
  )
};

const stateContent = (game: any, isContentPlaying: boolean) => {
  switch (game.state.value) {
    case 'waiting_for_players':
      return <QRCodeContent value={generateClientUrl('/jeopardy/client?token=' + game.token)}>{game.token}</QRCodeContent>;
    case 'themes_all': return <ThemesGrid game={game} />;
    case 'round': return <ViewTextContent>{getRoundName(game)}</ViewTextContent>;
    case 'round_themes': case 'final_themes': return <ThemesList game={game} />;
    case 'questions': return <QuestionsGrid game={game} />;
    case 'question_event': return <ViewTextContent><EventType type={game.state.question.type} /></ViewTextContent>;
    case 'question': case 'answer': case 'final_question': case 'final_answer':
      return <QuestionMessage game={game} text={game.state.question.text} image={game.state.question.image}
        audio={game.state.question.audio} video={game.state.question.video} isContentPlaying={isContentPlaying} />;
    case 'question_end': {
      let ansImg = game.state.question.answerImage;
      if (!ansImg && !game.state.question.answerText && !game.state.question.answerVideo) ansImg = game.state.question.image;
      return <QuestionMessage game={game} text={game.state.question.answerText} image={ansImg}
        audio={game.state.question.answerAudio} video={game.state.question.answerVideo} isContentPlaying={isContentPlaying} />;
    }
    case 'final_player_answer': return <ViewTextContent>{game.state.answerer?.finalAnswer || '\u2E3B'}</ViewTextContent>;
    case 'final_player_bet': return <ViewTextContent>{game.state.answerer?.finalBet}</ViewTextContent>;
    default: return <ViewTextContent>Jeopardy</ViewTextContent>;
  }
};

const JeopardyView: React.FC = () => {
  const [isContentPlaying, setContentPlaying] = useState(true);
  const contentPlayingRef = useRef(isContentPlaying);
  contentPlayingRef.current = isContentPlaying;

  const onStateChange = useCallback((game: any) => {
    resetSounds();
    if (['question', 'question_end', 'final_question'].includes(game.state.value) && !contentPlayingRef.current) setContentPlaying(true);
    switch (game.state.value) {
      case 'intro': Music.intro.play(); break;
      case 'round': Music.round.play(); break;
      case 'round_themes': Music.themes.play(); break;
      case 'question_event':
        if (game.state.question.type === 'auction') Music.auction.play();
        else if (game.state.question.type === 'bagcat') Music.bagcat.play();
        break;
      case 'final_answer': Music.minute.play(); break;
      case 'game_end': Music.game_end.play(); break;
      default: break;
    }
  }, []);

  const onIntercom = useCallback((message: string) => {
    switch (message) {
      case 'skip': Sounds.skip.play(); break;
      case 'sound_stop': resetSounds(); setContentPlaying(false); break;
      case 'replay': setContentPlaying(true); break;
      default: break;
    }
  }, []);

  const game = useGame(JEOPARDY_API, onStateChange, onIntercom);

  useEffect(() => { loadSounds(); return resetSounds; }, []);

  const [connected, setConnected] = useAuth(JEOPARDY_API);
  const navigate = useNavigate();
  const onLogout = () => { JEOPARDY_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={JEOPARDY_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  return (
    <GameView>
      <ViewExitButton onClick={onLogout} />
      <ViewContent>{stateContent(game, isContentPlaying)}</ViewContent>
    </GameView>
  );
};

export default JeopardyView;
