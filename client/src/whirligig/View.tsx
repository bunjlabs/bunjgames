import React, { useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { GiMusicalNotes } from 'react-icons/gi';

import { AudioPlayer, ImagePlayer, VideoPlayer, HowlWrapper } from 'components/Media';
import { Loading } from 'components/UI';
import { useGame, useAuth } from 'components/hooks';
import { AdminAuth } from 'components/Auth';
import { GameView, ViewContent, ViewExitButton, ViewTextContent } from 'components/ViewLayout';
import Whirligig from './Whirligig';
import { WHIRLIGIG_API } from './api';

const QuestionsEndMusic = {
  current: 0,
  music: [
    HowlWrapper('/sounds/whirligig/question_end_1.mp3'),
    HowlWrapper('/sounds/whirligig/question_end_2.mp3'),
    HowlWrapper('/sounds/whirligig/question_end_3.mp3'),
    HowlWrapper('/sounds/whirligig/question_end_4.mp3'),
    HowlWrapper('/sounds/whirligig/question_end_5.mp3'),
  ],
};

const Music = {
  start: HowlWrapper('/sounds/whirligig/start.mp3'),
  intro: HowlWrapper('/sounds/whirligig/intro.mp3'),
  questions: HowlWrapper('/sounds/whirligig/questions.mp3'),
  whirligig: HowlWrapper('/sounds/whirligig/whirligig.mp3'),
  end: HowlWrapper('/sounds/whirligig/end_defeat.mp3'),
  end_victory: HowlWrapper('/sounds/whirligig/end_victory.mp3'),
  black_box: HowlWrapper('/sounds/whirligig/black_box.mp3'),
};

const Sounds = {
  timerBegin: HowlWrapper('/sounds/whirligig/sig1.mp3'),
  timerWarning: HowlWrapper('/sounds/whirligig/sig2.mp3'),
  timerEnd: HowlWrapper('/sounds/whirligig/sig3.mp3'),
  gong: HowlWrapper('/sounds/whirligig/gong.mp3'),
};

const loadSounds = () => {
  QuestionsEndMusic.music.forEach((m) => m.load());
  Object.values(Music).forEach((m) => m.load());
  Object.values(Sounds).forEach((m) => m.load());
};

const resetSounds = () => {
  QuestionsEndMusic.music.forEach((m) => m.stop());
  Object.values(Music).forEach((m) => m.stop());
};

const QuestionMessageInner: React.FC<{
  game: any; text?: string; image?: string; audio?: string; video?: string;
}> = ({ game, text, image, audio, video }) => {
  const sv = game.state.value;
  const videoKey = useMemo(() => video ? `${sv}-${video}` : null, [sv, video]);
  const audioKey = useMemo(() => audio ? `${sv}-${audio}` : null, [sv, audio]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', width: '100%', height: '100%', fontSize: 38, color: 'var(--text)' }}>
      {text && !image && !video && <p style={{ whiteSpace: 'pre-wrap', textAlign: 'center', padding: 16 }}>{text}</p>}
      {image && <ImagePlayer game={game} url={image} style={{ width: '95vw', height: '95vh' }} />}
      {['question_start', 'right_answer'].includes(sv) && audio && (
        <div key={audioKey!}><AudioPlayer controls={true} playing={true} game={game} url={audio} /></div>
      )}
      {['question_start', 'right_answer'].includes(sv) && video && (
        <VideoPlayer key={videoKey!} controls={true} playing={true} game={game} url={video} style={{ width: '95vw', height: '95vh' }} />
      )}
      {!text && !image && !video && audio && <p style={{ fontSize: 150 }}><GiMusicalNotes /></p>}
    </div>
  );
};

const QuestionMessage = React.memo(QuestionMessageInner);

const WhirligigView: React.FC = () => {
  const onStateChange = useCallback((game: any) => {
    resetSounds();
    switch (game.state.value) {
      case 'start': Music.start.play(); break;
      case 'intro': Music.intro.play(); break;
      case 'questions': Music.questions.play(); break;
      case 'question_whirligig': Music.whirligig.play(); break;
      case 'question_end':
        QuestionsEndMusic.music[QuestionsEndMusic.current].play();
        QuestionsEndMusic.current = (QuestionsEndMusic.current + 1) % QuestionsEndMusic.music.length;
        break;
      case 'end': Music.end.play(); break;
      default: break;
    }
  }, []);

  const onIntercom = useCallback((message: string) => {
    switch (message) {
      case 'gong': Sounds.gong.play(); break;
      case 'sound_stop': resetSounds(); break;
      case 'timer_begin': Sounds.timerBegin.play(); break;
      case 'timer_warning': Sounds.timerWarning.play(); break;
      case 'timer_end': Sounds.timerEnd.play(); break;
      default: break;
    }
  }, []);

  const game = useGame(WHIRLIGIG_API, onStateChange, onIntercom);

  useEffect(() => { loadSounds(); return resetSounds; }, []);

  const [connected, setConnected] = useAuth(WHIRLIGIG_API);
  const navigate = useNavigate();
  const onLogout = () => { WHIRLIGIG_API.logout(); navigate('/admin'); };

  if (!connected) return <AdminAuth api={WHIRLIGIG_API} setConnected={setConnected} />;
  if (!game) return <Loading />;

  const isQuestion = ['question_start', 'question_discussion', 'answer'].includes(game.state.value)
    && game.state.question && ['text', 'image', 'audio', 'video'].some((v) => game.state.question[v]);

  const isAnswer = game.state.value === 'right_answer'
    && game.state.question?.answer && ['text', 'image', 'audio', 'video'].some((v) => game.state.question.answer[v]);

  const isWhirligig = game.state.value === 'question_whirligig';

  let content;
  if (isWhirligig) {
    content = <Whirligig game={game} callback={() => Music.whirligig.stop()} />;
  } else if (isQuestion) {
    const { text, image, audio, video } = game.state.question;
    content = <QuestionMessage game={game} text={text} image={image} audio={audio} video={video} />;
  } else if (isAnswer) {
    const { text, image, audio, video } = game.state.question.answer;
    content = <QuestionMessage game={game} text={text} image={image} audio={audio} video={video} />;
  } else {
    content = (
      <ViewTextContent style={{ fontWeight: 'bold', fontSize: '30vmin' }}>
        {game.score.connoisseurs} : {game.score.viewers}
      </ViewTextContent>
    );
  }

  return (
    <GameView>
      <ViewExitButton onClick={onLogout} />
      <ViewContent>{content}</ViewContent>
    </GameView>
  );
};

export default WhirligigView;
